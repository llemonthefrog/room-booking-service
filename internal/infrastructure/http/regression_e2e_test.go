package http

import (
	"avito-task/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func requestJSON(t *testing.T, router http.Handler, method, path, token, body string, status int, result any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res := executeRequest(req, router)
	if res.Code != status {
		t.Fatalf("%s %s: got %d, want %d: %s", method, path, res.Code, status, res.Body)
	}
	if result != nil {
		if err := json.Unmarshal(res.Body.Bytes(), result); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2E_ConcurrencyAndFullyBookedDay(t *testing.T) {
	router, db := setupApp(t)
	admin := getDummyToken(t, router, "admin")
	user := getDummyToken(t, router, "user")
	getDummyToken(t, router, "user")
	var room struct {
		Room domain.Room `json:"room"`
	}
	requestJSON(t, router, "POST", "/rooms/create", admin, `{"name":"Concurrent Room"}`, 201, &room)
	roomPath := "/rooms/" + room.Room.Id.String()
	schedule := `{"daysOfWeek":[1,2,3,4,5,6,7],"startTime":"09:00","endTime":"09:30"}`

	// Start concurrent requests together; assertions stay in the test goroutine.
	parallelRequests := func(method, path, body, token string) []*httptest.ResponseRecorder {
		const count = 8
		responses := make([]*httptest.ResponseRecorder, count)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := range responses {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				req := httptest.NewRequest(method, path, strings.NewReader(body))
				req.Header.Set("Authorization", "Bearer "+token)
				responses[i] = executeRequest(req, router)
			}()
		}
		close(start)
		wg.Wait()
		return responses
	}

	created := 0
	for _, res := range parallelRequests("POST", roomPath+"/schedule/create", schedule, admin) {
		switch res.Code {
		case 201:
			created++
		case 409:
		default:
			t.Fatalf("schedule status %d: %s", res.Code, res.Body)
		}
	}
	if created != 1 {
		t.Fatalf("created %d schedules, want 1", created)
	}

	path := roomPath + "/slots/list?date=" + time.Now().UTC().AddDate(0, 0, 1).Format(time.DateOnly)
	var slotID string
	for _, res := range parallelRequests("GET", path, "", user) {
		if res.Code != 200 {
			t.Fatalf("slots status %d: %s", res.Code, res.Body)
		}
		var list SlotsListResponse
		if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil {
			t.Fatal(err)
		}
		if len(list.Slots) != 1 {
			t.Fatalf("got %d slots, want 1", len(list.Slots))
		}
		if slotID != "" && slotID != list.Slots[0].ID {
			t.Fatal("slot ID changed between requests")
		}
		slotID = list.Slots[0].ID
	}

	created = 0
	var booking BookingResponse
	payload := fmt.Sprintf(`{"slotId":%q}`, slotID)
	for _, res := range parallelRequests("POST", "/bookings/create", payload, user) {
		switch res.Code {
		case 201:
			created++
			if err := json.Unmarshal(res.Body.Bytes(), &booking); err != nil {
				t.Fatal(err)
			}
		case 409:
		default:
			t.Fatalf("booking status %d: %s", res.Code, res.Body)
		}
	}
	if created != 1 {
		t.Fatalf("created %d bookings, want 1", created)
	}
	// Verify the database constraint itself, independently of the slot lock.
	_, err := db.Exec(`INSERT INTO bookings (id, user_id, slot_id)
		VALUES ($1, $2, $3)`, uuid.New(), booking.Booking.UserID, slotID)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("duplicate active booking: want unique violation, got %v", err)
	}
	var slots SlotsListResponse
	requestJSON(t, router, "GET", path, user, "", 200, &slots)
	if len(slots.Slots) != 0 {
		t.Fatal("booked slot returned as available")
	}

	var anotherUser TokenResponse
	registration := `{"email":"other@example.com","password":"password","role":"user"}`
	requestJSON(t, router, "POST", "/register", "", registration, 201, &anotherUser)
	requestJSON(t, router, "POST", "/register", "", registration, 400, nil)
	cancelPath := "/bookings/" + booking.Booking.ID + "/cancel"
	requestJSON(t, router, "POST", cancelPath, anotherUser.Token, "", 403, nil)
	requestJSON(t, router, "POST", cancelPath, user, "", 200, nil)
	requestJSON(t, router, "POST", cancelPath, anotherUser.Token, "", 403, nil)
	requestJSON(t, router, "POST", cancelPath, user, "", 200, nil)
	requestJSON(t, router, "GET", path, user, "", 200, &slots)
	if len(slots.Slots) != 1 || slots.Slots[0].ID != slotID {
		t.Fatal("cancelled slot not available again")
	}
	requestJSON(t, router, "POST", "/bookings/create", user, payload, 201, nil)

	seed, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "seed.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := db.Exec(string(seed)); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}
