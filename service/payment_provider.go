package service

import (
	"cashier-api/model"
	"cashier-api/repository"
)

type PaymentProvider interface {
	CreateTransaction(params *repository.CreateTransactionParams) (*model.PaymentProviderResponse, error)
	CheckTransaction(transactionId string) (model.PaymentStatusResponse, error)
}
