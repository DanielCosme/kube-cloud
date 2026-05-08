package proxy

import (
	"danicos.dev/daniel/go-kube/pkg/kube"
	"danicos.dev/daniel/go-kube/pkg/stack"

	// "danicos.dev/daniel/kube-deploy/pkg/services"

	// apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
)

const TailscaleHost = "hydra-0.orca-uaru.ts.net"

var Namespace = kube.Namespace("proxy")
var linkding_srv core.Service
var immich_srv core.Service
var glance_srv core.Service
var vaultwarden_srv core.Service

func init() {
	glance_srv = core.Service{
		TypeMeta:   kube.ServiceMeta,
		ObjectMeta: kube.ObjectMeta("proxy-glance", Namespace.Name),
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Port: 30009,
				},
			},
			Type:         core.ServiceTypeExternalName,
			ExternalName: TailscaleHost,
		},
	}
	linkding_srv = core.Service{
		TypeMeta:   kube.ServiceMeta,
		ObjectMeta: kube.ObjectMeta("proxy", Namespace.Name),
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Port: 30010,
				},
			},
			Type:         core.ServiceTypeExternalName,
			ExternalName: TailscaleHost,
		},
	}
	immich_srv = core.Service{
		TypeMeta:   kube.ServiceMeta,
		ObjectMeta: kube.ObjectMeta("proxy-immich", Namespace.Name),
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Port: 30011,
				},
			},
			Type:         core.ServiceTypeExternalName,
			ExternalName: TailscaleHost,
		},
	}
	vaultwarden_srv = core.Service{
		TypeMeta:   kube.ServiceMeta,
		ObjectMeta: kube.ObjectMeta("proxy-vaultwarden", Namespace.Name),
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Port: 30012,
				},
			},
			Type:         core.ServiceTypeExternalName,
			ExternalName: TailscaleHost,
		},
	}
}

func Ingress() net.Ingress {
	rules := []kube.IngressRule{
		{
			Host:        "link.danicos.me",
			ServiceName: linkding_srv.Name,
			PortNumber:  linkding_srv.Spec.Ports[0].Port,
		},
		{
			Host:        "photos.danicos.me",
			ServiceName: immich_srv.Name,
			PortNumber:  immich_srv.Spec.Ports[0].Port,
		},
		{
			Host:        "home.danicos.me",
			ServiceName: glance_srv.Name,
			PortNumber:  glance_srv.Spec.Ports[0].Port,
		},
		{
			Host:        "vault.danicos.me",
			ServiceName: vaultwarden_srv.Name,
			PortNumber:  vaultwarden_srv.Spec.Ports[0].Port,
		},
	}
	return kube.Ingress(Namespace.Name, rules, true)
}

func Stack() stack.Stack {
	return stack.NewStack("proxy", map[string]any{
		"namespace":       Namespace,
		"ingress":         Ingress(),
		"linkding-srv":    linkding_srv,
		"immich-srv":      immich_srv,
		"glance-srv":      glance_srv,
		"vaultwarden-srv": vaultwarden_srv,
	})
}
