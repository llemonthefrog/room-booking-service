package http

import (
	"encoding/json"
	"net/http"
)

const (
	ErrCodeInvalidRequest    = "INVALID_REQUEST"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeRoomNotFound      = "ROOM_NOT_FOUND"
	ErrCodeSlotNotFound      = "SLOT_NOT_FOUND"
	ErrCodeSlotAlreadyBooked = "SLOT_ALREADY_BOOKED"
	ErrCodeBookingNotFound   = "BOOKING_NOT_FOUND"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeScheduleExists    = "SCHEDULE_EXISTS"
	ErrCodeInternalError     = "INTERNAL_ERROR"
)

type ErrorDetails struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"Detailed error message"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

func renderError(w http.ResponseWriter, status int, code string, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: msg,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
