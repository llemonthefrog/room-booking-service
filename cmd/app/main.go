package main

import (
	postgres2 "avito-task/internal/infrastructure/postgres"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	routing "avito-task/internal/infrastructure/http"
	"avito-task/internal/services"
)

// @title           Booking Service API
// @version         1.0
// @description     API for conference room booking system.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter Bearer followed by a space and the JWT token.
func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("Unable to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(pingCtx)
	cancelPing()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	roomRepository := postgres2.NewRoomPostgresRepository(db)
	slotRepository := postgres2.NewSlotPostgresRepository(db)
	bookingRepository := postgres2.NewBookingPostgresRepository(db)
	scheduleRepository := postgres2.NewSchedulePostgresRepository(db)
	userRepository := postgres2.NewUserPostgresRepository(db)

	jwtService := services.NewJWTService(jwtSecret, 24*time.Hour)
	dummyAuthService := services.NewDummyAuthService(jwtSecret, userRepository)
	roomService := services.NewRoomService(roomRepository)
	scheduleService := services.NewScheduleService(scheduleRepository, roomRepository)
	slotService := services.NewSlotsService(slotRepository, scheduleRepository, roomRepository)
	bookingService := services.NewBookingService(bookingRepository, slotRepository)
	authService := services.NewAuthService(jwtSecret, userRepository)

	router := routing.InitRouter(
		jwtService,
		dummyAuthService,
		roomService,
		scheduleService,
		slotService,
		bookingService,
		authService,
	)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
