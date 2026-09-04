package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"ms-telecontrol/internal/domain"
	"ms-telecontrol/internal/usecase/port"
)

const defaultTopicPattern = "telecontrol/{device_id}/commands"

type PublisherOptions struct {
	BrokerURLs     []string
	ClientID       string
	Username       string
	Password       string
	TopicPattern   string
	QoS            byte
	Retained       bool
	ConnectTimeout time.Duration
	PublishTimeout time.Duration
}

type Publisher struct {
	client         paho.Client
	topicPattern   string
	qos            byte
	retained       bool
	publishTimeout time.Duration
}

func NewPublisher(options PublisherOptions) (*Publisher, error) {
	if len(options.BrokerURLs) == 0 {
		return nil, errors.New("mqtt broker url is required")
	}

	if strings.TrimSpace(options.ClientID) == "" {
		return nil, errors.New("mqtt client id is required")
	}

	if options.QoS > 2 {
		return nil, fmt.Errorf("mqtt qos must be between 0 and 2, got %d", options.QoS)
	}

	topicPattern := strings.TrimSpace(options.TopicPattern)
	if topicPattern == "" {
		topicPattern = defaultTopicPattern
	}
	if strings.ContainsAny(topicPattern, "+#") {
		return nil, errors.New("mqtt command topic pattern must not contain wildcards")
	}

	connectTimeout := options.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}

	publishTimeout := options.PublishTimeout
	if publishTimeout <= 0 {
		publishTimeout = 3 * time.Second
	}

	clientOptions := paho.NewClientOptions().
		SetClientID(strings.TrimSpace(options.ClientID)).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second)

	addedBroker := false
	for _, brokerURL := range options.BrokerURLs {
		brokerURL = strings.TrimSpace(brokerURL)
		if brokerURL != "" {
			clientOptions.AddBroker(brokerURL)
			addedBroker = true
		}
	}

	if !addedBroker {
		return nil, errors.New("mqtt broker url is required")
	}

	if options.Username != "" {
		clientOptions.SetUsername(options.Username)
		clientOptions.SetPassword(options.Password)
	}

	client := paho.NewClient(clientOptions)
	token := client.Connect()
	if !token.WaitTimeout(connectTimeout) {
		return nil, fmt.Errorf("connect mqtt broker: timeout after %s", connectTimeout)
	}

	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("connect mqtt broker: %w", err)
	}

	return &Publisher{
		client:         client,
		topicPattern:   topicPattern,
		qos:            options.QoS,
		retained:       options.Retained,
		publishTimeout: publishTimeout,
	}, nil
}

func (p *Publisher) PublishCommand(ctx context.Context, command domain.TelecontrolCommand) (port.CommandPublication, error) {
	if err := ctx.Err(); err != nil {
		return port.CommandPublication{}, err
	}

	if !p.client.IsConnectionOpen() {
		return port.CommandPublication{}, errors.New("mqtt client is not connected")
	}

	topic := p.resolveTopic(command)
	payload, err := json.Marshal(domain.NewTelecontrolCommandIssuedEvent(command))
	if err != nil {
		return port.CommandPublication{}, fmt.Errorf("marshal telecontrol command event: %w", err)
	}

	token := p.client.Publish(topic, p.qos, p.retained, payload)
	if err := waitToken(ctx, token, p.publishTimeout); err != nil {
		return port.CommandPublication{}, fmt.Errorf("publish mqtt message: %w", err)
	}

	return port.CommandPublication{
		Destination: topic,
	}, nil
}

func (p *Publisher) Close() error {
	if p.client != nil && p.client.IsConnected() {
		p.client.Disconnect(250)
	}

	return nil
}

func (p *Publisher) resolveTopic(command domain.TelecontrolCommand) string {
	topic := strings.ReplaceAll(p.topicPattern, "{device_id}", command.DeviceID)
	topic = strings.ReplaceAll(topic, "{command}", command.Command)
	topic = strings.ReplaceAll(topic, "{command_id}", command.ID)

	return topic
}

func waitToken(ctx context.Context, token paho.Token, timeout time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if deadline, ok := ctx.Deadline(); ok {
		deadlineTimeout := time.Until(deadline)
		if deadlineTimeout <= 0 {
			return ctx.Err()
		}

		if timeout <= 0 || deadlineTimeout < timeout {
			timeout = deadlineTimeout
		}
	}

	if timeout > 0 {
		if !token.WaitTimeout(timeout) {
			return context.DeadlineExceeded
		}
	} else {
		token.Wait()
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return token.Error()
}
