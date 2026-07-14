package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func InitRouter(
	jwtUseCase usecase.JWTUseCase,
	dummyAuthUseCase usecase.DummyAuthUseCase,
	roomUseCase usecase.RoomUseCase,
	scheduleUseCase usecase.ScheduleUseCase,
	slotsUseCase usecase.SlotsUseCase,
	bookingUseCase usecase.BookingUseCase,
	authUseCase usecase.AuthUseCase,
) *http.ServeMux {
	mux := http.NewServeMux()

	dummyAuthHandler := NewDummyAuthHandler(dummyAuthUseCase)
	authHandler := NewAuthHandler(authUseCase)
	roomHandler := NewRoomHandler(roomUseCase)
	scheduleHandler := NewScheduleHandler(scheduleUseCase)
	slotHandler := NewSlotHandler(slotsUseCase)
	bookingHandler := NewBookingHandler(bookingUseCase)

	authMiddleware := AuthMiddleware(jwtUseCase)
	adminRoleMiddleware := RoleMiddleware(domain.RoleAdmin)
	userRoleMiddleware := RoleMiddleware(domain.RoleUser)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	mux.Handle("POST /dummyLogin", http.HandlerFunc(dummyAuthHandler.Login))

	mux.Handle("POST /register", http.HandlerFunc(authHandler.Register))
	mux.Handle("POST /login", http.HandlerFunc(authHandler.Login))

	mux.Handle("GET /rooms/list", Chain(
		http.HandlerFunc(roomHandler.List),
		authMiddleware,
	))

	mux.Handle("POST /rooms/create", Chain(
		http.HandlerFunc(roomHandler.Create),
		authMiddleware,
		adminRoleMiddleware,
	))

	mux.Handle("POST /rooms/{roomId}/schedule/create", Chain(
		http.HandlerFunc(scheduleHandler.Create),
		authMiddleware,
		adminRoleMiddleware,
	))

	mux.Handle("GET /rooms/{roomId}/slots/list", Chain(
		http.HandlerFunc(slotHandler.List),
		authMiddleware,
	))

	mux.Handle("POST /bookings/create", Chain(
		http.HandlerFunc(bookingHandler.Create),
		authMiddleware,
		userRoleMiddleware,
	))

	mux.Handle("GET /bookings/list", Chain(
		http.HandlerFunc(bookingHandler.List),
		authMiddleware,
		adminRoleMiddleware,
	))

	mux.Handle("GET /bookings/my", Chain(
		http.HandlerFunc(bookingHandler.UserList),
		authMiddleware,
		userRoleMiddleware,
	))

	mux.Handle("POST /bookings/{bookingId}/cancel", Chain(
		http.HandlerFunc(bookingHandler.Cancel),
		authMiddleware,
		userRoleMiddleware,
	))

	return mux
}
