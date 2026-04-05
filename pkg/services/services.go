package services

const (
	Gitea         = "gitea"
	GiteaHost     = "danicos.dev"
	GiteaImage    = "docker.gitea.com/gitea:1.25.4"
	GiteaPort     = 3000
	GiteaReplicas = 1

	CuriousApe      = "curious-ape"
	CuriusApeHost   = "ape.danicos.me"
	CuriousApeImage = "danicos.dev/daniel/curious-ape:v1.1.0"
	CuriuosApePort  = 4000

	Temporal            = "temporal"
	TemporalServerImage = "temporalio/server:latest"

	TemporalUI      = "temporal-ui"
	TemporalUIImage = "temporalio/ui:latest"
	TemporalUIHost  = "temporal.danicos.dev"
	TemporalUIPort  = 8080

	LitestreamImage = "docker.io/litestream/litestream:0.5-scratch"

	ObservabilityNamespace = "observe"
	Alloy                  = "alloy"
	Prometheus             = "prometheus"
	Grafana                = "grafana"
	GrafanaHost            = "grafana.danicos.dev"
	GrafanaPort            = 80
)

// Static Web Pages
const (
	Homepage      = "homepage"
	HomepageImage = "danicos.dev/daniel/homepage"
	HomepageHost  = "danicos.me"

	SmallTech      = "small-tech"
	SmallTechImage = "danicos.dev/daniel/small_tech"
	SmallTechHost  = "small.tech"
)

// Other
const ReplicaOne = 1
