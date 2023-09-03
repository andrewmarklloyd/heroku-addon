package datadog

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type Client struct {
	api    *datadogV2.MetricsApi
	apiKey string
	env    string
}

func NewDatadogClient(apiKey string, testMode bool) Client {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewMetricsApi(apiClient)

	env := "prod"
	if testMode {
		env = "test"
	}

	return Client{
		api:    api,
		apiKey: apiKey,
		env:    env,
	}
}

func (c *Client) Publish(ctx context.Context, metricName MetricName, metricValue float64) error {
	valueCtx := context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: c.apiKey,
			},
		},
	)

	body := datadogV2.MetricPayload{
		Series: []datadogV2.MetricSeries{
			{
				Metric: string(metricName),
				Type:   datadogV2.METRICINTAKETYPE_COUNT.Ptr(),
				Points: []datadogV2.MetricPoint{
					{
						Timestamp: datadog.PtrInt64(time.Now().Unix()),
						Value:     datadog.PtrFloat64(metricValue),
					},
				},
				Resources: []datadogV2.MetricResource{
					{
						Type: datadog.PtrString("source"),
						Name: datadog.PtrString("heroku-addon"),
					},
					{
						Type: datadog.PtrString("env"),
						Name: datadog.PtrString(c.env),
					},
				},
			},
		},
	}

	_, _, err := c.api.SubmitMetrics(valueCtx, body, *datadogV2.NewSubmitMetricsOptionalParameters())
	if err != nil {
		return fmt.Errorf("submitting metrics: %s", err)
	}

	return nil
}
