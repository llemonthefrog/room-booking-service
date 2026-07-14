package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"encoding/json"
	"net/http"
)

type CreateRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
	CreatedAt   string  `json:"createdAt" example:"2024-04-04T13:00:00Z"`
}

type RoomResponse struct {
	Room *domain.Room `json:"room"`
}

type RoomListResponse struct {
	Rooms []*domain.Room `json:"rooms"`
}

type RoomHandler struct {
	service usecase.RoomUseCase
}

func NewRoomHandler(service usecase.RoomUseCase) *RoomHandler {
	return &RoomHandler{service: service}
}

// Create godoc
// @Summary      Create room
// @Tags         Rooms
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      CreateRoomRequest  true  "Room data"
// @Success      201      {object}  RoomResponse
// @Failure      400,401,403,500 {object} ErrorResponse
// @Router       /rooms/create [post]
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid json")
		return
	}

	room, err := h.service.Create(r.Context(), req.Name)
	if err != nil {
		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RoomResponse{Room: room})
}

// List godoc
// @Summary      Rooms list
// @Tags         Rooms
// @Security     BearerAuth
// @Produce      json
// @Success      200      {object}  RoomListResponse
// @Failure      401,500  {object}  ErrorResponse
// @Router       /rooms/list [get]
func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rooms, err := h.service.GetList(r.Context())
	if err != nil {
		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, err.Error())
		return
	}

	if rooms == nil {
		rooms = []*domain.Room{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RoomListResponse{Rooms: rooms})
}
