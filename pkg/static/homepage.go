package static

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/kube-deploy/pkg/services"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var webPort int32
var homepage_meta kube.Metadata
var HomepageSRV core.Service

func init() {
	webPort = 3001
	homepage_meta = kube.NewMetadata(services.Homepage, Namespace)
	HomepageSRV = homepage_meta.Service(webPort)
}

func HomepageDeployment() apps.Deployment {
	podSpec := core.PodSpec{
		Containers: []core.Container{{
			Name:  services.Homepage,
			Image: services.HomepageImage,
			Ports: []core.ContainerPort{{ContainerPort: webPort}},
		}},
	}
	return kube.NewDeployment(homepage_meta, podSpec)
}
