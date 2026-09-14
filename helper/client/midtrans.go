package client

import (
	"cashier-api/model"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	log "github.com/sirupsen/logrus"
)

type MidtransProvider struct {
	snapClient snap.Client
	coreApi    coreapi.Client
}

// newMidtransProvider builds ONE immutable client bound to a single server key.
// Nothing on this struct is ever mutated after construction — that immutability
// is what guarantees tenant A's requests can never end up running against
// tenant B's key.
func NewMidtransProvider(serverKey string, env string) *MidtransProvider {
	midtransEnv := resolveMidtransEnvironment(env)

	var s snap.Client
	s.New(serverKey, midtransEnv)

	var c coreapi.Client
	c.New(serverKey, midtransEnv)

	return &MidtransProvider{snapClient: s, coreApi: c}
}

// resolveMidtransEnvironment maps our own app-level environment names to the
// midtrans-go SDK's environment type.
func resolveMidtransEnvironment(env string) midtrans.EnvironmentType {
	switch env {
	case "prod", "production":
		return midtrans.Production
	default:
		return midtrans.Sandbox
	}
}

func (m *MidtransProvider) CreateTransaction(req *snap.Request) (*model.PaymentProviderResponse, error) {
	resp, err := m.snapClient.CreateTransaction(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s (%d)", err.GetMessage(), err.GetStatusCode()))
	}

	response := model.PaymentProviderResponse{
		Token:         resp.Token,
		RedirectURL:   resp.RedirectURL,
		StatusCode:    resp.StatusCode,
		ErrorMessages: resp.ErrorMessages,
	}

	return &response, nil
}

func (m *MidtransProvider) CheckTransaction(orderId string) (model.PaymentStatusResponse, error) {
	// In midtrans transaction id called order id
	// In our database transaction id
	// transaction_id == orderId
	transactionStatusResp, err := m.coreApi.CheckTransaction(orderId)
	var paymentStatus model.PaymentStatus
	if err != nil {
		return model.PaymentStatusResponse{}, err
	} else {
		if transactionStatusResp != nil {
			// Do set transaction status based on response from check transaction status
			switch transactionStatusResp.TransactionStatus {
			case "capture":
				if transactionStatusResp.FraudStatus == "challenge" {
					// TODO set transaction status on your database to 'challenge'
					// e.g: 'Payment status challenged. Please take action on your Merchant Administration Portal
				} else if transactionStatusResp.FraudStatus == "accept" {
					// TODO set transaction status on your database to 'success'
				}
				break
			case "settlement":
				// Called settlement for midtrans
				// TODO set transaction status to 'settlement'
				paymentStatus = model.PaymentStatusSuccess
				break
			case "deny":
				// TODO you can ignore 'deny', because most of the time it allows payment retries
				// and later can become success
				paymentStatus = model.PaymentStatusPending
				break
			case "cancel", "expire":
				paymentStatus = model.PaymentStatusCancelled
				break
			case "pending":
				paymentStatus = model.PaymentStatusPending
				break
			default:
				log.Errorf("[FATAL ERROR] Unknown transaction status detected. Transaction status: %s. From transaction id: %s", transactionStatusResp.TransactionStatus, orderId)
				return model.PaymentStatusResponse{}, fmt.Errorf("[FATAL ERROR] Unknown transaction status detected. Transaction status: %s. From transaction id: %s", transactionStatusResp.TransactionStatus, orderId)
			}
		} else {
			log.Errorf("[FATAL ERROR] Something gone wrong for transaction id: %s", orderId)
			return model.PaymentStatusResponse{}, fmt.Errorf("[FATAL ERROR] Something gone wrong for transaction id: %s", orderId)
		}
	}

	log.Debugf("Transaction status: %s. For transaction id: %s", transactionStatusResp.TransactionStatus, orderId)

	return model.PaymentStatusResponse{PaymentStatus: paymentStatus, StatusCode: transactionStatusResp.StatusCode}, nil
}

func (m *MidtransProvider) CancelTransaction(orderId string) (model.PaymentStatusResponse, error) {

	cancelResponse, err := m.coreApi.CancelTransaction(orderId)
	var paymentStatus model.PaymentStatus
	if err != nil {
		// 412
		// Already deleted
		if err.StatusCode == http.StatusPreconditionFailed {
			code := strconv.Itoa(err.StatusCode)
			return model.PaymentStatusResponse{
				PaymentStatus: model.PaymentStatusCancelled,
				StatusCode:    code,
				Message:       "Already deleted",
			}, nil
		}
		return model.PaymentStatusResponse{}, err
	} else {
		if cancelResponse != nil {
			// Do set transaction status based on response from check transaction status
			switch cancelResponse.TransactionStatus {
			case "cancel":
				paymentStatus = model.PaymentStatusCancelled
				break
			default:
				log.Errorf("[FATAL ERROR] Failed to cancel transaction. Transaction status: %s. From transaction id: %s", cancelResponse.TransactionStatus, orderId)
				return model.PaymentStatusResponse{}, fmt.Errorf("[FATAL ERROR] Failed to cancel transaction. Transaction status: %s. From transaction id: %s", cancelResponse.TransactionStatus, orderId)
			}
		} else {
			log.Errorf("[FATAL ERROR] Something gone wrong while canceling transaction id: %s", orderId)
			return model.PaymentStatusResponse{}, fmt.Errorf("[FATAL ERROR] Something gone wrong while canceling transaction id: %s", orderId)
		}
	}

	res := model.PaymentStatusResponse{
		PaymentStatus: paymentStatus,
		StatusCode:    cancelResponse.StatusCode,
		Message:       "Cancelled successfully",
	}

	return res, nil
}
