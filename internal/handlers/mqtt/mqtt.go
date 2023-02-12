package mqtt

import (
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/shutdown-listener/internal"
	"github.com/soerenschneider/shutdown-listener/internal/config"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	waitTimeout = 15 * time.Second
	Name        = "mqtt"
)

type Mqtt struct {
	client               mqtt.Client
	conf                 *config.MqttConfig
	incomingMessageQueue chan string
	mutex                sync.Mutex
	started              bool
}

func NewMqttHandler(conf *config.MqttConfig) (*Mqtt, error) {
	if conf == nil {
		return nil, errors.New("empty conf provided")
	}

	return &Mqtt{
		client: nil,
		conf:   conf,
		mutex:  sync.Mutex{},
	}, nil
}

func (m *Mqtt) msgHandler(client mqtt.Client, message mqtt.Message) {
	if m.incomingMessageQueue == nil {
		return
	}

	m.incomingMessageQueue <- string(message.Payload())
}

func getClientId() string {
	prefix := "shutdown-listener"
	hostname, err := os.Hostname()
	if err != nil {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		return fmt.Sprintf("%s-%d", prefix, r1.Int63())
	}

	return fmt.Sprintf("%s-%s", prefix, hostname)
}

func (m *Mqtt) OnConnectHandler(client mqtt.Client) {
	log.Info().Msg("Successfully connected to broker")

	if token := client.Subscribe(m.conf.Topic, 1, m.msgHandler); token.WaitTimeout(waitTimeout) && token.Error() != nil {
		log.Error().Msgf("Can not subscribe to topic %s: %v", m.conf.Topic, token.Error())
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Info().Msgf("Connection lost: %v", err)
	internal.MetricMqttReconnections.Inc()
}

func (m *Mqtt) Start(queue chan string) error {
	defer m.mutex.Unlock()
	m.mutex.Lock()

	if nil == queue {
		return errors.New("received uninitialized channel to write to")
	}

	if m.started {
		return fmt.Errorf("handler %s already started", Name)
	}

	m.started = true
	m.incomingMessageQueue = queue

	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.conf.Host)
	opts.OnConnect = m.OnConnectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true
	opts.ClientID = getClientId()

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.WaitTimeout(waitTimeout) && token.Error() != nil {
		log.Error().Msgf("Can not connect to broker %s: %v", m.conf.Host, token.Error())
		return token.Error()
	}

	return nil
}

func (m *Mqtt) Shutdown() {
	log.Info().Msg("Shutting down mqtt handler...")
	m.client.Unsubscribe(m.conf.Topic)
	m.client.Disconnect(1000)
	log.Info().Msg("Shut down mqtt handler")
}

func (m *Mqtt) Name() string {
	return Name
}
