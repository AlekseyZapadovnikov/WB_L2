package repository

import (
	"calendar/internal/domain"
	"reflect"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond) // для стабильности сравнения
	tests := []struct {
		name   string
		event  domain.Event
		wantID bool // ожидается ли генерация ID
	}{
		{
			name: "with explicit ID",
			event: domain.Event{
				UserID: 1,
				ID:     "explicit_id",
				Title:  "Event 1",
				Date:   now,
			},
			wantID: false,
		},
		{
			name: "with empty ID",
			event: domain.Event{
				UserID: 2,
				ID:     "",
				Title:  "Event 2",
				Date:   now.Add(24 * time.Hour),
			},
			wantID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewLocalStorage()
			id, err := repo.Create(tt.event)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if id == "" {
				t.Fatal("Create() returned empty ID")
			}

			stored, ok := repo.events[id]
			if !ok {
				t.Fatal("event not stored")
			}

			if stored.ID != id {
				t.Errorf("stored.ID = %q, want %q", stored.ID, id)
			}

			if tt.wantID && id == tt.event.ID {
				t.Errorf("expected generated ID, got original empty ID")
			}

			expected := tt.event
			if expected.ID == "" {
				expected.ID = id
			}
			if !reflect.DeepEqual(stored, expected) {
				t.Errorf("stored event = %+v, want %+v", stored, expected)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	repo := NewLocalStorage()
	event := domain.Event{
		UserID: 1,
		Title:  "Original",
		Date:   now,
	}
	id, err := repo.Create(event)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tests := []struct {
		name    string
		event   domain.Event
		wantErr bool
	}{
		{
			name: "valid update",
			event: domain.Event{
				ID:     id,
				UserID: 1,
				Title:  "Updated",
				Date:   now.Add(1 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "update non-existent",
			event: domain.Event{
				ID:     "nonexistent",
				UserID: 2,
				Title:  "Fake",
				Date:   now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr = %t", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			stored, ok := repo.events[tt.event.ID]
			if !ok {
				t.Fatal("updated event not found")
			}
			if !reflect.DeepEqual(stored, tt.event) {
				t.Errorf("stored = %+v, want %+v", stored, tt.event)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	repo := NewLocalStorage()
	event := domain.Event{UserID: 1, Title: "To delete", Date: time.Now()}
	id, _ := repo.Create(event)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "existing", id: id, wantErr: false},
		{name: "non-existent", id: "nonexistent", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := NewLocalStorage()
			if tt.id == id {
				testRepo.Create(event)
			}

			err := testRepo.Delete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr = %t", err, tt.wantErr)
			}
			if !tt.wantErr {
				if _, exists := testRepo.events[tt.id]; exists {
					t.Errorf("event still exists after deletion")
				}
			}
		})
	}
}

func TestGetByUserAndRange(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	repo := NewLocalStorage()

	events := []domain.Event{
		{UserID: 1, ID: "ev1", Title: "Before range", Date: baseTime.Add(-1 * time.Hour)},
		{UserID: 1, ID: "ev2", Title: "Start boundary", Date: baseTime},
		{UserID: 1, ID: "ev3", Title: "Inside", Date: baseTime.Add(30 * time.Minute)},
		{UserID: 1, ID: "ev4", Title: "End boundary", Date: baseTime.Add(1 * time.Hour)},
		{UserID: 1, ID: "ev5", Title: "After range", Date: baseTime.Add(2 * time.Hour)},
		{UserID: 2, ID: "ev6", Title: "Other user inside", Date: baseTime.Add(15 * time.Minute)},
	}
	var ids []string
	for _, e := range events {
		id, _ := repo.Create(e)
		ids = append(ids, id)
	}

	tests := []struct {
		name    string
		userID  int
		start   time.Time
		end     time.Time
		wantIDs map[string]bool
	}{
		{
			name:    "user 1 in [baseTime, baseTime+1h)",
			userID:  1,
			start:   baseTime,
			end:     baseTime.Add(1 * time.Hour),
			wantIDs: map[string]bool{ids[1]: true, ids[2]: true}, // Start boundary, Inside
		},
		{
			name:    "user 2 in [baseTime, baseTime+1h)",
			userID:  2,
			start:   baseTime,
			end:     baseTime.Add(1 * time.Hour),
			wantIDs: map[string]bool{ids[5]: true}, // Other user inside
		},
		{
			name:    "no events",
			userID:  3,
			start:   baseTime,
			end:     baseTime.Add(1 * time.Hour),
			wantIDs: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByUserAndRange(tt.userID, tt.start, tt.end)
			if err != nil {
				t.Fatalf("GetByUserAndRange() error = %v", err)
			}
			if len(got) != len(tt.wantIDs) {
				t.Errorf("got %d events, want %d", len(got), len(tt.wantIDs))
			}
			for _, e := range got {
				if !tt.wantIDs[e.ID] {
					t.Errorf("unexpected event ID: %s", e.ID)
				}
			}
		})
	}
}
