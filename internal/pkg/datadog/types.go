package datadog

type MetricName string

const (
	MetricNameHerokuLogin  MetricName = "heroku.login"
	MetricNameHerokuGithub MetricName = "heroku.github"
)
