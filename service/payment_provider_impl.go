package service

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"cashier-api/repository"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type PaymentProviderImpl struct {
	MidtransProvider client.MidtransProvider
}

// CancelTransaction implements [PaymentProvider].
func (p *PaymentProviderImpl) CancelTransaction(transactionId string) (model.PaymentStatusResponse, error) {
	return p.MidtransProvider.CancelTransaction(transactionId)
}

// CheckTransaction implements [PaymentProvider].
func (p *PaymentProviderImpl) CheckTransaction(transactionId string) (model.PaymentStatusResponse, error) {
	return p.MidtransProvider.CheckTransaction(transactionId)
}

// CreateTransaction implements [PaymentProvider].
func (p *PaymentProviderImpl) CreateTransaction(params *repository.CreateTransactionParams) (*model.PaymentProviderResponse, error) {
	// Initiate midtrans Snap Request
	var items []midtrans.ItemDetails = make([]midtrans.ItemDetails, 0)
	for _, item := range params.Items {
		itemId := strconv.Itoa(item.ItemId)
		items = append(items, midtrans.ItemDetails{
			ID:    itemId,
			Price: int64(item.StorePriceSnapshot),
			Qty:   int32(item.Quantity),
			Name:  item.ItemNameSnapshot,
		})
	}
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  params.TransactionId,
			GrossAmt: int64(params.TotalAmount),
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},

		Items: &items,
	}

	response, err := p.MidtransProvider.CreateTransaction(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func NewPaymentProviderImpl() PaymentProvider {
	return &PaymentProviderImpl{
		MidtransProvider: *client.NewMidtransProvider(),
	}
}
