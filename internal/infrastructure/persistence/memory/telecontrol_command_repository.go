package memory

import (
	"context"
	"sort"
	"sync"

	"ms-telecontrol/internal/domain"
)

type TelecontrolCommandRepository struct {
	mu       sync.RWMutex
	commands map[string]domain.TelecontrolCommand
}

func NewTelecontrolCommandRepository() *TelecontrolCommandRepository {
	return &TelecontrolCommandRepository{
		commands: make(map[string]domain.TelecontrolCommand),
	}
}

func (r *TelecontrolCommandRepository) Create(_ context.Context, command domain.TelecontrolCommand) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[command.ID]; exists {
		return domain.ErrTelecontrolCommandAlreadyExists
	}

	r.commands[command.ID] = command.Clone()

	return nil
}

func (r *TelecontrolCommandRepository) Update(_ context.Context, command domain.TelecontrolCommand) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[command.ID]; !exists {
		return domain.ErrTelecontrolCommandNotFound
	}

	r.commands[command.ID] = command.Clone()

	return nil
}

func (r *TelecontrolCommandRepository) GetByID(_ context.Context, id string) (domain.TelecontrolCommand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	command, exists := r.commands[id]
	if !exists {
		return domain.TelecontrolCommand{}, domain.ErrTelecontrolCommandNotFound
	}

	return command.Clone(), nil
}

func (r *TelecontrolCommandRepository) List(_ context.Context) ([]domain.TelecontrolCommand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]domain.TelecontrolCommand, 0, len(r.commands))
	for _, command := range r.commands {
		commands = append(commands, command.Clone())
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].CreatedAt.Before(commands[j].CreatedAt)
	})

	return commands, nil
}
