package usecase

import (
	"calendar/internal/domain"
	"calendar/internal/repository"
	"time"
)

type EventUseCase struct {
	repo repository.EventRepository
}

func NewEventUseCase(repo repository.EventRepository) *EventUseCase {
	return &EventUseCase{repo: repo}
}

func (uc *EventUseCase) CreateEvent(userID int, dateStr, title string) (string, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", domain.ErrDateInvalid
	}

	event := domain.Event{
		UserID: userID,
		Title:  title,
		Date:   date,
	}
	return uc.repo.Create(event)
}

func (uc *EventUseCase) UpdateEvent(id string, userID int, dateStr, title string) error {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return domain.ErrDateInvalid
	}

	event := domain.Event{
		ID:     id,
		UserID: userID,
		Title:  title,
		Date:   date,
	}
	return uc.repo.Update(event)
}

func (uc *EventUseCase) DeleteEvent(id string) error {
	return uc.repo.Delete(id)
}

func (uc *EventUseCase) GetEventsForDay(userID int, dateStr string) ([]domain.Event, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, domain.ErrDateInvalid
	}
	return uc.repo.GetByUserAndRange(userID, t, t.AddDate(0, 0, 1))
}

func (uc *EventUseCase) GetEventsForWeek(userID int, dateStr string) ([]domain.Event, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, domain.ErrDateInvalid
	}
	
	return uc.repo.GetByUserAndRange(userID, t, t.AddDate(0, 0, 7))
}

func (uc *EventUseCase) GetEventsForMonth(userID int, dateStr string) ([]domain.Event, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, domain.ErrDateInvalid
	}
	return uc.repo.GetByUserAndRange(userID, t, t.AddDate(0, 1, 0))
}