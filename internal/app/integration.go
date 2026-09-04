package app

import (
	"io"

	"ms-telecontrol/internal/config"
	mqttadapter "ms-telecontrol/internal/infrastructure/external/messaging/mqtt"
	"ms-telecontrol/internal/infrastructure/external/service"
	"ms-telecontrol/internal/usecase/port"
)

func buildTelecontrolIntegrations(cfg config.Config) (port.CommandPublisher, []port.CommandIssuedHook, []io.Closer, error) {
	var commandPublisher port.CommandPublisher
	hooks := make([]port.CommandIssuedHook, 0, 1)
	closers := make([]io.Closer, 0, 1)

	if cfg.External.MQTT.Enabled {
		publisher, err := mqttadapter.NewPublisher(mqttadapter.PublisherOptions{
			BrokerURLs:     cfg.External.MQTT.BrokerURLs,
			ClientID:       cfg.External.MQTT.ClientID,
			Username:       cfg.External.MQTT.Username,
			Password:       cfg.External.MQTT.Password,
			TopicPattern:   cfg.External.MQTT.CommandTopicPattern,
			QoS:            cfg.External.MQTT.QoS,
			Retained:       cfg.External.MQTT.Retained,
			ConnectTimeout: cfg.External.MQTT.ConnectTimeout,
			PublishTimeout: cfg.External.MQTT.PublishTimeout,
		})
		if err != nil {
			return nil, nil, nil, err
		}

		commandPublisher = publisher
		closers = append(closers, publisher)
	}

	if cfg.External.GatewayTelemetry.BaseURL != "" {
		hooks = append(hooks, service.NewGatewayTelemetryHook(
			cfg.External.GatewayTelemetry.BaseURL,
			cfg.External.GatewayTelemetry.Timeout,
			cfg.External.GatewayTelemetry.Path,
		))
	}

	return commandPublisher, hooks, closers, nil
}
