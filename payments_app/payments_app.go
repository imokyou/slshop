package paymentsapp

import (
	"context"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Payments APP API Service
// =====================================================================

type Service interface {
	NotifyActivation(ctx context.Context, req ActivationNotification) error
	NotifyPaymentSuccess(ctx context.Context, req PaymentNotification) error
	NotifyRefundSuccess(ctx context.Context, req RefundNotification) error
	NotifyDeviceBinding(ctx context.Context, req DeviceBindingNotification) error
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type ActivationNotification struct {
	ExternalAccountID string `json:"external_account_id,omitempty"`
	Status            string `json:"status,omitempty"`
	Message           string `json:"message,omitempty"`
}

type PaymentNotification struct {
	ExternalPaymentID string `json:"external_payment_id,omitempty"`
	GID               string `json:"gid,omitempty"`
	Amount            string `json:"amount,omitempty"`
	Currency          string `json:"currency,omitempty"`
	Status            string `json:"status,omitempty"`
	Message           string `json:"message,omitempty"`
}

type RefundNotification struct {
	ExternalRefundID string `json:"external_refund_id,omitempty"`
	GID              string `json:"gid,omitempty"`
	Amount           string `json:"amount,omitempty"`
	Currency         string `json:"currency,omitempty"`
	Status           string `json:"status,omitempty"`
	Message          string `json:"message,omitempty"`
}

type DeviceBindingNotification struct {
	ExternalDeviceID string `json:"external_device_id,omitempty"`
	Status           string `json:"status,omitempty"`
	Message          string `json:"message,omitempty"`
}

// =====================================================================
// Implementation
// =====================================================================

// POST payments_apps/api/activate.json
func (s *serviceOp) NotifyActivation(ctx context.Context, req ActivationNotification) error {
	return s.client.Post(ctx, s.client.CreatePath("payments_apps/api/activate.json"), req, nil)
}

// POST payments_apps/api/payment_success.json
func (s *serviceOp) NotifyPaymentSuccess(ctx context.Context, req PaymentNotification) error {
	return s.client.Post(ctx, s.client.CreatePath("payments_apps/api/payment_success.json"), req, nil)
}

// POST payments_apps/api/refund_success.json
func (s *serviceOp) NotifyRefundSuccess(ctx context.Context, req RefundNotification) error {
	return s.client.Post(ctx, s.client.CreatePath("payments_apps/api/refund_success.json"), req, nil)
}

// POST payments_apps/api/device_binding.json
func (s *serviceOp) NotifyDeviceBinding(ctx context.Context, req DeviceBindingNotification) error {
	return s.client.Post(ctx, s.client.CreatePath("payments_apps/api/device_binding.json"), req, nil)
}
