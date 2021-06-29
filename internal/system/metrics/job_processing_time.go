package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const jobName = "job_name"

type jobProcessingTime struct {
	mu     *sync.Mutex
	metric *prometheus.GaugeVec
	values map[string]TimeRecord
}

func newJobProcessingTime() *jobProcessingTime {
	return &jobProcessingTime{
		mu: &sync.Mutex{},
		metric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "notifs_job_processing_time",
			Help: "Job processing time",
		}, []string{jobName, metricName}),
		values: map[string]TimeRecord{},
	}
}

func (m *jobProcessingTime) Register(registry *prometheus.Registry) {
	registry.MustRegister(m.metric)
}

func (m *jobProcessingTime) SetToPrometheus() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metric.Reset()
	for key, v := range m.values {
		m.metric.With(prometheus.Labels{jobName: key, metricName: "max"}).Set(v.Max)
		m.metric.With(prometheus.Labels{jobName: key, metricName: "sum"}).Set(v.Sum)
		m.metric.With(prometheus.Labels{jobName: key, metricName: "amount"}).Set(float64(v.Amount))
	}
	m.values = map[string]TimeRecord{}
}

func (m *jobProcessingTime) Add(jobName string, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var record TimeRecord
	if existRecord, ok := m.values[jobName]; ok {
		record = existRecord
	}

	record.Add(duration)
	m.values[jobName] = record
}
