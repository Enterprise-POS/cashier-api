package controller

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type WebhookControllerImpl struct {
}

func NewWebhookControllerImpl() WebhookController {
	return &WebhookControllerImpl{}
}

// HandlePaymentGateWayWebhook implements [WebhookController].
func (controller *WebhookControllerImpl) HandlePaymentGateWayWebhook(ctx *fiber.Ctx) error {
	// MidtransQRISNotification represents the POST body Midtrans sends to your
	// webhook endpoint for QRIS payment notifications.
	// Reference: https://docs.midtrans.com/reference/qris (POST Body Definition Table - QRIS)
	type MidtransQRISNotification struct {
		TransactionType          string `json:"transaction_type"`   // "on-us" or "off-us"
		TransactionTime          string `json:"transaction_time"`   // format: YYYY-MM-DD HH:MM:SS
		TransactionStatus        string `json:"transaction_status"` // see Midtrans Transaction Status docs
		TransactionID            string `json:"transaction_id"`
		StatusMessage            string `json:"status_message"`
		StatusCode               string `json:"status_code"`
		SignatureKey             string `json:"signature_key"` // used to verify notification authenticity
		SettlementTime           string `json:"settlement_time"`
		PaymentType              string `json:"payment_type"`
		OrderID                  string `json:"order_id"`
		MerchantID               string `json:"merchant_id"`
		MerchantCrossReferenceID string `json:"merchant_cross_reference_id"` // QRIS code identifier
		Issuer                   string `json:"issuer"`                      // provider paying over the QR
		GrossAmount              string `json:"gross_amount"`
		FraudStatus              string `json:"fraud_status"`
		Currency                 string `json:"currency"`
		Acquirer                 string `json:"acquirer"` // "airpay shopee", "gopay", etc.
	}

	var notification MidtransQRISNotification

	if err := ctx.BodyParser(&notification); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid webhook payload",
		})
	}

	fmt.Println(notification)

	// TODO: verify notification.SignatureKey before trusting anything else in the payload
	// TODO: map notification.TransactionStatus to your model.PaymentStatus and call SetPaymentStatus

	return ctx.SendStatus(fiber.StatusOK)
}
