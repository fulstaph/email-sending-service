package metrics

import (
	"math"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	messageType = "message_type"
	metricName  = "metric_name"
)

type rmqMessageProcessingTime struct {
	mu     *sync.Mutex
	metric *prometheus.GaugeVec
	values map[string]TimeRecord
}

func newRMQMessageProcessingTime() *rmqMessageProcessingTime {
	return &rmqMessageProcessingTime{
		mu: &sync.Mutex{},
		metric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "notifs_message_processing_time",
			Help: "message processing time",
		}, []string{messageType, metricName}),
		values: map[string]TimeRecord{},
	}
}

func (m *rmqMessageProcessingTime) Register(registry *prometheus.Registry) {
	registry.MustRegister(m.metric)
}

func (m *rmqMessageProcessingTime) SetToPrometheus() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metric.Reset()
	for key, v := range m.values {
		m.metric.With(prometheus.Labels{messageType: key, metricName: "max"}).Set(v.Max)
		m.metric.With(prometheus.Labels{messageType: key, metricName: "sum"}).Set(v.Sum)
		m.metric.With(prometheus.Labels{messageType: key, metricName: "amount"}).Set(float64(v.Amount))
	}
	m.values = map[string]TimeRecord{}
}

// Add duration in seconds of message processing time
func (m *rmqMessageProcessingTime) Add(messageType string, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var record TimeRecord
	if existRecord, ok := m.values[messageType]; ok {
		record = existRecord
	}

	// update record with new values
	record.Add(duration)
	m.values[messageType] = record
}

type TimeRecord struct {
	Max    float64
	Sum    float64
	Amount int
}

func (d *TimeRecord) Add(duration float64) {
	d.Max = math.Max(d.Max, duration)
	d.Sum += duration
	d.Amount++
}
