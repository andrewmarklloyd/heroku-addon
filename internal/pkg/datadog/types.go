package datadog

type MetricName string

const (
	MetricNameHerokuLogin MetricName = "heroku.login"
	MetricNameGithubLogin MetricName = "github.login"
)
