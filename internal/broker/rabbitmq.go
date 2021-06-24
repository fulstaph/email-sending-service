package broker

import (
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"projects/email-sending-service/config"
)

type MessageQueue interface {
	Open() error
	Close() error
	Publish(body []byte) error
	Subscribe(handler func([]byte) error) error
}

type mq struct {
	ch     *amqp.Channel
	conn   *amqp.Connection
	cfg    config.Broker
	closed int32
}

var delay = time.Duration(5)

// Open -
//nolint:nakedret
func (m *mq) Open() (err error) {
	m.conn, err = amqp.Dial(m.cfg.Host + m.cfg.Port)
	if err != nil {
		return err //nolint:wrapcheck
	}

	go func() {
		for {
			reason, ok := <-m.conn.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok {
				log.Debugln("connection closed")
				break
			}
			log.Debugf("connection closed, reason: %v", reason)

			// reconnect if not closed by developer
			for {
				// wait 1s for reconnect
				time.Sleep(time.Second * delay)

				conn, err := amqp.Dial(m.cfg.Host + m.cfg.Port)
				if err == nil {
					m.conn = conn
					log.Debugln("reconnect success")
					break
				}

				log.Debugf("reconnect failed, err: %v", err)
			}
		}
	}()

	m.ch, err = m.conn.Channel()
	if err != nil {
		return err
	}
	return
}

// IsClosed indicate closed by developer
func (m *mq) IsClosed() bool {
	return atomic.LoadInt32(&m.closed) == 1
}

func (m *mq) Close() error {
	if m.IsClosed() {
		return amqp.ErrClosed
	}

	atomic.StoreInt32(&m.closed, 1)

	return m.ch.Close() //nolint:wrapcheck
}

func (m *mq) Publish(body []byte) error {
	//_, err := m.ch.QueueDeclare(
	//	m.cfg.QueueName, // name
	//	true,            // durable
	//	false,           // delete when unused
	//	false,           // exclusive
	//	false,           // no-wait
	//	nil,             // arguments
	//)
	//if err != nil {
	//	return err
	//}

	if m.IsClosed() {
		return amqp.ErrClosed
	}

	err := m.ch.Publish(
		m.cfg.Exchange,   // exchange
		m.cfg.RoutingKey, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

func (m *mq) Subscribe(handler func([]byte) error) error {
	if m.IsClosed() {
		return amqp.ErrClosed
	}

	q, err := m.ch.QueueDeclare(
		m.cfg.QueueName, // name
		true,            // durable
		false,           // delete when usused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// bind the queue to the routing key
	err = m.ch.QueueBind(
		m.cfg.QueueName,
		m.cfg.RoutingKey,
		m.cfg.Exchange,
		false,
		nil,
	)
	if err != nil {
		return err //nolint:wrapcheck
	}

	msgs, err := m.ch.Consume( //nolint:ineffassign
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			if handler(d.Body) == nil {
				_ = d.Ack(false)
			} else {
				_ = d.Nack(false, true)
			}
		}
	}()
	return nil
}

func NewMessageQueue(cfg config.Broker) MessageQueue {
	return &mq{
		cfg: cfg,
	}
}
