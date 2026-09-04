package domain

import (
	"encoding/json"
	"time"
)

type TelecontrolCommandIssuedEvent struct {
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

func NewTelecontrolCommandIssuedEvent(command TelecontrolCommand) TelecontrolCommandIssuedEvent {
	return TelecontrolCommandIssuedEvent{
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
	}
}
