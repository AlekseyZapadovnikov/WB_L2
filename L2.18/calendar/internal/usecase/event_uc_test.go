// internal/usecase/event_usecase_test.go
package usecase

import (
	"errors"
	"testing"
	"time"

	"calendar/internal/domain"
	repoMocks "calendar/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventUseCase_CreateEvent_InvalidDate(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	id, err := uc.CreateEvent(1, "not-a-date", "title")

	require.ErrorIs(t, err, domain.ErrDateInvalid)
	require.Empty(t, id)

	// If date invalid, repo.Create must not be called; AssertExpectations is handled by mock cleanup.
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestEventUseCase_CreateEvent_OK(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	userID := 42
	dateStr := "2026-02-09"
	title := "Meet"
	wantID := "evt-1"

	wantDate, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)

	repo.EXPECT().
		Create(domain.Event{
			UserID: userID,
			Title:  title,
			Date:   wantDate,
		}).
		Return(wantID, nil).
		Once()

	id, err := uc.CreateEvent(userID, dateStr, title)

	require.NoError(t, err)
	require.Equal(t, wantID, id)
}

func TestEventUseCase_CreateEvent_RepoError(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	userID := 1
	dateStr := "2026-02-09"
	title := "t"
	wantErr := errors.New("db down")

	wantDate, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)

	repo.EXPECT().
		Create(domain.Event{UserID: userID, Title: title, Date: wantDate}).
		Return("", wantErr).
		Once()

	id, err := uc.CreateEvent(userID, dateStr, title)

	require.ErrorIs(t, err, wantErr)
	require.Empty(t, id)
}

func TestEventUseCase_UpdateEvent_InvalidDate(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	err := uc.UpdateEvent("id1", 1, "bad-date", "title")

	require.ErrorIs(t, err, domain.ErrDateInvalid)
	repo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestEventUseCase_UpdateEvent_OK(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	id := "evt-42"
	userID := 7
	dateStr := "2026-02-09"
	title := "Updated title"

	wantDate, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)

	repo.EXPECT().
		Update(domain.Event{
			ID:     id,
			UserID: userID,
			Title:  title,
			Date:   wantDate,
		}).
		Return(nil).
		Once()

	err = uc.UpdateEvent(id, userID, dateStr, title)
	require.NoError(t, err)
}

func TestEventUseCase_UpdateEvent_RepoError(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	id := "evt-42"
	userID := 7
	dateStr := "2026-02-09"
	title := "Updated title"
	wantErr := errors.New("update failed")

	wantDate, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)

	repo.EXPECT().
		Update(domain.Event{ID: id, UserID: userID, Title: title, Date: wantDate}).
		Return(wantErr).
		Once()

	err = uc.UpdateEvent(id, userID, dateStr, title)
	require.ErrorIs(t, err, wantErr)
}

func TestEventUseCase_DeleteEvent_OK(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	repo.EXPECT().Delete("evt-1").Return(nil).Once()

	err := uc.DeleteEvent("evt-1")
	require.NoError(t, err)
}

func TestEventUseCase_DeleteEvent_RepoError(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	wantErr := errors.New("delete failed")
	repo.EXPECT().Delete("evt-1").Return(wantErr).Once()

	err := uc.DeleteEvent("evt-1")
	require.ErrorIs(t, err, wantErr)
}

func TestEventUseCase_GetEventsForDay_InvalidDate(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	events, err := uc.GetEventsForDay(1, "bad-date")

	require.ErrorIs(t, err, domain.ErrDateInvalid)
	require.Nil(t, events)
	repo.AssertNotCalled(t, "GetByUserAndRange", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventUseCase_GetEventsForDay_OK(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	userID := 10
	dateStr := "2026-02-09"

	from, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)
	to := from.AddDate(0, 0, 1)

	want := []domain.Event{
		{ID: "1", UserID: userID, Title: "A", Date: from},
	}

	repo.EXPECT().
		GetByUserAndRange(userID, from, to).
		Return(want, nil).
		Once()

	got, err := uc.GetEventsForDay(userID, dateStr)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEventUseCase_GetEventsForWeek_InvalidDate(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	events, err := uc.GetEventsForWeek(1, "bad-date")

	require.ErrorIs(t, err, domain.ErrDateInvalid)
	require.Nil(t, events)
	repo.AssertNotCalled(t, "GetByUserAndRange", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventUseCase_GetEventsForWeek_OK(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	userID := 10
	dateStr := "2026-02-09"

	from, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)
	to := from.AddDate(0, 0, 7)

	want := []domain.Event{
		{ID: "1", UserID: userID, Title: "W", Date: from},
	}

	repo.EXPECT().
		GetByUserAndRange(userID, from, to).
		Return(want, nil).
		Once()

	got, err := uc.GetEventsForWeek(userID, dateStr)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEventUseCase_GetEventsForMonth_InvalidDate(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	events, err := uc.GetEventsForMonth(1, "bad-date")

	require.ErrorIs(t, err, domain.ErrDateInvalid)
	require.Nil(t, events)
	repo.AssertNotCalled(t, "GetByUserAndRange", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventUseCase_GetEventsForMonth_OK(t *testing.T) {
	repo := repoMocks.NewMockEventRepository(t)
	uc := NewEventUseCase(repo)

	userID := 10
	dateStr := "2026-02-09"

	from, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)
	to := from.AddDate(0, 1, 0)

	want := []domain.Event{
		{ID: "1", UserID: userID, Title: "M", Date: from},
	}

	repo.EXPECT().
		GetByUserAndRange(userID, from, to).
		Return(want, nil).
		Once()

	got, err := uc.GetEventsForMonth(userID, dateStr)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
