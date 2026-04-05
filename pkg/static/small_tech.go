package static

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/kube-deploy/pkg/services"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var small_techPort int32
var small_tech_meta kube.Metadata
var SmallTechSRV core.Service

func init() {
	small_tech_meta = kube.NewMetadata(services.SmallTech, Namespace)
	SmallTechSRV = small_tech_meta.Service(webPort)
}

func SmallTechDeployment() apps.Deployment {
	podSpec := core.PodSpec{
		Containers: []core.Container{{
			Name:  services.SmallTech,
			Image: services.SmallTechImage,
			Ports: []core.ContainerPort{{ContainerPort: webPort}},
		}},
	}
	return kube.NewDeployment(small_tech_meta, podSpec)
}
