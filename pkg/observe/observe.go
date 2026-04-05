package observe

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/go-kube/pkg/stack"
	"danicos.dev/daniel/kube-deploy/pkg/services"

	net "k8s.io/api/networking/v1"
)

var Namespace = kube.Namespace(services.ObservabilityNamespace)

func Stack() stack.Stack {
	return stack.NewStack(services.ObservabilityNamespace, map[string]any{
		"namespace":       Namespace,
		"alloy-configmap": AlloyConfigMap,
		"grafana-ingress": Ingress(),
	})
}

func Ingress() net.Ingress {
	rules := []kube.IngressRule{
		{
			Host:        services.GrafanaHost,
			ServiceName: services.Grafana,
			PortNumber:  services.GrafanaPort,
		},
	}
	return kube.Ingress(Namespace.Name, rules, true)
}
