package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ms-telecontrol/internal/domain"
	"ms-telecontrol/internal/infrastructure/external/httpclient"
)

type GatewayTelemetryHook struct {
	client *httpclient.Client
	path   string
}

type gatewayTelemetryCommandRequest struct {
	ID            string          `json:"id"`
	DeviceID      string          `json:"device_id"`
	Command       string          `json:"command"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	RequestedBy   string          `json:"requested_by,omitempty"`
	Status        string          `json:"status"`
	Destination   string          `json:"destination,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
}

func NewGatewayTelemetryHook(baseURL string, timeout time.Duration, path string) *GatewayTelemetryHook {
	return &GatewayTelemetryHook{
		client: httpclient.New(baseURL, timeout),
		path:   path,
	}
}

func (h *GatewayTelemetryHook) HandleCommandIssued(ctx context.Context, command domain.TelecontrolCommand) error {
	if err := h.client.PostJSON(ctx, h.path, gatewayTelemetryCommandRequest{
		ID:            command.ID,
		DeviceID:      command.DeviceID,
		Command:       command.Command,
		Payload:       command.Payload,
		CorrelationID: command.CorrelationID,
		RequestedBy:   command.RequestedBy,
		Status:        string(command.Status),
		Destination:   command.Destination,
		CreatedAt:     command.CreatedAt,
		PublishedAt:   command.PublishedAt,
		ExpiresAt:     command.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("send telecontrol command to gateway telemetry: %w", err)
	}

	return nil
}
