package static

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/go-kube/pkg/stack"
	"danicos.dev/daniel/kube-deploy/pkg/services"
	net "k8s.io/api/networking/v1"
)

var Namespace = kube.Namespace("web-pages")

func Stack() stack.Stack {
	return stack.NewStack("web-pages", map[string]any{
		"namespace":             Namespace,
		"ingress":               Ingress(),
		"homepage-srv":          HomepageSRV,
		"homepage-deployment":   SmallTechDeployment(),
		"small-tech-srv":        SmallTechSRV,
		"small-tech-deployment": SmallTechDeployment(),
	})
}

func Ingress() net.Ingress {
	rules := []kube.IngressRule{
		{
			Host:        services.HomepageHost,
			ServiceName: HomepageSRV.Name,
			PortNumber:  webPort,
		},
		{
			Host:        "www." + services.HomepageHost,
			ServiceName: HomepageSRV.Name,
			PortNumber:  webPort,
		},
		{
			Host:        services.SmallTechHost,
			ServiceName: SmallTechSRV.Name,
			PortNumber:  webPort,
		},
		{
			Host:        "www." + services.SmallTechHost,
			ServiceName: SmallTechSRV.Name,
			PortNumber:  webPort,
		},
	}
	return kube.Ingress(Namespace.Name, rules, true)
}
