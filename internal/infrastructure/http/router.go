package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"net/http"

	_ "avito-task/docs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)

	dummyAuthHandler := NewDummyAuthHandler(dummyAuthUseCase)
	authHandler := NewAuthHandler(authUseCase)
	roomHandler := NewRoomHandler(roomUseCase)
	scheduleHandler := NewScheduleHandler(scheduleUseCase)
	slotHandler := NewSlotHandler(slotsUseCase)
	bookingHandler := NewBookingHandler(bookingUseCase)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/_info", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Post("/dummyLogin", dummyAuthHandler.Login)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(jwtUseCase))
		r.Get("/rooms/list", roomHandler.List)
		r.Get("/rooms/{roomId}/slots/list", slotHandler.List)

		r.Group(func(r chi.Router) {
			r.Use(RoleMiddleware(domain.RoleAdmin))
			r.Post("/rooms/create", roomHandler.Create)
			r.Post("/rooms/{roomId}/schedule/create", scheduleHandler.Create)
			r.Get("/bookings/list", bookingHandler.List)
		})
		r.Group(func(r chi.Router) {
			r.Use(RoleMiddleware(domain.RoleUser))
			r.Post("/bookings/create", bookingHandler.Create)
			r.Get("/bookings/my", bookingHandler.UserList)
			r.Post("/bookings/{bookingId}/cancel", bookingHandler.Cancel)
		})
	})

	return r
}
