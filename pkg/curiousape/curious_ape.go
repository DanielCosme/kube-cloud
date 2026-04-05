package curiousape

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/go-kube/pkg/stack"
	"danicos.dev/daniel/kube-deploy/pkg/services"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

var Secret = struct {
	Name          string
	ConfigKey     string
	LitestreamKey string
}{
	Name:          "curious-ape-secret",
	ConfigKey:     "config.json",
	LitestreamKey: "litestream.yml",
}

var Namespace = kube.Namespace(services.CuriousApe)

var meta kube.Metadata
var SRV core.Service
var PVC core.PersistentVolumeClaim

func init() {
	meta = kube.NewMetadata(services.CuriousApe, Namespace)
	PVC = meta.PVC()
	SRV = meta.Service(services.CuriuosApePort)
}

func Stack() stack.Stack {
	return stack.NewStack(services.CuriousApe, map[string]any{
		"namespace":  Namespace,
		"pvc":        PVC,
		"srv":        SRV,
		"deployment": Deployment(),
		"ingress":    Ingress(),
	})
}

func Deployment() apps.Deployment {
	dataVolume := kube.NewVolumeFrom(kube.VolumeSourcePVC, "data", PVC.Name)
	configVolume := kube.NewVolumeFromSecret("config", Secret.Name, []core.KeyToPath{{
		Key:  Secret.ConfigKey,
		Path: Secret.ConfigKey,
	}})
	litestreamVolume := kube.NewVolumeFromSecret("litestream-vol", Secret.Name, []core.KeyToPath{{
		Key:  Secret.LitestreamKey,
		Path: Secret.LitestreamKey,
	}})
	podSpec := core.PodSpec{
		InitContainers: []core.Container{{
			Name:  "restore-litestream",
			Image: services.LitestreamImage,
			Command: []string{
				"litestream",
				"restore",
				"-if-db-not-exists",
				"-if-replica-exists",
				"/db-data/ape.db",
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      dataVolume.Name,
					MountPath: "/db-data",
				},
				{
					Name:      litestreamVolume.Name,
					MountPath: "/etc/litestream.yml",
					SubPath:   Secret.LitestreamKey,
				},
			},
		}},
		Containers: []core.Container{
			{
				Name:  services.CuriousApe,
				Image: services.CuriousApeImage,
				Ports: []core.ContainerPort{{ContainerPort: int32(services.CuriuosApePort)}},
				Env:   []core.EnvVar{{Name: "APE_ENVIRONMENT", Value: "prod"}},
				VolumeMounts: []core.VolumeMount{
					{
						Name:      configVolume.Name,
						MountPath: "/app/config.json",
						SubPath:   Secret.ConfigKey,
					},
					{
						Name:      dataVolume.Name,
						MountPath: "/app/db-data",
					},
				},
			},
			{
				Name:  "replicate-litestream",
				Image: services.LitestreamImage,
				Command: []string{
					"litestream",
					"replicate",
				},
				VolumeMounts: []core.VolumeMount{
					{
						Name:      dataVolume.Name,
						MountPath: "/db-data",
					},
					{
						Name:      litestreamVolume.Name,
						MountPath: "/etc/litestream.yml",
						SubPath:   Secret.LitestreamKey,
					},
				},
			},
		},
		Volumes: []core.Volume{
			dataVolume,
			configVolume,
			litestreamVolume,
		},
		// TODO: Sidecar Container -> Litestream (replicate)
	}
	d := kube.NewDeployment(meta, podSpec)
	return d
}

func Ingress() net.Ingress {
	rules := []kube.IngressRule{
		{
			Host:        services.CuriusApeHost,
			ServiceName: SRV.Name,
			PortNumber:  services.CuriuosApePort,
		},
	}
	return kube.Ingress(Namespace.Name, rules, true)
}
