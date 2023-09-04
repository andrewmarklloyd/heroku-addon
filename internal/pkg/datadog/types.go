package datadog

type MetricName string

const (
	MetricNameLogin MetricName = "login"
)

type MetricTag string

const (
	MetricTagGithub MetricTag = "github"
	MetricTagHeroku MetricTag = "heroku"
)

type CustomMetric struct {
	MetricName  MetricName
	MetricValue float64
	Tags        map[string]string
}
