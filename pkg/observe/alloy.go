package observe

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/kube-deploy/pkg/services"
	core "k8s.io/api/core/v1"
)

var alloy_meta kube.Metadata
var AlloyConfigMap core.ConfigMap

func init() {
	alloy_meta = kube.NewMetadata(services.Alloy, Namespace)
	AlloyConfigMap = kube.ConfigFromFile("config.alloy", "./config/alloy/config.alloy", alloy_meta)
}
