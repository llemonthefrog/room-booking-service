package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     domain.Role `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Id        uuid.UUID   `json:"id"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	CreatedAt time.Time   `json:"createdAt" example:"2024-04-04T13:00:00Z" format:"date-time"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type AuthHandler struct {
	authService usecase.AuthUseCase
}

func NewAuthHandler(authService usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      User registration
// @Description  Create new user and return his data
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      http.RegisterRequest  true  "Registration data"
// @Success      201      {object}  map[string]http.UserResponse "User create"
// @Failure      400      {object}  http.ErrorResponse           "Invalid request or email already taken"
// @Failure      500      {object}  http.ErrorResponse   "Internal server data"
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	user, token, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidData) {
			renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid email, password or role")
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "user with this email already exists")
			return
		}

		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "failed to register user")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": UserResponse{
			Id:        user.Id,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
		"token": token,
	})
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user by email and password, returning a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      http.LoginRequest   true  "User credentials"
// @Success      200      {object}  http.TokenResponse  "Successfully authenticated"
// @Failure      401      {object}  http.ErrorResponse  "Invalid email or password"
// @Failure      500      {object}  http.ErrorResponse "Internal server error"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			renderError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid email or password")
			return
		}
		renderError(w, http.StatusInternalServerError, ErrCodeInternalError, "failed to login")
		return
	}

	json.NewEncoder(w).Encode(TokenResponse{Token: token})
}
