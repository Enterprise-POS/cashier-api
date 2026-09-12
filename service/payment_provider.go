package service

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"cashier-api/repository"
)

type PaymentProvider interface {
	CreateTransaction(midClient *client.MidtransProvider, params *repository.CreateTransactionParams) (*model.PaymentProviderResponse, error)
	CheckTransaction(midClient *client.MidtransProvider, transactionId string) (model.PaymentStatusResponse, error)
	CancelTransaction(midClient *client.MidtransProvider, transactionId string) (model.PaymentStatusResponse, error)
}
