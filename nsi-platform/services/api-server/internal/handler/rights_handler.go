package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type RightsRepository interface {
	GetPaymentRecords(ctx context.Context, userID string) ([]models.PaymentRecord, error)
	GetAlerts(ctx context.Context, userID string, unreadOnly bool) ([]models.Alert, error)
	MarkAlertRead(ctx context.Context, alertID, userID string) error
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetDueSoonPayments(ctx context.Context, withinDays int) ([]models.PaymentRecord, error)
	UpsertPaymentRecord(ctx context.Context, record *models.PaymentRecord) error
}

type paymentSummary struct {
	TotalRecords  int                  `json:"total_records"`
	PaidCount     int                  `json:"paid_count"`
	PendingCount  int                  `json:"pending_count"`
	MissedCount   int                  `json:"missed_count"`
	RecentRecords []models.PaymentRecord `json:"recent_records"`
}

func PaymentStatusHandler(repo RightsRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		records, err := repo.GetPaymentRecords(r.Context(), userID)
		if err != nil {
			respondError(w, err)
			return
		}

		summary := paymentSummary{
			TotalRecords:  len(records),
			RecentRecords: records,
		}

		for _, rec := range records {
			switch rec.Status {
			case "paid":
				summary.PaidCount++
			case "pending":
				summary.PendingCount++
			case "missed":
				summary.MissedCount++
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": summary})
	})
}

func AlertListHandler(repo RightsRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		unreadOnly := r.URL.Query().Get("unread_only") == "true"

		alerts, err := repo.GetAlerts(r.Context(), userID, unreadOnly)
		if err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": alerts})
	})
}

func MarkAlertReadHandler(repo RightsRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		alertID := r.URL.Query().Get("alert_id")
		if alertID == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "alert_id is required"})
			return
		}

		if err := repo.MarkAlertRead(r.Context(), alertID, userID); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": map[string]string{"status": "ok"}})
	})
}

type submitPaymentInput struct {
	PolicyType string  `json:"policy_type"`
	Month      string  `json:"month"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	DueDate    string  `json:"due_date"`
}

func SubmitPaymentRecordHandler(repo RightsRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var input submitPaymentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}

		validPolicyTypes := map[string]bool{"pension": true, "medical": true, "unemployment": true, "injury": true, "maternity": true}
		if !validPolicyTypes[input.PolicyType] {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "policy_type must be one of: pension, medical, unemployment, injury, maternity"})
			return
		}

		if _, err := time.Parse("2006-01", input.Month); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "month must be in YYYY-MM format"})
			return
		}

		if input.Amount <= 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "amount must be positive"})
			return
		}

		validStatus := map[string]bool{"paid": true, "pending": true, "missed": true}
		if !validStatus[input.Status] {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "status must be one of: paid, pending, missed"})
			return
		}

		if _, err := time.Parse("2006-01-02", input.DueDate); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "due_date must be a valid date in YYYY-MM-DD format"})
			return
		}

		record := &models.PaymentRecord{
			RecordID:   fmt.Sprintf("rec-%d", time.Now().UnixNano()),
			UserID:     userID,
			PolicyType: input.PolicyType,
			Month:      input.Month,
			Amount:     input.Amount,
			Status:     input.Status,
			DueDate:    input.DueDate,
		}

		if input.Status == "paid" {
			today := time.Now().Format("2006-01-02")
			record.PaidDate = &today
		}

		if err := repo.UpsertPaymentRecord(r.Context(), record); err != nil {
			respondError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": record})
	})
}

// AlertScheduler 断缴风险检测定时器
type AlertScheduler struct {
	repo        RightsRepository
	policyRepo  PolicyChangeChecker
	pusher      AlertPusher
	stopCh      chan struct{}
	ticker      *time.Ticker
	once        sync.Once
	lastCheck   time.Time
}

type AlertPusher interface {
	NotifyAlert(payload NotifierPayload)
}

type NotifierPayload struct {
	UserID  string `json:"user_id"`
	Phone   string `json:"phone"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func NewAlertScheduler(repo RightsRepository) *AlertScheduler {
	return &AlertScheduler{
		repo:      repo,
		stopCh:    make(chan struct{}),
		lastCheck: time.Now(),
	}
}

func (s *AlertScheduler) SetPolicyChecker(pc PolicyChangeChecker) {
	s.policyRepo = pc
}

func (s *AlertScheduler) SetPusher(p AlertPusher) {
	s.pusher = p
}

type PolicyChangeChecker interface {
	GetRecentlyUpdatedPolicies(ctx context.Context, since time.Time) ([]string, error)
	GetUsersWithPlans(ctx context.Context) (map[string][]string, error)
}

func (s *AlertScheduler) Start(interval time.Duration) {
	s.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkDisconnectionRisk()
				s.checkPolicyChanges()
			case <-s.stopCh:
				s.ticker.Stop()
				return
			}
		}
	}()
}

func (s *AlertScheduler) Stop() {
	s.once.Do(func() {
		close(s.stopCh)
	})
}

func (s *AlertScheduler) checkDisconnectionRisk() {
	ctx := context.Background()
	duePayments, err := s.repo.GetDueSoonPayments(ctx, 7)
	if err != nil {
		log.Printf("[alerts] failed to query due payments: %v", err)
		return
	}
	for _, p := range duePayments {
		alertID := fmt.Sprintf("DISCONN-%s-%s", p.UserID, p.RecordID)
		existing, err := s.repo.GetAlerts(ctx, p.UserID, false)
		if err == nil {
			alreadyExists := false
			for _, a := range existing {
				if a.AlertID == alertID {
					alreadyExists = true
					break
				}
			}
			if alreadyExists {
				continue
			}
		}
		alert := &models.Alert{
			AlertID:   alertID,
			UserID:    p.UserID,
			AlertType: "disconnection_risk",
			Severity:  "high",
			Title:     fmt.Sprintf("社保断缴风险 - %s", p.PolicyType),
			Message:   fmt.Sprintf("您有一笔 %s 费用（%.2f元）即将在 %s 到期，请及时缴费避免断缴影响社保待遇", p.PolicyType, p.Amount, p.DueDate),
		}
		if err := s.repo.CreateAlert(ctx, alert); err != nil {
			log.Printf("[alerts] failed to create alert for %s: %v", p.UserID, err)
		} else {
			log.Printf("[alerts] created disconnection risk alert for %s (%s)", p.UserID, p.PolicyType)
			if s.pusher != nil {
				s.pusher.NotifyAlert(NotifierPayload{
					UserID: p.UserID, Title: alert.Title, Message: alert.Message, Type: "disconnection_risk",
				})
			}
		}
	}
}

func PolicyChangeNotifyHandler(repo RightsRepository, serviceSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serviceSecret == "" || r.Header.Get("X-Service-Secret") != serviceSecret {
			respondJSON(w, 401, map[string]interface{}{"code": "UNAUTHORIZED", "message": "invalid service secret"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			PolicyID     string `json:"policy_id"`
			PolicyTitle  string `json:"policy_title"`
			ChangeType   string `json:"change_type"`
			NewVersion   int    `json:"new_version"`
			AffectedUsers []string `json:"affected_users"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "invalid JSON"})
			return
		}
		if req.PolicyID == "" || len(req.AffectedUsers) == 0 {
			respondJSON(w, 400, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "policy_id and affected_users required"})
			return
		}

		ctx := r.Context()
		title := req.PolicyTitle
		if title == "" {
			title = req.PolicyID
		}
		msg := fmt.Sprintf("您关注的政策「%s」已更新到第%d版，建议重新查看您的社会保方案", title, req.NewVersion)
		created := 0
		for _, userID := range req.AffectedUsers {
			alertID := fmt.Sprintf("POLCHG-%s-%s-v%d", userID, req.PolicyID, req.NewVersion)
			alert := &models.Alert{
				AlertID:   alertID,
				UserID:    userID,
				AlertType: "policy_change",
				Severity:  "medium",
				Title:     fmt.Sprintf("政策变更通知 - %s", title),
				Message:   msg,
			}
			policyID := req.PolicyID
			alert.PolicyID = &policyID
			if err := repo.CreateAlert(ctx, alert); err != nil {
				log.Printf("[alerts] failed to create policy_change alert for %s: %v", userID, err)
				continue
			}
			created++
		}
		log.Printf("[alerts] policy_change: notified %d/%d users for policy %s v%d", created, len(req.AffectedUsers), req.PolicyID, req.NewVersion)
	})
}

func (s *AlertScheduler) checkPolicyChanges() {
	if s.policyRepo == nil {
		return
	}
	ctx := context.Background()
	now := time.Now()
	since := s.lastCheck
	s.lastCheck = now

	policyIDs, err := s.policyRepo.GetRecentlyUpdatedPolicies(ctx, since)
	if err != nil || len(policyIDs) == 0 {
		return
	}

	userPolicies, err := s.policyRepo.GetUsersWithPlans(ctx)
	if err != nil {
		return
	}

	for _, pid := range policyIDs {
		for userID, pids := range userPolicies {
			for _, upid := range pids {
				if upid == pid {
					alertID := fmt.Sprintf("POLCHG-AUTO-%s-%s-%d", userID, pid, now.Unix())
					alert := &models.Alert{
						AlertID:   alertID,
						UserID:    userID,
						AlertType: "policy_change",
						Severity:  "medium",
						Title:     "政策变更通知",
						Message:   fmt.Sprintf("您关注的政策「%s」已更新，建议重新查看社保方案", pid),
					}
					policyIDCopy := pid
					alert.PolicyID = &policyIDCopy
					s.repo.CreateAlert(ctx, alert)
				}
			}
		}
	}
}


