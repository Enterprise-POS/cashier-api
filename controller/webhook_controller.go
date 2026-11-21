package controller

import "github.com/gofiber/fiber/v3"

type WebhookController interface {
	/*
		Handle payment gate webhook
	*/
	HandlePaymentGateWayWebhook(ctx fiber.Ctx) error
}
