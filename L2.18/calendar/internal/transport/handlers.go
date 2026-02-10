package transport

import (
	"calendar/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type EventUseCase interface {
	CreateEvent(userID int, dateStr, title string) (string, error)
	UpdateEvent(id string, userID int, dateStr, title string) error
	DeleteEvent(id string) error
	GetEventsForDay(userID int, dateStr string) ([]domain.Event, error)
	GetEventsForWeek(userID int, dateStr string) ([]domain.Event, error)
	GetEventsForMonth(userID int, dateStr string) ([]domain.Event, error)
}

type Handler struct {
	uc EventUseCase
}

func NewHandler(uc EventUseCase) *Handler {
	return &Handler{uc: uc}
}



type response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type createRequest struct {
	UserID int    `json:"user_id"`
	Date   string `json:"date"`
	Event  string `json:"event"`
}

type updateRequest struct {
	ID     string `json:"id"`
	UserID int    `json:"user_id"`
	Date   string `json:"date"`
	Event  string `json:"event"`
}

type deleteRequest struct {
	ID string `json:"id"`
}

// --- Handlers ---

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := decodeBody(r, &req); err != nil {
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	id, err := h.uc.CreateEvent(req.UserID, req.Date, req.Event)
	if err != nil {
		h.handleLogicError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if err := decodeBody(r, &req); err != nil {
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.uc.UpdateEvent(req.ID, req.UserID, req.Date, req.Event); err != nil {
		h.handleLogicError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"result": "updated"})
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	var req deleteRequest
	if err := decodeBody(r, &req); err != nil {
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.uc.DeleteEvent(req.ID); err != nil {
		h.handleLogicError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"result": "deleted"})
}

func (h *Handler) EventsForDay(w http.ResponseWriter, r *http.Request) {
	userID, date, err := parseQueryParams(r)
	if err != nil {
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	events, err := h.uc.GetEventsForDay(userID, date)
	if err != nil {
		h.handleLogicError(w, err)
		return
	}
	h.sendJSON(w, http.StatusOK, events)
}

func (h *Handler) EventsForWeek(w http.ResponseWriter, r *http.Request) {
	userID, date, err := parseQueryParams(r)
	if err != nil {
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	events, err := h.uc.GetEventsForWeek(userID, date)
	if err != nil {
		h.handleLogicError(w, err)
		return
	}
	h.sendJSON(w, http.StatusOK, events)
}

func (h *Handler) EventsForMonth(w http.ResponseWriter, r *http.Request) {
	userID, date, err := parseQueryParams(r)
	if err != nil {
		h.sendError(w, err, http.StatusBadRequest)
		return
	}

	events, err := h.uc.GetEventsForMonth(userID, date)
	if err != nil {
		h.handleLogicError(w, err)
		return
	}
	h.sendJSON(w, http.StatusOK, events)
}

// --- Helpers ---

func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	// Для поддержки x-www-form-urlencoded можно добавить проверку Content-Type
	// Здесь реализация для JSON как основного формата в Best Practices
	return json.NewDecoder(r.Body).Decode(v)
}

func parseQueryParams(r *http.Request) (int, string, error) {
	q := r.URL.Query()
	userIDStr := q.Get("user_id")
	date := q.Get("date")

	if userIDStr == "" || date == "" {
		return 0, "", errors.New("missing user_id or date")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, "", errors.New("invalid user_id")
	}
	return userID, date, nil
}

func (h *Handler) handleLogicError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEventNotFound):
		h.sendError(w, err, http.StatusServiceUnavailable) // ТЗ: 503
	case errors.Is(err, domain.ErrDateInvalid):
		h.sendError(w, err, http.StatusBadRequest) // ТЗ: 400
	default:
		h.sendError(w, err, http.StatusInternalServerError) // ТЗ: 500
	}
}

func (h *Handler) sendJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// Оборачиваем результат в ключ result согласно ТЗ
	json.NewEncoder(w).Encode(response{Result: payload})
}

func (h *Handler) sendError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response{Error: err.Error()})
}