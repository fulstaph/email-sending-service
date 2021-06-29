package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelType  = "type"
	labelTopic = "topic"
	labelEvent = "event"
)

const (
	typeClaimed = "claimed"
	typeSkipped = "skipped"
	typeSuccess = "success"
	typeFailed  = "failed"
)

type rmqMessageCount struct {
	mu     *sync.Mutex
	metric *prometheus.GaugeVec
	values map[messageCountData]int
}

func newRMQMessageCount() *rmqMessageCount {
	return &rmqMessageCount{
		mu: &sync.Mutex{},
		metric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "notifs_message_count",
			Help: "Count of messages",
		}, []string{labelType, labelTopic, labelEvent}),
		values: map[messageCountData]int{},
	}
}

func (m *rmqMessageCount) Register(registry *prometheus.Registry) {
	registry.MustRegister(m.metric)
}

func (m *rmqMessageCount) SetToPrometheus() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metric.Reset()
	for data, value := range m.values {
		m.metric.With(prometheus.Labels{
			labelType:  data.metricType,
			labelTopic: data.topic,
			labelEvent: data.event,
		}).Set(float64(value))
	}
	m.values = map[messageCountData]int{}
}

func (m *rmqMessageCount) add(data messageCountData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[data]++
}

func (m *rmqMessageCount) AddClaimed(topic string, event string) {
	m.add(messageCountData{
		metricType: typeClaimed,
		topic:      topic,
		event:      event,
	})
}

func (m *rmqMessageCount) AddSkipped(topic string, event string) {
	m.add(messageCountData{
		metricType: typeSkipped,
		topic:      topic,
		event:      event,
	})
}

func (m *rmqMessageCount) AddSuccess(topic string, event string) {
	m.add(messageCountData{
		metricType: typeSuccess,
		topic:      topic,
		event:      event,
	})
}

func (m *rmqMessageCount) AddFailed(topic string, event string) {
	m.add(messageCountData{
		metricType: typeFailed,
		topic:      topic,
		event:      event,
	})
}

type messageCountData struct {
	metricType string
	topic      string
	event      string
}
