package temporal

import (
	"fmt"

	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/go-kube/pkg/stack"
	"danicos.dev/daniel/kube-deploy/pkg/services"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

var Secret = struct {
	Name         string
	ClientID     string
	ClientSecret string
}{
	Name:         "temporal-secret",
	ClientID:     "client_id",
	ClientSecret: "client_secret",
}

var (
	meta_server            kube.Metadata
	Namespace              = kube.Namespace(services.Temporal)
	pvc                    core.PersistentVolumeClaim
	server_srv             core.Service
	server_config          core.ConfigMap
	server_grpc_port       kube.ServicePort
	server_http_port       kube.ServicePort
	meta_ui                kube.Metadata
	ui_srv                 core.Service
	ui_port                kube.ServicePort
	ui_config              core.ConfigMap
	defaultNamespaceScript core.ConfigMap
	meta_job               kube.Metadata
)

func init() {
	meta_server = kube.NewMetadata(services.Temporal, Namespace)
	pvc = meta_server.PVC()
	server_config = kube.ConfigFromFile("config.yaml", "./config/temporal/server-config.yaml", meta_server)
	server_grpc_port = kube.ServicePort{
		Name: "grpc",
		Port: 7233,
	}
	server_http_port = kube.ServicePort{
		Name: "http",
		Port: 7243,
	}
	server_srv = meta_server.ServiceFrom(server_grpc_port, server_http_port)
	meta_ui = kube.NewMetadata(services.TemporalUI, Namespace)
	ui_port = kube.ServicePort{Name: "web", Port: services.TemporalUIPort}
	ui_srv = meta_ui.ServiceFrom(ui_port)
	ui_config = kube.ConfigFromFile("config.yaml", "./config/temporal/ui-config.yaml", meta_ui)
	meta_job = kube.NewMetadata("namespace-job", Namespace)
	defaultNamespaceScript = kube.ConfigFromFile("create-namespace.sh", "./config/temporal/create_namespace.sh", meta_job)
}

func Stack() stack.Stack {
	// We use SQLite becaues we use Temporal in a low-throughput single-instance setup.
	return stack.NewStack("temporal", map[string]any{
		"namespace":            Namespace,
		"pvc":                  pvc,
		"server-configmap":     server_config,
		"server-service":       server_srv,
		"server-deployment":    server_deployment(),
		"ui-service":           ui_srv,
		"ui-deployment":        ui_deployment(),
		"ingress":              ingress(),
		"namespace-job-script": defaultNamespaceScript,
		"namespace-job":        namespace_job(),
	})
}

func server_deployment() apps.Deployment {
	envVars := map[string]string{
		"TEMPORAL_SERVER_CONFIG_FILE_PATH": "/etc/temporal/config/config.yaml",
	}
	dataVolume := kube.NewVolumeFrom(kube.VolumeSourcePVC, "data", pvc.Name)
	configVolume := kube.NewVolumeFrom(kube.VolumeSourceConfigMap, "config", server_config.Name)
	podSpec := core.PodSpec{
		Containers: []core.Container{
			{
				Name:  services.Temporal,
				Image: services.TemporalServerImage,
				Env:   kube.NewEnvVar(envVars),
				Ports: []core.ContainerPort{
					{
						Name:          server_grpc_port.Name,
						ContainerPort: server_grpc_port.Port,
					},
					{
						Name:          server_http_port.Name,
						ContainerPort: server_http_port.Port,
					},
				},
				VolumeMounts: []core.VolumeMount{
					{
						Name:      dataVolume.Name,
						MountPath: "/data",
					},
					{
						Name:      configVolume.Name,
						MountPath: "/etc/temporal/config/config.yaml",
						SubPath:   "config.yaml",
					},
				},
			},
		},
		Volumes: []core.Volume{
			dataVolume,
			configVolume,
		},
	}
	return kube.NewDeployment(meta_server, podSpec)
}

func namespace_job() batch.Job {
	scriptVol := kube.NewVolumeFrom(kube.VolumeSourceConfigMap, "script-volume", defaultNamespaceScript.Name)
	restartPolicy := core.RestartPolicyOnFailure
	pod_spec := core.PodSpec{
		RestartPolicy: restartPolicy,
		Containers: []core.Container{{
			Name:    "create-default-namespace",
			Image:   "temporalio/admin-tools:latest",
			Command: []string{"/bin/sh"},
			Args:    []string{"/scripts/create-namespace.sh"},
			Env: kube.NewEnvVar(map[string]string{
				"TEMPORAL_ADDRESS":  "temporal:7233",
				"DEFAULT_NAMESPACE": "default",
			}),
			VolumeMounts: []core.VolumeMount{{
				Name:      scriptVol.Name,
				MountPath: "/scripts",
			}},
		}},
		Volumes: []core.Volume{scriptVol},
	}
	return kube.NewJob(meta_job, pod_spec)
}

func ui_deployment() apps.Deployment {
	env_map := map[string]string{
		"TEMPORAL_ADDRESS":           fmt.Sprintf("%s:%d", meta_server.Meta().Name, server_grpc_port.Port),
		"TEMPORAL_CORS_ORIGINS":      "https://" + services.TemporalUIHost,
		"TEMPORAL_AUTH_ENABLED":      "true",
		"TEMPORAL_AUTH_TYPE":         "oidc",
		"TEMPORAL_AUTH_PROVIDER_URL": "https://" + services.GiteaHost,
		"TEMPORAL_AUTH_ISSUER_URL":   "https://" + services.GiteaHost,
		"TEMPORAL_AUTH_CALLBACK_URL": fmt.Sprintf("https://%s/auth/sso/callback", services.TemporalUIHost),
		"TEMPORAL_AUTH_SCOPES":       "openid,profile,email",
		"TEMPORAL_AUTH_LABEL":        "Gitea SSO",
		"TEMPORAL_UI_PORT":           fmt.Sprintf("%d", services.TemporalUIPort),
		// Open ID Config Discovery:
		// https://danicos.dev/.well-known/openid-configuration
	}
	secrets_map := map[string]string{
		"TEMPORAL_AUTH_CLIENT_ID":     Secret.ClientID,
		"TEMPORAL_AUTH_CLIENT_SECRET": Secret.ClientSecret,
	}
	pod_spec := core.PodSpec{
		Containers: []core.Container{
			{
				Name:  services.TemporalUI,
				Image: services.TemporalUIImage,
				Ports: []core.ContainerPort{{Name: ui_port.Name, ContainerPort: ui_port.Port}},
				Env:   kube.NewEnvVarWithSecret(env_map, secrets_map, Secret.Name),
			},
		},
	}
	return kube.NewDeployment(meta_ui, pod_spec)
}

func ingress() net.Ingress {
	rules := []kube.IngressRule{
		{
			Host:        services.TemporalUIHost,
			ServiceName: ui_srv.Name,
			PortNumber:  services.TemporalUIPort,
		},
	}
	return kube.Ingress(Namespace.Name, rules, true)
}

/*
 GITEA open_ID Config
 {
   "issuer": "https://danicos.dev",
   "authorization_endpoint": "https://danicos.dev/login/oauth/authorize",
   "token_endpoint": "https://danicos.dev/login/oauth/access_token",
   "jwks_uri": "https://danicos.dev/login/oauth/keys",
   "userinfo_endpoint": "https://danicos.dev/login/oauth/userinfo",
   "introspection_endpoint": "https://danicos.dev/login/oauth/introspect",
   "response_types_supported": [
     "code",
     "id_token"
   ],
   "id_token_signing_alg_values_supported": [
     "RS256"
   ],
   "subject_types_supported": [
     "public"
   ],
   "scopes_supported": [
     "openid",
     "profile",
     "email",
     "groups"
   ],
   "claims_supported": [
     "aud",
     "exp",
     "iat",
     "iss",
     "sub",
     "name",
     "preferred_username",
     "profile",
     "picture",
     "website",
     "locale",
     "updated_at",
     "email",
     "email_verified",
     "groups"
   ],
   "code_challenge_methods_supported": [
     "plain",
     "S256"
   ],
   "grant_types_supported": [
     "authorization_code",
     "refresh_token"
   ]
 }
*/
