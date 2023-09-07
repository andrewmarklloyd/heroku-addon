package datadog

type MetricName string

const (
	MetricNameLogin              MetricName = "login"
	MetricNameProvision          MetricName = "instance.provision"
	MetricNameDeprovision        MetricName = "instance.deprovision"
	MetricNameDeleteInstance     MetricName = "instance.delete"
	MetricNameStripeWebhookEvent MetricName = "stripe.webhook_event"
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
