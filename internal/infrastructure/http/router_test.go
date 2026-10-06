package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"avito-task/internal/services"
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type bookingStub struct {
	usecase.BookingUseCase
	cancel func(context.Context, uuid.UUID, uuid.UUID) (*domain.Book, error)
}

func (s bookingStub) Cancel(ctx context.Context, userID, bookingID uuid.UUID) (*domain.Book, error) {
	return s.cancel(ctx, userID, bookingID)
}

func TestRouter(t *testing.T) {
	jwtSvc := services.NewJWTService("test-secret", time.Hour)
	userID := uuid.New()
	userToken, err := jwtSvc.GenerateToken(userID.String(), domain.RoleUser)
	if err != nil {
		t.Fatal(err)
	}
	adminToken, err := jwtSvc.GenerateToken(uuid.NewString(), domain.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	bookingID := uuid.New()
	booking := bookingStub{cancel: func(_ context.Context, uid, bid uuid.UUID) (*domain.Book, error) {
		if uid != userID || bid != bookingID {
			t.Errorf("incorrect user or path parameter: %s, %s", uid, bid)
		}
		return nil, fmt.Errorf("cancel: %w", domain.ErrDomainValidation)
	}}
	router := InitRouter(jwtSvc, nil, nil, nil, nil, booking, nil)

	for _, tt := range []struct {
		name, method, path, token string
		status                    int
	}{
		{"health", "GET", "/_info", "", 200},
		{"unknown path", "GET", "/missing", "", 404},
		{"wrong method", "DELETE", "/rooms/list", "", 405},
		{"missing token", "GET", "/rooms/list", "", 401},
		{"invalid token", "GET", "/rooms/list", "invalid", 401},
		{"user cannot create room", "POST", "/rooms/create", userToken, 403},
		{"admin cannot book", "POST", "/bookings/create", adminToken, 403},
		{"schedule path", "POST", "/rooms/bad-id/schedule/create", adminToken, 400},
		{"slots path", "GET", "/rooms/bad-id/slots/list", userToken, 400},
		{"booking path and owner", "POST", "/bookings/" + bookingID.String() + "/cancel", userToken, 403},
		{"pagination overflow", "GET", fmt.Sprintf("/bookings/list?page=%d&pageSize=100", math.MaxInt), adminToken, 400},
		{"swagger redirect", "GET", "/swagger", "", 301},
		{"swagger UI", "GET", "/swagger/index.html", "", 200},
		{"swagger document", "GET", "/swagger/doc.json", "", 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "bearer "+tt.token)
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tt.status {
				t.Fatalf("status = %d, want %d; body: %s", res.Code, tt.status, res.Body)
			}
		})
	}
}

func TestBookingHandlersRequireIdentity(t *testing.T) {
	h := NewBookingHandler(nil)
	for _, handler := range []http.HandlerFunc{h.Create, h.Cancel, h.UserList} {
		res := httptest.NewRecorder()
		handler(res, httptest.NewRequest("POST", "/", strings.NewReader(`{}`)))
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", res.Code)
		}
	}
}
