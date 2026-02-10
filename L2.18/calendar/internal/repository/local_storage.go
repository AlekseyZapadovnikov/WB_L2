package repository

import (
	"calendar/internal/domain"
	"fmt"
	"sync"
	"time"
)

type EventRepository interface {
	Create(e domain.Event) (string, error)
	Update(e domain.Event) error
	Delete(id string) error
	GetByUserAndRange(userID int, from, to time.Time) ([]domain.Event, error)
}

type localStorage struct {
	mu       sync.RWMutex
	events   map[string]domain.Event
	nextID   int64
}

func NewLocalStorage() *localStorage {
	return &localStorage{
		events: make(map[string]domain.Event),
	}
}

func (s *localStorage) Create(e domain.Event) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e.ID == "" {
		s.nextID++
		e.ID = fmt.Sprintf("event_%d", s.nextID)
	}
	s.events[e.ID] = e
	return e.ID, nil
}

func (s *localStorage) Update(e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; !exists {
		return domain.ErrEventNotFound
	}
	s.events[e.ID] = e
	return nil
}

func (s *localStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return domain.ErrEventNotFound
	}
	delete(s.events, id)
	return nil
}

func (s *localStorage) GetByUserAndRange(userID int, start, end time.Time) ([]domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Event
	for _, event := range s.events {
		if event.UserID == userID {
			if (event.Date.Equal(start) || event.Date.After(start)) && event.Date.Before(end) {
				result = append(result, event)
			}
		}
	}
	return result, nil
}

