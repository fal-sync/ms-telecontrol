package httpdelivery

import (
	"context"
	"net/http"

	"ms-telecontrol/internal/domain"
	"ms-telecontrol/internal/usecase"
)

type TelecontrolUsecase interface {
	IssueCommand(rctx context.Context, input usecase.IssueCommandInput) (domain.TelecontrolCommand, error)
	GetCommand(rctx context.Context, id string) (domain.TelecontrolCommand, error)
	ListCommands(rctx context.Context) ([]domain.TelecontrolCommand, error)
}

func NewRouter(telecontrolUsecase TelecontrolUsecase) http.Handler {
	telecontrolHandler := newTelecontrolHandler(telecontrolUsecase)
	healthHandler := newHealthHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.check)
	mux.HandleFunc("POST /telecontrol/commands", telecontrolHandler.issueCommand)
	mux.HandleFunc("GET /telecontrol/commands", telecontrolHandler.listCommands)
	mux.HandleFunc("GET /telecontrol/commands/{id}", telecontrolHandler.getCommandByID)

	return mux
}
