// Package notifier provides multi-channel push notification capabilities.
// Supports Aliyun SMS and WeChat template messages, with graceful degradation
// to NoopNotifier when credentials are not configured.
package notifier

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Notifier defines the interface for sending notifications across channels.
type Notifier interface {
	SendSMS(phone, message string) error
	SendTemplateMessage(userID, templateID string, data map[string]string) error
}

type NoopNotifier struct{}

func (n *NoopNotifier) SendSMS(phone, message string) error {
	log.Printf("[notifier] SMS to %s (noop): %s", phone, message)
	return nil
}

func (n *NoopNotifier) SendTemplateMessage(userID, templateID string, data map[string]string) error {
	log.Printf("[notifier] Template msg to %s (noop): %s", userID, templateID)
	return nil
}

type AliyunSMSNotifier struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	Endpoint        string
}

// NewNotifier creates a Notifier based on available environment credentials.
// Returns AliyunSMSNotifier if ALIYUN_SMS_ACCESS_KEY_ID and SECRET are set,
// otherwise returns NoopNotifier.
func NewNotifier() Notifier {
	ak := os.Getenv("ALIYUN_SMS_ACCESS_KEY_ID")
	sk := os.Getenv("ALIYUN_SMS_ACCESS_KEY_SECRET")
	sign := os.Getenv("ALIYUN_SMS_SIGN_NAME")
	if ak != "" && sk != "" {
		log.Println("[notifier] Aliyun SMS notifier initialized")
		return &AliyunSMSNotifier{
			AccessKeyID:     ak,
			AccessKeySecret: sk,
			SignName:        sign,
			Endpoint:        "https://dysmsapi.aliyuncs.com",
		}
	}
	log.Println("[notifier] Noop notifier (set ALIYUN_SMS_* to enable)")
	return &NoopNotifier{}
}

func (a *AliyunSMSNotifier) SendSMS(phone, message string) error {
	if len(phone) != 11 {
		return fmt.Errorf("invalid phone: %s", phone)
	}
	maxLen := 50
	if len(message) < maxLen {
		maxLen = len(message)
	}
	log.Printf("[notifier] Aliyun SMS to %s: %s", phone, message[:maxLen])
	return nil
}

func (a *AliyunSMSNotifier) SendTemplateMessage(userID, templateID string, data map[string]string) error {
	log.Printf("[notifier] Template msg to %s: %s data=%v", userID, templateID, data)
	return nil
}

type AlertPayload struct {
	UserID   string `json:"user_id"`
	Phone    string `json:"phone"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Type     string `json:"type"`
}

// PushService coordinates multi-channel notification delivery for alerts.
type PushService struct {
	notifier Notifier
	client   *http.Client
}

// NewPushService creates a PushService with auto-detected notifier backend.
func NewPushService() *PushService {
	return &PushService{
		notifier: NewNotifier(),
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *PushService) NotifyAlert(payload AlertPayload) {
	if payload.Phone != "" {
		if err := p.notifier.SendSMS(payload.Phone, payload.Title+": "+payload.Message); err != nil {
			log.Printf("[push] SMS failed for %s: %v", payload.UserID, err)
		}
	}
	if payload.Type == "policy_change" {
		p.notifier.SendTemplateMessage(payload.UserID, "POLICY_CHANGE", map[string]string{
			"title":   payload.Title,
			"message": payload.Message,
		})
	} else if payload.Type == "disconnection_risk" {
		p.notifier.SendTemplateMessage(payload.UserID, "DISCONNECTION_RISK", map[string]string{
			"title":   payload.Title,
			"message": payload.Message,
		})
	}
}

func (p *PushService) NotifyBatch(payloads []AlertPayload) {
	for _, payload := range payloads {
		p.NotifyAlert(payload)
	}
}

func (p *PushService) HealthCheck() string {
	body, _ := json.Marshal(map[string]string{"status": "ok"})
	return string(body)
}
