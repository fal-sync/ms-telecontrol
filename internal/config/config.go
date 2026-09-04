package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName         string
	HTTPPort        string
	ShutdownTimeout time.Duration
	External        ExternalConfig
}

type ExternalConfig struct {
	GatewayTelemetry GatewayTelemetryConfig
	MQTT             MQTTConfig
}

type GatewayTelemetryConfig struct {
	BaseURL string
	Path    string
	Timeout time.Duration
}

type MQTTConfig struct {
	Enabled             bool
	BrokerURLs          []string
	ClientID            string
	Username            string
	Password            string
	CommandTopicPattern string
	QoS                 byte
	Retained            bool
	ConnectTimeout      time.Duration
	PublishTimeout      time.Duration
}

func Load() Config {
	appName := getEnv("APP_NAME", "ms-telecontrol")

	return Config{
		AppName:         appName,
		HTTPPort:        getEnv("APP_PORT", "8080"),
		ShutdownTimeout: getDurationEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		External: ExternalConfig{
			GatewayTelemetry: GatewayTelemetryConfig{
				BaseURL: getEnv("GATEWAY_TELEMETRY_BASE_URL", ""),
				Path:    getEnv("GATEWAY_TELEMETRY_COMMAND_PATH", "/internal/telecontrol/commands"),
				Timeout: getDurationEnv("GATEWAY_TELEMETRY_TIMEOUT", 3*time.Second),
			},
			MQTT: MQTTConfig{
				Enabled:             getBoolEnv("MQTT_ENABLED", false),
				BrokerURLs:          getListEnv("MQTT_BROKERS", "tcp://localhost:1883"),
				ClientID:            getEnv("MQTT_CLIENT_ID", appName),
				Username:            getEnv("MQTT_USERNAME", ""),
				Password:            getEnv("MQTT_PASSWORD", ""),
				CommandTopicPattern: getEnv("MQTT_COMMAND_TOPIC_PATTERN", "telecontrol/{device_id}/commands"),
				QoS:                 getByteEnv("MQTT_QOS", 1),
				Retained:            getBoolEnv("MQTT_RETAINED", false),
				ConnectTimeout:      getDurationEnv("MQTT_CONNECT_TIMEOUT", 5*time.Second),
				PublishTimeout:      getDurationEnv("MQTT_PUBLISH_TIMEOUT", 3*time.Second),
			},
		},
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}

func getBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getByteEnv(key string, fallback byte) byte {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return fallback
	}

	return byte(parsed)
}

func getListEnv(key string, fallback ...string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	rawItems := strings.Split(value, ",")
	items := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}

	return items
}
