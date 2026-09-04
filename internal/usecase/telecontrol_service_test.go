package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"ms-telecontrol/internal/domain"
	"ms-telecontrol/internal/infrastructure/persistence/memory"
	"ms-telecontrol/internal/usecase/port"
)

func TestIssueCommand(t *testing.T) {
	publisher := &publisherSpy{
		publication: port.CommandPublication{Destination: "telecontrol/device-1/commands"},
	}
	service := NewTelecontrolService(memory.NewTelecontrolCommandRepository(), publisher)

	command, err := service.IssueCommand(context.Background(), IssueCommandInput{
		DeviceID: "device-1",
		Command:  "relay.set",
		Payload:  json.RawMessage(`{"state":"on"}`),
	})
	if err != nil {
		t.Fatalf("IssueCommand returned error: %v", err)
	}

	if command.ID == "" {
		t.Fatal("expected generated command ID")
	}

	if command.Status != domain.TelecontrolCommandStatusPublished {
		t.Fatalf("expected status %q, got %q", domain.TelecontrolCommandStatusPublished, command.Status)
	}

	if command.Destination != "telecontrol/device-1/commands" {
		t.Fatalf("expected command destination to be set, got %q", command.Destination)
	}

	if !publisher.called {
		t.Fatal("expected publisher to be called")
	}

	if publisher.command.DeviceID != "device-1" {
		t.Fatalf("expected published device_id to be device-1, got %q", publisher.command.DeviceID)
	}
}

func TestIssueCommandValidation(t *testing.T) {
	service := NewTelecontrolService(memory.NewTelecontrolCommandRepository(), &publisherSpy{})

	_, err := service.IssueCommand(context.Background(), IssueCommandInput{
		DeviceID: "",
		Command:  "relay.set",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !errors.Is(err, ErrInvalidDeviceID) {
		t.Fatalf("expected ErrInvalidDeviceID, got %v", err)
	}
}

func TestIssueCommandRequiresPublisher(t *testing.T) {
	service := NewTelecontrolService(memory.NewTelecontrolCommandRepository(), nil)

	_, err := service.IssueCommand(context.Background(), IssueCommandInput{
		DeviceID: "device-1",
		Command:  "relay.set",
	})
	if err == nil {
		t.Fatal("expected publisher unavailable error")
	}

	if !errors.Is(err, ErrCommandPublisherUnavailable) {
		t.Fatalf("expected ErrCommandPublisherUnavailable, got %v", err)
	}
}

func TestIssueCommandStoresFailedStatusWhenPublishFails(t *testing.T) {
	repository := memory.NewTelecontrolCommandRepository()
	publisher := &publisherSpy{err: errors.New("broker unavailable")}
	service := NewTelecontrolService(repository, publisher)

	command, err := service.IssueCommand(context.Background(), IssueCommandInput{
		DeviceID: "device-1",
		Command:  "relay.set",
	})
	if err == nil {
		t.Fatal("expected publish error")
	}

	if !errors.Is(err, ErrCommandPublishFailed) {
		t.Fatalf("expected ErrCommandPublishFailed, got %v", err)
	}

	stored, getErr := repository.GetByID(context.Background(), command.ID)
	if getErr != nil {
		t.Fatalf("GetByID returned error: %v", getErr)
	}

	if stored.Status != domain.TelecontrolCommandStatusFailed {
		t.Fatalf("expected stored status %q, got %q", domain.TelecontrolCommandStatusFailed, stored.Status)
	}

	if stored.FailureReason == "" {
		t.Fatal("expected failure reason to be stored")
	}
}

func TestGetCommandNotFound(t *testing.T) {
	service := NewTelecontrolService(memory.NewTelecontrolCommandRepository(), &publisherSpy{})

	_, err := service.GetCommand(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected not found error")
	}

	if !errors.Is(err, domain.ErrTelecontrolCommandNotFound) {
		t.Fatalf("expected domain.ErrTelecontrolCommandNotFound, got %v", err)
	}
}

func TestIssueCommandRunsHooks(t *testing.T) {
	hook := &hookSpy{}
	service := NewTelecontrolService(
		memory.NewTelecontrolCommandRepository(),
		&publisherSpy{publication: port.CommandPublication{Destination: "telecontrol/device-1/commands"}},
		hook,
	)

	command, err := service.IssueCommand(context.Background(), IssueCommandInput{
		DeviceID: "device-1",
		Command:  "relay.set",
	})
	if err != nil {
		t.Fatalf("IssueCommand returned error: %v", err)
	}

	if !hook.called {
		t.Fatal("expected command issued hook to be called")
	}

	if hook.command.ID != command.ID {
		t.Fatalf("expected hook command ID to be %q, got %q", command.ID, hook.command.ID)
	}
}

type publisherSpy struct {
	called      bool
	command     domain.TelecontrolCommand
	publication port.CommandPublication
	err         error
}

func (p *publisherSpy) PublishCommand(_ context.Context, command domain.TelecontrolCommand) (port.CommandPublication, error) {
	p.called = true
	p.command = command

	if p.err != nil {
		return port.CommandPublication{}, p.err
	}

	return p.publication, nil
}

type hookSpy struct {
	called  bool
	command domain.TelecontrolCommand
}

func (h *hookSpy) HandleCommandIssued(_ context.Context, command domain.TelecontrolCommand) error {
	h.called = true
	h.command = command
	return nil
}
