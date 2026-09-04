package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ms-telecontrol/internal/domain"
	"ms-telecontrol/internal/usecase/port"
)

var (
	ErrInvalidDeviceID             = errors.New("device_id is required")
	ErrInvalidCommandName          = errors.New("command is required")
	ErrInvalidCommandPayload       = errors.New("payload must be valid JSON")
	ErrInvalidCommandTTL           = errors.New("ttl_seconds must not be negative")
	ErrCommandPublisherUnavailable = errors.New("telecontrol command publisher is not configured")
	ErrCommandPublishFailed        = errors.New("publish telecontrol command failed")
)

type IssueCommandInput struct {
	DeviceID      string          `json:"device_id"`
	Command       string          `json:"command"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlation_id"`
	RequestedBy   string          `json:"requested_by"`
	TTL           time.Duration   `json:"-"`
}

type TelecontrolService struct {
	commandRepository domain.TelecontrolCommandRepository
	commandPublisher  port.CommandPublisher
	commandHooks      []port.CommandIssuedHook
	now               func() time.Time
	generateID        func() string
}

func NewTelecontrolService(
	commandRepository domain.TelecontrolCommandRepository,
	commandPublisher port.CommandPublisher,
	commandHooks ...port.CommandIssuedHook,
) *TelecontrolService {
	return &TelecontrolService{
		commandRepository: commandRepository,
		commandPublisher:  commandPublisher,
		commandHooks:      commandHooks,
		now:               time.Now,
		generateID:        generateCommandID,
	}
}

func (s *TelecontrolService) IssueCommand(ctx context.Context, input IssueCommandInput) (domain.TelecontrolCommand, error) {
	deviceID := strings.TrimSpace(input.DeviceID)
	if deviceID == "" || containsMQTTWildcard(deviceID) {
		return domain.TelecontrolCommand{}, ErrInvalidDeviceID
	}

	commandName := strings.TrimSpace(input.Command)
	if commandName == "" || containsMQTTWildcard(commandName) {
		return domain.TelecontrolCommand{}, ErrInvalidCommandName
	}

	payload, err := normalizePayload(input.Payload)
	if err != nil {
		return domain.TelecontrolCommand{}, err
	}

	if input.TTL < 0 {
		return domain.TelecontrolCommand{}, ErrInvalidCommandTTL
	}

	if s.commandPublisher == nil {
		return domain.TelecontrolCommand{}, ErrCommandPublisherUnavailable
	}

	createdAt := s.now()
	command := domain.TelecontrolCommand{
		ID:            s.generateID(),
		DeviceID:      deviceID,
		Command:       commandName,
		Payload:       payload,
		CorrelationID: strings.TrimSpace(input.CorrelationID),
		RequestedBy:   strings.TrimSpace(input.RequestedBy),
		Status:        domain.TelecontrolCommandStatusAccepted,
		CreatedAt:     createdAt,
	}

	if input.TTL > 0 {
		expiresAt := createdAt.Add(input.TTL)
		command.ExpiresAt = &expiresAt
	}

	if err := s.commandRepository.Create(ctx, command); err != nil {
		return domain.TelecontrolCommand{}, err
	}

	publication, err := s.commandPublisher.PublishCommand(ctx, command)
	if err != nil {
		command.MarkFailed(s.now(), err.Error())
		if updateErr := s.commandRepository.Update(ctx, command); updateErr != nil {
			return command, fmt.Errorf("%w: %v: update command status: %v", ErrCommandPublishFailed, err, updateErr)
		}

		return command, fmt.Errorf("%w: %v", ErrCommandPublishFailed, err)
	}

	command.MarkPublished(s.now(), publication.Destination)
	if err := s.commandRepository.Update(ctx, command); err != nil {
		return domain.TelecontrolCommand{}, err
	}

	if err := s.runCommandIssuedHooks(ctx, command); err != nil {
		return command, err
	}

	return command, nil
}

func (s *TelecontrolService) GetCommand(ctx context.Context, id string) (domain.TelecontrolCommand, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.TelecontrolCommand{}, domain.ErrTelecontrolCommandNotFound
	}

	return s.commandRepository.GetByID(ctx, id)
}

func (s *TelecontrolService) ListCommands(ctx context.Context) ([]domain.TelecontrolCommand, error) {
	return s.commandRepository.List(ctx)
}

func (s *TelecontrolService) runCommandIssuedHooks(ctx context.Context, command domain.TelecontrolCommand) error {
	for _, hook := range s.commandHooks {
		if err := hook.HandleCommandIssued(ctx, command); err != nil {
			return fmt.Errorf("execute command issued hook: %w", err)
		}
	}

	return nil
}

func normalizePayload(payload json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return json.RawMessage(`{}`), nil
	}

	if !json.Valid(trimmed) {
		return nil, ErrInvalidCommandPayload
	}

	normalized := make([]byte, len(trimmed))
	copy(normalized, trimmed)

	return normalized, nil
}

func containsMQTTWildcard(value string) bool {
	return strings.ContainsAny(value, "+#")
}

func generateCommandID() string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err == nil {
		return "tcmd_" + hex.EncodeToString(buffer)
	}

	return "tcmd_fallback"
}
