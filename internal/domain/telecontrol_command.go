package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrTelecontrolCommandNotFound      = errors.New("telecontrol command not found")
	ErrTelecontrolCommandAlreadyExists = errors.New("telecontrol command already exists")
)

type TelecontrolCommandStatus string

const (
	TelecontrolCommandStatusAccepted  TelecontrolCommandStatus = "accepted"
	TelecontrolCommandStatusPublished TelecontrolCommandStatus = "published"
	TelecontrolCommandStatusFailed    TelecontrolCommandStatus = "failed"
)

type TelecontrolCommand struct {
	ID            string                   `json:"id"`
	DeviceID      string                   `json:"device_id"`
	Command       string                   `json:"command"`
	Payload       json.RawMessage          `json:"payload"`
	CorrelationID string                   `json:"correlation_id,omitempty"`
	RequestedBy   string                   `json:"requested_by,omitempty"`
	Status        TelecontrolCommandStatus `json:"status"`
	Destination   string                   `json:"destination,omitempty"`
	FailureReason string                   `json:"failure_reason,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
	PublishedAt   *time.Time               `json:"published_at,omitempty"`
	FailedAt      *time.Time               `json:"failed_at,omitempty"`
	ExpiresAt     *time.Time               `json:"expires_at,omitempty"`
}

type TelecontrolCommandRepository interface {
	Create(ctx context.Context, command TelecontrolCommand) error
	Update(ctx context.Context, command TelecontrolCommand) error
	GetByID(ctx context.Context, id string) (TelecontrolCommand, error)
	List(ctx context.Context) ([]TelecontrolCommand, error)
}

func (c TelecontrolCommand) Clone() TelecontrolCommand {
	c.Payload = cloneRawMessage(c.Payload)
	return c
}

func (c *TelecontrolCommand) MarkPublished(now time.Time, destination string) {
	publishedAt := now
	c.Status = TelecontrolCommandStatusPublished
	c.Destination = destination
	c.PublishedAt = &publishedAt
	c.FailedAt = nil
	c.FailureReason = ""
}

func (c *TelecontrolCommand) MarkFailed(now time.Time, reason string) {
	failedAt := now
	c.Status = TelecontrolCommandStatusFailed
	c.FailedAt = &failedAt
	c.FailureReason = reason
}

func cloneRawMessage(payload json.RawMessage) json.RawMessage {
	if payload == nil {
		return nil
	}

	cloned := make([]byte, len(payload))
	copy(cloned, payload)

	return cloned
}
