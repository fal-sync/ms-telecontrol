package port

import (
	"context"

	"ms-telecontrol/internal/domain"
)

type CommandPublication struct {
	Destination string
}

type CommandPublisher interface {
	PublishCommand(ctx context.Context, command domain.TelecontrolCommand) (CommandPublication, error)
}
