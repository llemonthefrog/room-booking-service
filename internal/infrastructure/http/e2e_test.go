package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"avito-task/internal/infrastructure/postgres"
	"avito-task/internal/services"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func executeRequest(req *http.Request, router http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func applyMigrations(t *testing.T, db *sql.DB) {
	currDir, _ := os.Getwd()
	rootPath := currDir
	for {
		if _, err := os.Stat(filepath.Join(rootPath, "migrations")); err == nil {
			break
		}
		parent := filepath.Dir(rootPath)
		if parent == rootPath {
			t.Fatalf("Could not find migrations directory starting from %s", currDir)
		}
		rootPath = parent
	}

	migrationsDir := filepath.Join(rootPath, "migrations")

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	for _, fileName := range migrationFiles {
		t.Logf("Applying migration: %s", fileName)

		content, err := os.ReadFile(filepath.Join(migrationsDir, fileName))
		if err != nil {
			t.Fatalf("Failed to read migration file %s: %v", fileName, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("Failed to execute migration %s: %v", fileName, err)
		}
	}
	t.Log("All migrations applied successfully")
}

func setupRouter(t *testing.T) http.Handler {
	t.Helper()
	router, _ := setupApp(t)
	return router
}

func setupApp(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	if testing.Short() {
		t.Skip("E2E tests require Docker")
	}
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("booking_test"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)

	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Unable to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	applyMigrations(t, db)

	jwtSecret := "test_e2e_secret"

	roomRepo := postgres.NewRoomPostgresRepository(db)
	slotRepo := postgres.NewSlotPostgresRepository(db)
	bookingRepo := postgres.NewBookingPostgresRepository(db)
	scheduleRepo := postgres.NewSchedulePostgresRepository(db)
	userRepo := postgres.NewUserPostgresRepository(db)

	jwtSvc := services.NewJWTService(jwtSecret, 24*time.Hour)
	dummyAuthSvc := services.NewDummyAuthService(jwtSecret, userRepo)
	roomSvc := services.NewRoomService(roomRepo)
	scheduleSvc := services.NewScheduleService(scheduleRepo, roomRepo)
	slotSvc := services.NewSlotsService(slotRepo, scheduleRepo, roomRepo)
	bookingSvc := services.NewBookingService(bookingRepo, slotRepo)
	authSvc := services.NewAuthService(jwtSecret, userRepo)

	return InitRouter(jwtSvc, dummyAuthSvc, roomSvc, scheduleSvc, slotSvc, bookingSvc, authSvc), db
}

func getDummyToken(t *testing.T, router http.Handler, role string) string {
	payload := []byte(fmt.Sprintf(`{"role":"%s"}`, role))
	req, _ := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, router)
	if response.Code != http.StatusOK {
		t.Fatalf("Failed to login as %s: %s", role, response.Body.String())
	}

	var res map[string]string
	json.Unmarshal(response.Body.Bytes(), &res)
	return res["token"]
}

func TestE2E_CreateBookingFlow(t *testing.T) {
	router := setupRouter(t)

	adminToken := getDummyToken(t, router, "admin")

	roomName := fmt.Sprintf("E2E Room %d", time.Now().UnixNano())
	roomPayload := []byte(fmt.Sprintf(`{"name":"%s", "description":"Test", "capacity":10}`, roomName))
	req, _ := http.NewRequest("POST", "/rooms/create", bytes.NewBuffer(roomPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resRoom := executeRequest(req, router)
	if resRoom.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for Room, got %d. Body: %s", resRoom.Code, resRoom.Body.String())
	}

	var roomResp map[string]map[string]interface{}
	json.Unmarshal(resRoom.Body.Bytes(), &roomResp)
	roomID := roomResp["room"]["Id"].(string)

	schedulePayload := []byte(`{"daysOfWeek":[1,2,3,4,5,6,7], "startTime":"09:00", "endTime":"18:00"}`)
	schedulePath := fmt.Sprintf("/rooms/%s/schedule/create", roomID)
	req, _ = http.NewRequest("POST", schedulePath, bytes.NewBuffer(schedulePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resSchedule := executeRequest(req, router)
	if resSchedule.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for Schedule, got %d. Body: %s", resSchedule.Code, resSchedule.Body.String())
	}

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	slotsPath := fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, tomorrow)
	req, _ = http.NewRequest("GET", slotsPath, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resSlots := executeRequest(req, router)
	if resSlots.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for Slots, got %d", resSlots.Code)
	}

	var slotsResp map[string][]map[string]interface{}
	json.Unmarshal(resSlots.Body.Bytes(), &slotsResp)

	slots := slotsResp["slots"]
	if len(slots) == 0 {
		t.Fatalf("No slots generated for E2E test")
	}
	slotID := slots[0]["id"].(string)

	userToken := getDummyToken(t, router, "user")

	bookingPayload := []byte(fmt.Sprintf(`{"slotId":"%s"}`, slotID))
	req, _ = http.NewRequest("POST", "/bookings/create", bytes.NewBuffer(bookingPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)

	resBooking := executeRequest(req, router)
	if resBooking.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for Booking, got %d: %s", resBooking.Code, resBooking.Body.String())
	}
}

func TestE2E_CancelBookingFlow(t *testing.T) {
	router := setupRouter(t)

	adminToken := getDummyToken(t, router, "admin")

	roomName := fmt.Sprintf("E2E Cancel Room %d", time.Now().UnixNano())
	roomPayload := []byte(fmt.Sprintf(`{"name":"%s"}`, roomName))
	req, _ := http.NewRequest("POST", "/rooms/create", bytes.NewBuffer(roomPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resRoom := executeRequest(req, router)

	var roomResp map[string]map[string]interface{}
	json.Unmarshal(resRoom.Body.Bytes(), &roomResp)
	roomID := roomResp["room"]["Id"].(string)

	schedulePayload := []byte(`{"daysOfWeek":[1,2,3,4,5,6,7], "startTime":"10:00", "endTime":"12:00"}`)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/rooms/%s/schedule/create", roomID), bytes.NewBuffer(schedulePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	executeRequest(req, router)

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	req, _ = http.NewRequest("GET", fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, tomorrow), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resSlots := executeRequest(req, router)

	var slotsResp map[string][]map[string]interface{}
	json.Unmarshal(resSlots.Body.Bytes(), &slotsResp)
	slotID := slotsResp["slots"][0]["id"].(string)

	userToken := getDummyToken(t, router, "user")
	bookingPayload := []byte(fmt.Sprintf(`{"slotId":"%s"}`, slotID))
	req, _ = http.NewRequest("POST", "/bookings/create", bytes.NewBuffer(bookingPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	resBooking := executeRequest(req, router)

	var bookingResp map[string]map[string]interface{}
	json.Unmarshal(resBooking.Body.Bytes(), &bookingResp)
	bookingID := bookingResp["booking"]["id"].(string)

	cancelPath := fmt.Sprintf("/bookings/%s/cancel", bookingID)
	req, _ = http.NewRequest("POST", cancelPath, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resCancel := executeRequest(req, router)

	if resCancel.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for Cancel, got %d. Body: %s", resCancel.Code, resCancel.Body.String())
	}

	var cancelResp map[string]map[string]interface{}
	json.Unmarshal(resCancel.Body.Bytes(), &cancelResp)
	status := cancelResp["booking"]["status"].(string)

	if status != "cancelled" {
		t.Errorf("Expected status to be 'cancelled', got '%s'", status)
	}

	req, _ = http.NewRequest("POST", cancelPath, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resCancelDouble := executeRequest(req, router)

	if resCancelDouble.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for repeated Cancel, got %d", resCancelDouble.Code)
	}
}
