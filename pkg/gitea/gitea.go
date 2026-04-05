package gitea

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/go-kube/pkg/stack"
	"danicos.dev/daniel/kube-deploy/pkg/services"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

var Secret = struct {
	Name                  string
	ActRunnerActionsToken string
	ActRunnerDanielToken  string
}{
	Name:                  "gitea-secret",
	ActRunnerActionsToken: "act-runner-actions-token",
	ActRunnerDanielToken:  "act-runner-daniel-token",
}

var meta kube.Metadata
var Namespace = kube.Namespace(services.Gitea)
var PVC core.PersistentVolumeClaim
var SRV core.Service
var CFG core.ConfigMap

var daniel_runner_meta kube.Metadata
var DanielRunnerCachePVC core.PersistentVolumeClaim
var DanielRunnerPVC core.PersistentVolumeClaim

var action_runner_meta kube.Metadata
var ActionsRunnerCachePVC core.PersistentVolumeClaim
var ActionsRunnerPVC core.PersistentVolumeClaim

func init() {
	meta = kube.NewMetadata(services.Gitea, Namespace)
	PVC = meta.PVC()
	CFG = kube.ConfigFromFile("config.yaml", "config/act_runner/act-runner-host.yaml", meta)
	SRV = meta.ServiceFrom(kube.ServicePort{
		Name: "http",
		Port: services.GiteaPort,
	}, kube.ServicePort{
		Name: "ssh",
		Port: 22,
	})

	action_runner_meta = kube.NewMetadata("actions-act-runner", Namespace)
	ActionsRunnerPVC = action_runner_meta.PVC()
	other_meta := kube.NewMetadata("actions-cache-act-runner", Namespace)
	ActionsRunnerCachePVC = other_meta.PVC()

	daniel_runner_meta = kube.NewMetadata("daniel-act-runner", Namespace)
	DanielRunnerPVC = daniel_runner_meta.PVC()
	daniel_other_meta := kube.NewMetadata("daniel-cache-act-runner", Namespace)
	DanielRunnerCachePVC = daniel_other_meta.PVC()
}

func Stack() stack.Stack {
	giteaManifests := map[string]any{
		"namespace":         Namespace,
		"srv":               SRV,
		"deployment":        StatefulSet(),
		"ingress":           Ingress(),
		"config-map":        CFG,
		"actions-cache-pvc": ActionsRunnerCachePVC,
		"actions-data-pvc":  ActionsRunnerPVC,
		"actions-worker":    ActionsActRunnerDeployment(),
		"daniel-cache-pvc":  DanielRunnerCachePVC,
		"daniel-data-pvc":   DanielRunnerPVC,
		"daniel-worker":     DanielActRunnerDeployment(),
	}
	return stack.NewStack(services.Gitea, giteaManifests)
}

func StatefulSet() apps.StatefulSet {
	/*
		TODO(daniel): Make sure the container has access to the local timezone .
		/etc/localtime:/etc/localtime:ro
	*/
	podSpec := core.PodSpec{
		Containers: []core.Container{{
			Name:          services.Gitea,
			Image:         services.GiteaImage,
			Env:           []core.EnvVar{{Name: "TZ", Value: "America/Toronto"}},
			LivenessProbe: kube.LivenessProbe("/api/healthz", "http"),
			Ports: []core.ContainerPort{
				{
					Name:          "http",
					ContainerPort: services.GiteaPort,
				},
				{
					Name:          "ssh",
					ContainerPort: 22,
				},
			},
			VolumeMounts: []core.VolumeMount{{
				Name:      PVC.Name,
				MountPath: "/data",
			}},
		}},
	}
	var replicas int32
	replicas = services.GiteaReplicas
	pvcs := []core.PersistentVolumeClaim{PVC}
	s := kube.NewStatefulSet(meta, podSpec, pvcs, &replicas)
	return s
}

func Ingress() net.Ingress {
	rules := []kube.IngressRule{
		{
			Host:        services.GiteaHost,
			ServiceName: SRV.Name,
			PortNumber:  services.GiteaPort,
		},
	}
	return kube.Ingress(Namespace.Name, rules, true)
}

func ActionsActRunnerDeployment() apps.Deployment {
	/*
		NOTE: Perhaps this sould be a statefulSet and not a Deployment.
	*/
	configVolume := kube.NewVolumeFrom(kube.VolumeSourceConfigMap, "config", CFG.Name)
	dataVolume := kube.NewVolumeFrom(kube.VolumeSourcePVC, "data", ActionsRunnerPVC.Name)
	cacheVolume := kube.NewVolumeFrom(kube.VolumeSourcePVC, "cache", ActionsRunnerCachePVC.Name)
	envMapping := map[string]string{
		"CONFIG_FILE":         "/etc/act_runner/config.yaml",
		"GITEA_INSTANCE_URL":  "https://" + services.GiteaHost,
		"GITEA_RUNNER_NAME":   action_runner_meta.Meta().Name,
		"GITEA_RUNNER_LABELS": "actions-act-runner:host",
	}
	secretMapping := map[string]string{
		"GITEA_RUNNER_REGISTRATION_TOKEN": Secret.ActRunnerActionsToken,
	}
	podSpec := core.PodSpec{
		Containers: []core.Container{{
			Name:  action_runner_meta.Meta().Name,
			Image: "docker.io/gitea/act_runner:latest",
			Env:   kube.NewEnvVarWithSecret(envMapping, secretMapping, Secret.Name),
			VolumeMounts: []core.VolumeMount{
				{
					Name:      configVolume.Name,
					MountPath: "/etc/act_runner",
					ReadOnly:  true,
				},
				{
					Name:      dataVolume.Name,
					MountPath: "/data",
				},
				{
					Name:      cacheVolume.Name,
					MountPath: "/root/.cache",
				},
			},
		}},
		Volumes: []core.Volume{
			configVolume,
			dataVolume,
			cacheVolume,
		},
	}
	return kube.NewDeployment(action_runner_meta, podSpec)
}

func DanielActRunnerDeployment() apps.Deployment {
	/*
		NOTE: Perhaps this sould be a statefulSet and not a Deployment.
		TODO: Make sure the container in this deployment has Go installed. apk add go
	*/
	configVolume := kube.NewVolumeFrom(kube.VolumeSourceConfigMap, "config", CFG.Name)
	dataVolume := kube.NewVolumeFrom(kube.VolumeSourcePVC, "data", DanielRunnerPVC.Name)
	cacheVolume := kube.NewVolumeFrom(kube.VolumeSourcePVC, "cache", DanielRunnerCachePVC.Name)
	envMapping := map[string]string{
		"CONFIG_FILE":         "/etc/act_runner/config.yaml",
		"GITEA_INSTANCE_URL":  "https://" + services.GiteaHost,
		"GITEA_RUNNER_NAME":   daniel_runner_meta.Meta().Name,
		"GITEA_RUNNER_LABELS": "daniel-act-runner:host",
	}
	secretMapping := map[string]string{
		"GITEA_RUNNER_REGISTRATION_TOKEN": Secret.ActRunnerDanielToken,
	}
	podSpec := core.PodSpec{
		Containers: []core.Container{{
			Name:  daniel_runner_meta.Meta().Name,
			Image: "docker.io/gitea/act_runner:latest",
			Env:   kube.NewEnvVarWithSecret(envMapping, secretMapping, Secret.Name),
			VolumeMounts: []core.VolumeMount{
				{
					Name:      configVolume.Name,
					MountPath: "/etc/act_runner",
					ReadOnly:  true,
				},
				{
					Name:      dataVolume.Name,
					MountPath: "/data",
				},
				{
					Name:      cacheVolume.Name,
					MountPath: "/root/.cache",
				},
			},
		}},
		Volumes: []core.Volume{
			configVolume,
			dataVolume,
			cacheVolume,
		},
	}
	return kube.NewDeployment(daniel_runner_meta, podSpec)
}

/*
	Backup Gitea folders
	- app.ini -> /data/gitea/conf/app.ini
	- data/* 	-> /data/gitea
	- repos/* -> /data/git/repositories/

		chown -R git:git /data
	# Regenerate Git Hooks
	/usr/local/bin/gitea -c '/data/gitea/conf/app.ini' admin regenerate hooks
*/
