package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SlotResponse struct {
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Start  string `json:"start" example:"2024-04-04T13:00:00Z" format:"date-time"`
	End    string `json:"end" example:"2024-04-04T13:00:00Z" format:"date-time"`
}

type SlotsListResponse struct {
	Slots []SlotResponse `json:"slots"`
}

type SlotHandler struct {
	service usecase.SlotsUseCase
}

func NewSlotHandler(service usecase.SlotsUseCase) *SlotHandler {
	return &SlotHandler{service: service}
}

// ListAvailableSlots godoc
// @Summary      List available slots for a room by date
// @Description  Returns a list of slots that are not occupied by an active booking for a specific room and date
// @Tags         Slots
// @Produce      json
// @Security     BearerAuth
// @Param        roomId  path      string  true  "Room ID (UUID)"
// @Param        date    query     string  true  "Date in ISO 8601 format (e.g., 2024-06-10)"
// @Success      200     {object}  SlotsListResponse "List of available slots"
// @Failure      400     {object}  http.ErrorResponse             "Invalid request"
// @Failure      401     {object}  http.ErrorResponse             "Unauthorized"
// @Failure      404     {object}  http.ErrorResponse             "Room not found"
// @Failure      500     {object}  http.ErrorResponse     "Internal server error"
// @Router       /rooms/{roomId}/slots/list [get]
func (h *SlotHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid room uuid format")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "missing required parameter: date")
		return
	}

	parsedDate, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid date format, expected YYYY-MM-DD")
		return
	}

	slots, err := h.service.GetSlotsByRoomId(r.Context(), roomID, parsedDate)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			renderError(w, http.StatusNotFound, ErrCodeRoomNotFound, "room not found")
		default:
			renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "internal server error")
		}
		return
	}

	resp := SlotsListResponse{
		Slots: make([]SlotResponse, 0, len(slots)),
	}

	for _, s := range slots {
		resp.Slots = append(resp.Slots, SlotResponse{
			ID:     s.Id.String(),
			RoomID: s.RoomId.String(),
			Start:  s.StartAt.Format(time.RFC3339),
			End:    s.EndAt.Format(time.RFC3339),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
