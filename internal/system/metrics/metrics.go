package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type promMetric interface {
	// SetToPrometheus is expected to bulk-update specific metrics.
	SetToPrometheus()
}

type Client struct {
	registry *prometheus.Registry

	RMQMessagesProcessingTime *rmqMessageProcessingTime
	RMQMessageCount           *rmqMessageCount
	JobProcessingTime         *jobProcessingTime
	JobErrorsTotal            *jobErrorsTotal
}

func New() *Client {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	client := &Client{
		registry: registry,

		RMQMessagesProcessingTime: newRMQMessageProcessingTime(),
		RMQMessageCount:           newRMQMessageCount(),
		JobProcessingTime:         newJobProcessingTime(),
		JobErrorsTotal:            newJobErrorsTotal(),
	}

	client.RMQMessagesProcessingTime.Register(registry)
	client.RMQMessageCount.Register(registry)
	client.JobProcessingTime.Register(registry)
	client.JobErrorsTotal.Register(registry)

	return client
}

func (c *Client) Metrics() []promMetric {
	return []promMetric{
		c.RMQMessageCount,
		c.RMQMessagesProcessingTime,
		c.JobProcessingTime,
	}
}

func (c *Client) Handler() http.Handler {
	return http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		for _, m := range c.Metrics() {
			m.SetToPrometheus()
		}
		promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{}).ServeHTTP(rsp, req)
	})
}
