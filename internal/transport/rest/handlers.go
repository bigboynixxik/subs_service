package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"subs_service/internal/models"
	"subs_service/internal/repository"
	"subs_service/internal/service"
	"subs_service/pkg/logger"
	"subs_service/pkg/response"
	"time"

	"github.com/google/uuid"
)

type SubsHandler struct {
	svc service.SubscriptionService
}

func NewSubsHandler(s service.SubscriptionService) *SubsHandler {
	return &SubsHandler{svc: s}
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
}

func mapToResponse(sub *models.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format("01-2006"), // Форматируем в Месяц-Год
	}
	if sub.EndDate != nil {
		endDateStr := sub.EndDate.Format("01-2006")
		resp.EndDate = &endDateStr
	}
	return resp
}

type SubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

// CreateSubscription создает новую запись о подписке
// @Summary Создать подписку
// @Description Добавляет новую онлайн подписку пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body SubscriptionRequest true "Данные подписки"
// @Success 201 {object} map[string]string "id созданной подписки"
// @Failure 400 {object} response.ErrorResponse "Некорректные данные запроса"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (h *SubsHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		l.Error("rest.CreateSubscription json decode",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		l.Error("rest.CreateSubscription start date format",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid start date")
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsedEndDate, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			l.Error("rest.CreateSubscription end date format",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid end date")
			return
		}
		endDate = &parsedEndDate
	}

	if req.ServiceName == "" {
		l.Error("rest.CreateSubscription service name is required")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid service name")
		return
	}

	if req.UserID == "" {
		l.Error("rest.CreateSubscription user_id is required")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user id")
		return
	}
	userIdParsed, err := uuid.Parse(req.UserID)
	if err != nil {
		l.Error("rest.CreateSubscription user_id parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user id")
		return
	}
	sub := models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userIdParsed,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	id, err := h.svc.Create(r.Context(), &sub)
	if err != nil {
		l.Error("rest.CreateSubscription create subscription error",
			slog.String("error", err.Error()))
		switch {
		case errors.Is(err, service.ErrInvalidName),
			errors.Is(err, service.ErrInvalidPrice),
			errors.Is(err, service.ErrInvalidDate):
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		default:

			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"id": id,
	})
}

// GetSubscription возвращает подписку по ID
// @Summary Получить подписку
// @Description Возвращает данные подписки по ее UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} SubscriptionResponse "Данные подписки"
// @Failure 400 {object} response.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} response.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [get]
func (h *SubsHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		l.Error("rest.GetSubscription id parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid subscription ID")
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		l.Error("rest.GetSubscription", slog.String("error", err.Error()))
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Subscription not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, mapToResponse(sub))
}

// UpdateSubscription обновляет существующую подписку
// @Summary Обновить подписку
// @Description Обновляет данные существующей подписки по ее ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Param request body SubscriptionRequest true "Новые данные подписки"
// @Success 200 {object} map[string]string "Статус обновления"
// @Failure 400 {object} response.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} response.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [put]
func (h *SubsHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		l.Error("rest.UpdateSubscription id parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid subscription ID")
		return
	}

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		l.Error("rest.UpdateSubscription json decode",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		l.Error("rest.UpdateSubscription start date format",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid start date")
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsedEndDate, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			l.Error("rest.UpdateSubscription end date format",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid end date")
			return
		}
		endDate = &parsedEndDate
	}

	userIdParsed, err := uuid.Parse(req.UserID)
	if err != nil {
		l.Error("rest.UpdateSubscription user_id parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user id")
		return
	}

	sub := models.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userIdParsed,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.svc.Update(r.Context(), &sub); err != nil {
		l.Error("rest.UpdateSubscription", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, repository.ErrSubscriptionNotFound):
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Subscription not found")
		case errors.Is(err, service.ErrInvalidName), errors.Is(err, service.ErrInvalidPrice), errors.Is(err, service.ErrInvalidDate):
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteSubscription удаляет подписку по ID
// @Summary Удалить подписку
// @Description Удаляет запись о подписке из базы данных
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} map[string]string "Статус удаления"
// @Failure 400 {object} response.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} response.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [delete]
func (h *SubsHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		l.Error("rest.DeleteSubscription id parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid subscription ID")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		l.Error("rest.DeleteSubscription", slog.String("error", err.Error()))
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Subscription not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListSubscriptions возвращает список всех подписок
// @Summary Список всех подписок
// @Description Получает массив всех подписок в системе
// @Tags subscriptions
// @Produce json
// @Success 200 {array} SubscriptionResponse "Список подписок"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions [get]
func (h *SubsHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	subs, err := h.svc.List(r.Context())
	if err != nil {
		l.Error("rest.ListSubscriptions", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		resp = append(resp, mapToResponse(&sub))
	}

	response.JSON(w, http.StatusOK, resp)
}

// CalculateTotalCost подсчитывает суммарную стоимость подписок
// @Summary Подсчет стоимости подписок
// @Description Считает суммарную стоимость подписок за выбранный период для пользователя с возможностью фильтрации по сервису
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса (опционально)"
// @Param from query string true "Начало периода (ММ-ГГГГ)"
// @Param to query string true "Конец периода (ММ-ГГГГ)"
// @Success 200 {object} map[string]int "Суммарная стоимость"
// @Failure 400 {object} response.ErrorResponse "Некорректные параметры запроса"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/cost [get]
func (h *SubsHandler) CalculateTotalCost(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userIDStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if userIDStr == "" || fromStr == "" || toStr == "" {
		l.Error("rest.CalculateTotalCost user_id, from, and to are required")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id, from, and to are required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		l.Error("rest.CalculateTotalCost user_id parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user_id format")
		return
	}

	from, err := time.Parse("01-2006", fromStr)
	if err != nil {
		l.Error("rest.CalculateTotalCost from string parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid from date, expected MM-YYYY")
		return
	}

	to, err := time.Parse("01-2006", toStr)
	if err != nil {
		l.Error("rest.CalculateTotalCost to string parse error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid to date, expected MM-YYYY")
		return
	}

	cost, err := h.svc.CalculateTotalCost(r.Context(), userID, serviceName, from, to)
	if err != nil {
		l.Error("rest.CalculateTotalCost", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]int{
		"total_cost": cost,
	})
}
