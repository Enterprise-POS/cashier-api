package client

import (
	"cashier-api/model"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	log "github.com/sirupsen/logrus"
)

var s snap.Client

var MIDTRANS_SERVER_KEY = os.Getenv("MIDTRANS_SERVER_KEY")

func setupGlobalMidtransConfig() {
	midtrans.ServerKey = MIDTRANS_SERVER_KEY
	midtrans.Environment = midtrans.Sandbox

	// Optional : here is how if you want to set append payment notification globally
	//midtrans.SetPaymentAppendNotification("https://example.com/append")
	// Optional : here is how if you want to set override payment notification globally
	//midtrans.SetPaymentOverrideNotification("https://example.com/override")

	//// remove the comment bellow, in cases you need to change the default for Log Level
	// midtrans.DefaultLoggerLevel = &midtrans.LoggerImplementation{
	//	 LogLevel: midtrans.LogInfo,
	// }
}

// func initializeSnapClient() {
// 	s.New(MIDTRANS_SERVER_KEY, midtrans.Sandbox)
// }

// func createTransactionWithGlobalConfig() {
// 	res, err := snap.CreateTransactionWithMap(&snap.RequestParamWithMap{
// 		"transaction_details": map[string]interface{}{
// 			"order_id":     "MID-GO-TEST-" + random(),
// 			"gross_amount": 10000,
// 		},
// 	})
// 	if err != nil {
// 		fmt.Println("Snap Request Error", err.GetMessage())
// 	}
// 	fmt.Println("Snap response", res)
// }

// func CreateTransaction(req snap.Request) {
// 	// Optional : here is how if you want to set append payment notification for this request
// 	//s.Options.SetPaymentAppendNotification("https://example.com/append")

// 	// Optional : here is how if you want to set override payment notification for this request
// 	//s.Options.SetPaymentOverrideNotification("https://example.com/override")
// 	// Send request to Midtrans Snap API

// 	resp, err := s.CreateTransaction(&req)
// 	if err != nil {
// 		fmt.Println("Error :", err.GetMessage())
// 	}
// 	fmt.Println("Response : ", resp)
// }

// func createTokenTransactionWithGateway() {
// 	//s.Options.SetPaymentOverrideNotification("https://example.com/url2")

// 	resp, err := s.CreateTransactionToken(GenerateSnapReq())
// 	if err != nil {
// 		fmt.Println("Error :", err.GetMessage())
// 	}
// 	fmt.Println("Response : ", resp)
// }

// func createUrlTransactionWithGateway() {
// 	s.Options.SetContext(context.Background())

// 	resp, err := s.CreateTransactionUrl(GenerateSnapReq())
// 	if err != nil {
// 		fmt.Println("Error :", err.GetMessage())
// 	}
// 	fmt.Println("Response : ", resp)
// }

// func main() {
// 	fmt.Println("================ Request with global config ================")
// 	setupGlobalMidtransConfig()
// 	createTransactionWithGlobalConfig()

// 	fmt.Println("================ Request with Snap Client ================")
// 	initializeSnapClient()
// 	createTransaction()

// 	fmt.Println("================ Request Snap token ================")
// 	createTokenTransactionWithGateway()

// 	fmt.Println("================ Request Snap URL ================")
// 	createUrlTransactionWithGateway()
// }

// func GenerateSnapReq() *snap.Request {

// 	// Initiate Customer address
// 	custAddress := &midtrans.CustomerAddress{
// 		FName:       "John",
// 		LName:       "Doe",
// 		Phone:       "081234567890",
// 		Address:     "Baker Street 97th",
// 		City:        "Jakarta",
// 		Postcode:    "16000",
// 		CountryCode: "IDN",
// 	}

// 	// Initiate Snap Request
// 	snapReq := &snap.Request{
// 		TransactionDetails: midtrans.TransactionDetails{
// 			OrderID:  "MID-GO-ID-" + random(),
// 			GrossAmt: 200000,
// 		},
// 		CreditCard: &snap.CreditCardDetails{
// 			Secure: true,
// 		},
// 		CustomerDetail: &midtrans.CustomerDetails{
// 			FName:    "John",
// 			LName:    "Doe",
// 			Email:    "john@doe.com",
// 			Phone:    "081234567890",
// 			BillAddr: custAddress,
// 			ShipAddr: custAddress,
// 		},
// 		EnabledPayments: snap.AllSnapPaymentType,
// 		Items: &[]midtrans.ItemDetails{
// 			{
// 				ID:    "ITEM1",
// 				Price: 200000,
// 				Qty:   1,
// 				Name:  "Someitem",
// 			},
// 		},
// 	}
// 	return snapReq
// }

// func random() string {
// 	time.Sleep(500 * time.Millisecond)
// 	return strconv.FormatInt(time.Now().Unix(), 10)
// }

type MidtransProvider struct {
	snapClient snap.Client
	coreApi    coreapi.Client
}

func NewMidtransProvider() *MidtransProvider {
	var s snap.Client
	s.New(MIDTRANS_SERVER_KEY, midtrans.Sandbox)

	var c coreapi.Client
	c.New(MIDTRANS_SERVER_KEY, midtrans.Sandbox)

	return &MidtransProvider{snapClient: s, coreApi: c}
}

func (m *MidtransProvider) CreateTransaction(req *snap.Request) (*model.PaymentProviderResponse, error) {
	resp, err := m.snapClient.CreateTransaction(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s (%s)", err.GetMessage(), err.GetStatusCode()))
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
