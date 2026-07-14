package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime" example:"09:00"`
	EndTime    string `json:"endTime" example:"18:00"`
}

type ScheduleDTO struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime" example:"2024-04-04T13:00:00Z" format:"date-time"`
	EndTime    string `json:"endTime" example:"2024-04-04T13:00:00Z" format:"date-time"`
}

type ScheduleResponse struct {
	Schedule ScheduleDTO `json:"schedule"`
}

type ScheduleHandler struct {
	service usecase.ScheduleUseCase
}

func NewScheduleHandler(service usecase.ScheduleUseCase) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// CreateSchedule godoc
// @Summary      Create room availability schedule
// @Description  Defines the availability rules for a conference room. Can be created only once per room
// @Tags         Schedules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        roomId   path      string          true  "Room ID (UUID)"
// @Param        request  body      http.CreateScheduleRequest true  "Schedule definition"
// @Success      201      {object}  http.ScheduleResponse "Schedule created successfully"
// @Failure      400      {object}  http.ErrorResponse               "Invalid request or invalid daysOfWeek values"
// @Failure      401      {object}  http.ErrorResponse               "Unauthorized"
// @Failure      403      {object}  http.ErrorResponse               "Forbidden"
// @Failure      404      {object}  http.ErrorResponse               "Room not found"
// @Failure      409      {object}  http.ErrorResponse               "Conflict: schedule already exists for this room"
// @Failure      500      {object}  http.ErrorResponse       "Internal server error"
// @Router       /rooms/{roomId}/schedule/create [post]
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomIDStr := r.PathValue("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid room uuid format")
		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "failed to decode json")
		return
	}

	sch, err := h.service.Create(r.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		switch err {
		case domain.ErrNotFound:
			renderError(w, http.StatusNotFound, ErrCodeRoomNotFound, "room not found")
		case domain.ErrAlreadyExists:
			renderError(
				w,
				http.StatusConflict,
				ErrCodeScheduleExists,
				"schedule for this room already exists",
			)
		case domain.ErrInvalidData:
			renderError(
				w,
				http.StatusBadRequest,
				ErrCodeInvalidRequest,
				"invalid days of week or time format",
			)
		default:
			renderError(
				w,
				http.StatusInternalServerError,
				ErrCodeInternalError,
				"internal server error",
			)
		}
		return
	}

	resp := ScheduleResponse{
		Schedule: ScheduleDTO{
			ID:         sch.Id.String(),
			RoomID:     sch.RoomId.String(),
			DaysOfWeek: sch.DaysOfWeek,
			StartTime:  sch.StartTime.Format("15:04"),
			EndTime:    sch.EndTime.Format("15:04"),
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
