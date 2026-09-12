package service

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"cashier-api/repository"

	"github.com/stretchr/testify/mock"
)

type PaymentProviderMock struct {
	Mock *mock.Mock
}

func NewPaymentProviderMock(m *mock.Mock) PaymentProvider {
	return &PaymentProviderMock{m}
}

// CreateTransaction implements PaymentProvider.
func (p *PaymentProviderMock) CreateTransaction(midClient *client.MidtransProvider, params *repository.CreateTransactionParams) (*model.PaymentProviderResponse, error) {
	args := p.Mock.Called(midClient, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.PaymentProviderResponse), args.Error(1)
}

// CheckTransaction implements PaymentProvider.
func (p *PaymentProviderMock) CheckTransaction(midClient *client.MidtransProvider, transactionId string) (model.PaymentStatusResponse, error) {
	args := p.Mock.Called(midClient, transactionId)

	if args.Get(0) == nil {
		return model.PaymentStatusResponse{}, args.Error(1)
	}

	return args.Get(0).(model.PaymentStatusResponse), args.Error(1)
}

// CancelTransaction implements PaymentProvider.
func (p *PaymentProviderMock) CancelTransaction(midClient *client.MidtransProvider, transactionId string) (model.PaymentStatusResponse, error) {
	args := p.Mock.Called(midClient, transactionId)

	if args.Get(0) == nil {
		return model.PaymentStatusResponse{}, args.Error(1)
	}

	return args.Get(0).(model.PaymentStatusResponse), args.Error(1)
}
