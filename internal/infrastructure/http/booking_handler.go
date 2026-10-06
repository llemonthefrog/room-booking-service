package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateBookingRequest struct {
	SlotID uuid.UUID `json:"slotId"`
}

type BookingDTO struct {
	ID        string `json:"id"`
	SlotID    string `json:"slotId"`
	UserID    string `json:"userId"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt" example:"2024-04-04T13:00:00Z" format:"date-time"`
}

type BookingResponse struct {
	Booking BookingDTO `json:"booking"`
}

type PaginatedBookingsResponse struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type BookingListResponse struct {
	Bookings   []BookingDTO              `json:"bookings"`
	Pagination PaginatedBookingsResponse `json:"pagination,omitempty"`
}

type BookingHandler struct {
	service usecase.BookingUseCase
}

func NewBookingHandler(service usecase.BookingUseCase) *BookingHandler {
	return &BookingHandler{service: service}
}

// CreateBooking godoc
// @Summary      Create a booking for a slot
// @Description  Creates a new booking for a specific slot. Only accessible by users with the 'user' role
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      http.CreateBookingRequest  true  "Booking details"
// @Success      201      {object}  http.BookingResponse "Booking created successfully"
// @Failure      400      {object}  http.ErrorResponse              "Invalid request"
// @Failure      401      {object}  http.ErrorResponse              "Unauthorized"
// @Failure      403      {object}  http.ErrorResponse              "Forbidden"
// @Failure      404      {object}  http.ErrorResponse              "Slot not found"
// @Failure      409      {object}  http.ErrorResponse              "Slot already booked"
// @Failure      500      {object}  http.ErrorResponse      "Internal server error"
// @Router       /bookings/create [post]
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uId, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		renderError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "authentication required")
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid json body")
		return
	}

	booking, err := h.service.Create(r.Context(), uId, req.SlotID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			renderError(w, http.StatusNotFound, ErrCodeSlotNotFound, "slot not found")
		case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyExists):
			renderError(w, http.StatusConflict, ErrCodeSlotAlreadyBooked, "slot is already booked")
		default:
			renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(BookingResponse{
		Booking: mapDomainToBookingDTO(booking),
	})
}

// ListBookings godoc
// @Summary      List all bookings with pagination
// @Description  Returns a paginated list of all bookings in the system. Accessible only by users with the 'admin' role.
// @Tags         Bookings
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int  false  "Page number (starting from 1). Default is 1"
// @Param        pageSize  query     int  false  "Number of items per page. Default 20, max 100"
// @Success      200       {object}  http.PaginatedBookingsResponse "List of all bookings"
// @Failure      400       {object}  http.ErrorResponse             "Invalid pagination parameters"
// @Failure      401       {object}  http.ErrorResponse             "Unauthorized"
// @Failure      403       {object}  http.ErrorResponse             "Forbidden"
// @Failure      500       {object}  http.ErrorResponse     "Internal server error"
// @Router       /bookings/list [get]
func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()

	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			renderError(
				w,
				http.StatusBadRequest,
				ErrCodeInvalidRequest,
				"invalid 'page' parameter: must be a positive integer",
			)
			return
		}
		page = p
	}

	pageSize := 20
	if pageSizeStr := query.Get("pageSize"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil || ps < 1 || ps > 100 {
			renderError(
				w,
				http.StatusBadRequest,
				ErrCodeInvalidRequest,
				"invalid 'pageSize' parameter: must be between 1 and 100",
			)
			return
		}
		pageSize = ps
	}

	if page-1 > math.MaxInt/pageSize {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "page is too large")
		return
	}
	offset := (page - 1) * pageSize

	books, total, err := h.service.GetAll(r.Context(), pageSize, offset)
	if err != nil {
		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "internal server error")
		return
	}

	resp := BookingListResponse{
		Bookings: make([]BookingDTO, 0, len(books)),
		Pagination: PaginatedBookingsResponse{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}

	for _, b := range books {
		resp.Bookings = append(resp.Bookings, mapDomainToBookingDTO(b))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UserList godoc
// @Summary      List bookings for the current user
// @Description  Returns a list of bookings belonging to the user identified by the JWT token
// @Tags         Bookings
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  http.BookingListResponse "List of future bookings"
// @Failure      401      {object}  http.ErrorResponse              "Unauthorized"
// @Failure      403      {object}  http.ErrorResponse              "Forbidden"
// @Failure      500      {object}  http.ErrorResponse      "Internal server error"
// @Router       /bookings/my [get]
func (h *BookingHandler) UserList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uId, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		renderError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "authentication required")
		return
	}

	books, err := h.service.GetUserBookings(r.Context(), uId)
	if err != nil {
		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "internal server error")
		return
	}

	resp := BookingListResponse{
		Bookings: make([]BookingDTO, 0, len(books)),
	}
	for _, b := range books {
		resp.Bookings = append(resp.Bookings, mapDomainToBookingDTO(b))
	}

	json.NewEncoder(w).Encode(resp)
}

// CancelBooking godoc
// @Summary      Cancel a booking
// @Description  Cancels a booking by its ID. Only the owner of the booking can perform this action
// @Tags         Bookings
// @Produce      json
// @Security     BearerAuth
// @Param        bookingId  path      string  true  "ID of the booking to cancel (UUID)"
// @Success      200      {object}  http.BookingResponse "Booking cancelled"
// @Failure      401      {object}  http.ErrorResponse              "Unauthorized"
// @Failure      403      {object}  http.ErrorResponse              "Forbidden"
// @Failure      404      {object}  http.ErrorResponse              "Booking not found"
// @Failure      500      {object}  http.ErrorResponse      "Internal server error"
// @Router       /bookings/{bookingId}/cancel [post]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uId, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		renderError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "authentication required")
		return
	}

	bIdStr := chi.URLParam(r, "bookingId")
	bId, err := uuid.Parse(bIdStr)
	if err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid booking uuid")
		return
	}

	booking, err := h.service.Cancel(r.Context(), uId, bId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDomainValidation):
			renderError(w, http.StatusForbidden, ErrCodeForbidden, "access denied")
		case errors.Is(err, domain.ErrNotFound):
			renderError(
				w,
				http.StatusNotFound,
				ErrCodeBookingNotFound,
				"resource not found",
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

	json.NewEncoder(w).Encode(BookingResponse{
		Booking: mapDomainToBookingDTO(booking),
	})
}

func mapDomainToBookingDTO(b *domain.Book) BookingDTO {
	return BookingDTO{
		ID:        b.Id.String(),
		SlotID:    b.SlotId.String(),
		UserID:    b.UserId.String(),
		Status:    string(b.Status),
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
	}
}
