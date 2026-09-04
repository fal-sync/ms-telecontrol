package port

import (
	"context"

	"ms-telecontrol/internal/domain"
)

type CommandIssuedHook interface {
	HandleCommandIssued(ctx context.Context, command domain.TelecontrolCommand) error
}
