package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
)

type DummyLoginRequest struct {
	Role string `json:"role" example:"user"`
}

type DummyAuthHandler struct {
	service usecase.DummyAuthUseCase
}

func NewDummyAuthHandler(service usecase.DummyAuthUseCase) *DummyAuthHandler {
	return &DummyAuthHandler{service: service}
}

// DummyLogin godoc
// @Summary      Get string JWT token by role
// @Description  Return string JWT token by role (admin / user) with fixed UUID.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      DummyLoginRequest  true  "User role"
// @Success      200      {object}  TokenResponse
// @Failure      400      {object}  http.ErrorResponse "Invalid role"
// @Failure      500      {object}  http.ErrorResponse "Internal error"
// @Router       /dummyLogin [post]
func (h *DummyAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid json format")
		return
	}

	token, err := h.service.Login(r.Context(), req.Role)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid role")
			return
		}

		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TokenResponse{Token: token})
}
