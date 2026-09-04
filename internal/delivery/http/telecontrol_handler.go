package httpdelivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ms-telecontrol/internal/domain"
	"ms-telecontrol/internal/usecase"
)

type telecontrolHandler struct {
	telecontrolUsecase TelecontrolUsecase
}

type issueCommandRequest struct {
	DeviceID      string          `json:"device_id"`
	Command       string          `json:"command"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlation_id"`
	RequestedBy   string          `json:"requested_by"`
	TTLSeconds    int64           `json:"ttl_seconds"`
}

func newTelecontrolHandler(telecontrolUsecase TelecontrolUsecase) telecontrolHandler {
	return telecontrolHandler{
		telecontrolUsecase: telecontrolUsecase,
	}
}

func (h telecontrolHandler) issueCommand(w http.ResponseWriter, r *http.Request) {
	var request issueCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	command, err := h.telecontrolUsecase.IssueCommand(r.Context(), usecase.IssueCommandInput{
		DeviceID:      request.DeviceID,
		Command:       request.Command,
		Payload:       request.Payload,
		CorrelationID: request.CorrelationID,
		RequestedBy:   request.RequestedBy,
		TTL:           time.Duration(request.TTLSeconds) * time.Second,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidDeviceID),
			errors.Is(err, usecase.ErrInvalidCommandName),
			errors.Is(err, usecase.ErrInvalidCommandPayload),
			errors.Is(err, usecase.ErrInvalidCommandTTL):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, usecase.ErrCommandPublisherUnavailable):
			writeError(w, http.StatusServiceUnavailable, err.Error())
		case errors.Is(err, usecase.ErrCommandPublishFailed):
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"error":   err.Error(),
				"command": command,
			})
		case errors.Is(err, domain.ErrTelecontrolCommandAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	writeJSON(w, http.StatusAccepted, command)
}

func (h telecontrolHandler) listCommands(w http.ResponseWriter, r *http.Request) {
	commands, err := h.telecontrolUsecase.ListCommands(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, commands)
}

func (h telecontrolHandler) getCommandByID(w http.ResponseWriter, r *http.Request) {
	command, err := h.telecontrolUsecase.GetCommand(r.Context(), r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTelecontrolCommandNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	writeJSON(w, http.StatusOK, command)
}
