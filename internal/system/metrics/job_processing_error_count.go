package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type jobErrorsTotal struct {
	metric *prometheus.CounterVec
}

func newJobErrorsTotal() *jobErrorsTotal {
	return &jobErrorsTotal{
		metric: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "notifs_job_errors_total",
			Help: "Count of job errors",
		}, []string{jobName}),
	}
}

func (m *jobErrorsTotal) Register(registry *prometheus.Registry) {
	registry.MustRegister(m.metric)
}

func (m *jobErrorsTotal) Inc(job string) {
	m.metric.With(prometheus.Labels{jobName: job}).Inc()
}
