package service_midtrans

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/robfig/cron"
	"gorm.io/gorm"
)

type SnapService interface {
	SetMidtransCredentials(c context.Context, newRequest *rq.MidtransCredentials) error
	GetMidtransCredentials() (*rs.MidtransCredentialsResponse, error)
	GetMidtransFrontendCredentials() (*rs.MidtransFrontendResponse, error)
	CreateTransaction(c context.Context, newRequest *snap.Request, billID string, email string) (*snap.Response, error)
	UpdateTransactionDetailByToken(token string) (*rs.TransactionResponse, error)
	CreateOneBill(c context.Context, newBill *rq.BillRequest) error
	EditOneBill(c context.Context, newBill *rq.BillRequest) error
	GetBillsByAccountID(c context.Context, accountID string, params *rq.PaginationParams[model.Bill]) (*rs.PaginationResponse[any, rs.BillResponse], error)
	GetAllBills(c context.Context, params *rq.PaginationParams[model.Bill]) (*rs.PaginationResponse[any, rs.BillResponse], error)
	GetBillByID(c context.Context, billID string) (*model.Bill, error)
	StartCronJob()
}

type impService struct {
	r          repo.Repository
	snapClient snap.Client
	cron       *cron.Cron
}

func Init(r repo.Repository) SnapService {
	// Create a new cron instance
	c := cron.New()
	s := &impService{
		r:    r,
		cron: c,
	}

	c.AddFunc("0 0 * * * *", s.UpdateDatabaseJob)

	c.Start()

	return s
}

func (s *impService) StartCronJob() {
	// Start the cron job
	s.cron.AddFunc("0 0 * * * *", s.UpdateDatabaseJob)
	s.cron.Start()
}

func (s *impService) SetMidtransCredentials(c context.Context, newRequest *rq.MidtransCredentials) error {
	err := s.r.Midtrans().SaveMidtransCredentials(newRequest)
	if err != nil {
		return err
	}
	return nil
}

func (s *impService) GetMidtransCredentials() (*rs.MidtransCredentialsResponse, error) {
	credentials, err := s.r.Midtrans().GetMidtransCredentials()
	if err != nil {
		return nil, err
	}

	credentialsResponse := rs.MidtransCredentialsResponse{
		ServerKey:      credentials.ServerKey,
		Environment:    credentials.Environment,
		TransactionAPI: credentials.TransactionAPI,
		ClientKey:      credentials.ClientKey,
		SnapJSUrl:      credentials.SnapJSUrl,
	}

	return &credentialsResponse, nil
}

func (s *impService) GetMidtransFrontendCredentials() (*rs.MidtransFrontendResponse, error) {
	frontendCredentials, err := s.r.Midtrans().GetMidtransCredentials()
	if err != nil {
		return nil, err
	}

	frontendCredentialsResponse := rs.MidtransFrontendResponse{
		ClientKey: frontendCredentials.ClientKey,
		SnapJSUrl: frontendCredentials.SnapJSUrl,
	}

	return &frontendCredentialsResponse, nil
}

func (s *impService) initializeSnapClient() {
	credentials, err := s.r.Midtrans().GetMidtransCredentials()
	if err != nil {
		return
	}
	s.snapClient.New(credentials.ServerKey, midtrans.EnvironmentType(credentials.Environment))

	// midtrans.SetPaymentAppendNotification("https://example.com/append")
	// midtrans.SetPaymentOverrideNotification("https://example.com/override")
}

func (s *impService) UpdateDatabaseJob() {
	// Fetch tokens to update from the database using the GetTokensFromDatabase function
	tokensToUpdate, err := s.r.Midtrans().GetTokensFromDatabase()
	if err != nil {
		fmt.Printf("Error fetching tokens from the database: %v\n", err)
		return
	}

	// tokensToUpdate = s.filterCompletedTransactions(tokensToUpdate)

	for _, token := range tokensToUpdate {
		transactionResp, err := s.UpdateTransactionDetailByToken(token)
		if err != nil {
			fmt.Printf("Error updating transaction with token %s: %v\n", token, err)
		} else {
			fmt.Printf("Transaction Status: %s", transactionResp.TransactionStatus) // Print the transaction status
			if transactionResp.TransactionStatus == "settlement" {
				billID, err := s.r.Midtrans().GetBillIDByToken(token) // Get the billID from your own database
				fmt.Printf("Bill ID is %s", billID)
				if err != nil {
					fmt.Printf("Error getting Bill ID for token %s: %v\n", token, err)
				} else if billID != uuid.Nil {
					fmt.Printf("Bill ID: %s", billID)                              // Print the Bill ID
					err := s.updateBillBasedOnSettledTransactions(billID.String()) // Pass the billID as a string
					if err != nil {
						fmt.Printf("Error updating Bill with ID %s: %v\n", billID, err)
					}
				} else {
					fmt.Println("Bill ID is empty.")
				}
			}
		}
	}
}

func (s *impService) UpdateTransactionDetailByToken(token string) (*rs.TransactionResponse, error) {
	s.initializeSnapClient()

	credentials, err := s.r.Midtrans().GetMidtransCredentials()
	if err != nil {
		return nil, err
	}

	apiUrl := credentials.TransactionAPI

	newAPIURL := apiUrl + token + "/status"

	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, newAPIURL, nil)
	if err != nil {
		return nil, err
	}

	// Set Basic Authentication headers
	username := credentials.ServerKey
	password := ""
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	req.Header.Set("Authorization", authHeader)

	// Send the request and retrieve the response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var transactionResp *rs.TransactionResponse

	if err := json.NewDecoder(resp.Body).Decode(&transactionResp); err != nil {
		fmt.Println("Error decoding API response", err)
		return nil, err
	}

	fmt.Printf("API response: %v", transactionResp)

	transaction := model.Transaction{
		TransactionID:     transactionResp.TransactionID,
		TransactionStatus: transactionResp.TransactionStatus,
		PaymentType:       transactionResp.PaymentType,
		TransactionTime:   transactionResp.TransactionTime,
		SettlementTime:    transactionResp.SettlementTime,
		ExpiryTime:        transactionResp.ExpiryTime,
	}

	// Update the database record with a WHERE condition
	err = s.r.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Transaction{}).Where("token = ?", transactionResp.Token).Updates(&transaction).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transactionResp, nil
}

func (s *impService) updateBillBasedOnSettledTransactions(billID string) error {
	// Fetch all settled transactions associated with the Bill ID
	settledTransactions, err := s.r.Midtrans().GetSettledTransactionsByBillID(billID)
	if err != nil {
		return err
	}

	// Calculate purchased and remaining amounts based on settled transactions
	purchasedAmount := calculatePurchasedAmount(settledTransactions)
	remainingAmount := s.calculateRemainingAmount(billID, purchasedAmount)

	// Update the Bill table with the calculated amounts
	err = s.r.Midtrans().UpdateBillAmounts(billID, purchasedAmount, remainingAmount)
	if err != nil {
		return err
	}

	return nil
}

func calculatePurchasedAmount(transactions []model.Transaction) int64 {
	var purchasedAmount int64 = 0
	for _, transaction := range transactions {
		purchasedAmount += transaction.BillFee
	}
	return purchasedAmount
}

func (s *impService) calculateRemainingAmount(billID string, purchasedAmount int64) int64 {

	bill, err := s.r.Midtrans().GetBillByID(billID)
	if err != nil {
		return 0
	}
	remainingAmount := bill.GrossAmount - purchasedAmount
	return remainingAmount
}

func (s *impService) CreateTransaction(c context.Context, newRequest *snap.Request, billID string, email string) (*snap.Response, error) {
	s.initializeSnapClient()

	// Retrieve the Bill information from the database
	bill, err := s.r.Midtrans().GetBillByID(billID) // Replace with your actual method and struct
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	currentLocalTime := time.Now().In(loc)

	if bill.Deadline.Before(currentLocalTime) {
		return nil, errors.New("cannot pay after the deadline")
	}

	if bill.RemainingAmount == 0 {
		return nil, errors.New("your bill is already paid completely")
	}

	// Add a condition to check if GrossAmt is less than AdminFee
	if newRequest.TransactionDetails.GrossAmt < bill.AdminFee {
		return nil, errors.New("gross amount should be greater than admin fee")
	}

	if (newRequest.TransactionDetails.GrossAmt - bill.AdminFee) > bill.RemainingAmount {
		return nil, errors.New("gross amount should be less than admin fee")
	}

	// Create a new request with the transaction details

	orderID := uuid.New().String()

	newRequest.TransactionDetails.OrderID = orderID

	if newRequest.CustomerDetail == nil {
		newRequest.CustomerDetail = &midtrans.CustomerDetails{}
	}

	newRequest.CustomerDetail.Email = email
	newRequest.CustomerDetail.FName = *bill.Account.Name

	var items []midtrans.ItemDetails

	// First item: Bill Fee
	billFee := midtrans.ItemDetails{
		ID:    "bill_fee",
		Name:  "Bill Fee",
		Price: newRequest.TransactionDetails.GrossAmt - bill.AdminFee,
		Qty:   1,
	}

	// Second item: Admin Fee
	adminFee := midtrans.ItemDetails{
		ID:    "admin_fee",
		Name:  "Admin Fee",
		Price: bill.AdminFee,
		Qty:   1,
	}

	items = append(items, billFee, adminFee)

	// Set the items in the newRequest
	newRequest.Items = &items

	var enablePayments []string
	if err := json.Unmarshal([]byte(bill.EnablePayments), &enablePayments); err != nil {
		return nil, err
	}

	// Create a slice to hold the converted payment types
	var enabledPayments []snap.SnapPaymentType

	for _, paymentType := range enablePayments {
		// Convert each payment type string to a snap.SnapPaymentType
		snapPayment := snap.SnapPaymentType(paymentType)
		enabledPayments = append(enabledPayments, snapPayment)
	}

	newRequest.EnabledPayments = enabledPayments

	response, createErr := s.snapClient.CreateTransaction(newRequest)
	if createErr != nil {
		return nil, createErr
	}

	parsedBillID, err := serviceutil.GetUUIDFromStringWithValidation("Bill ID", &billID)
	if err != nil {
		return nil, err
	}

	snap := model.Transaction{
		OrderID:     orderID,
		GrossAmount: newRequest.TransactionDetails.GrossAmt,
		BillFee:     newRequest.TransactionDetails.GrossAmt - bill.AdminFee,
		AdminFee:    bill.AdminFee,
		Email:       email,
		Token:       response.Token,
		BillID:      *parsedBillID,
	}

	transactionErr := s.r.Transaction(func(tx *gorm.DB) error {
		if err := s.r.Midtrans().CreateOne(tx, &snap); err != nil {
			return err
		}
		return nil
	})

	if transactionErr != nil {
		return response, &errmsg.ErrInternal{Err: transactionErr}
	}

	return response, nil
}

func (s *impService) CreateOneBill(c context.Context, newBill *rq.BillRequest) error {
	if newBill.AccountID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "AccountID"}
	}

	if newBill.GrossAmount == 0 {
		return &errmsg.ErrIsEmpty{FieldName: "GrossAmount"}
	}

	if len(newBill.Description) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Description"}
	}

	if newBill.Deadline == "" {
		return &errmsg.ErrIsEmpty{FieldName: "Deadline"}
	}

	parsedAccountID, err := serviceutil.GetUUIDFromStringWithValidation("Account ID", newBill.AccountID)
	if err != nil {
		return err
	}

	// Serialize the EnablePayments slice to a JSON string
	enablePaymentsJSON, err := json.Marshal(newBill.EnablePayments)
	if err != nil {
		return err
	}

	layout := "2006-01-02 15:04"
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return err
	}
	parsedDeadline, err := time.ParseInLocation(layout, newBill.Deadline, location)
	if err != nil {
		return err
	}

	bill := model.Bill{
		AccountID:       *parsedAccountID,
		GrossAmount:     newBill.GrossAmount,
		RemainingAmount: newBill.GrossAmount,
		Description:     &newBill.Description,
		AdminFee:        newBill.AdminFee,
		EnablePayments:  string(enablePaymentsJSON),
		Deadline:        parsedDeadline,
	}

	err = s.r.Transaction(func(tx *gorm.DB) error {
		if err := s.r.Midtrans().CreateOneBill(tx, &bill); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) GetAllBills(c context.Context, params *rq.PaginationParams[model.Bill]) (*rs.PaginationResponse[any, rs.BillResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotBills, maxPage, rowCount, err := s.r.Midtrans().GetAllBills(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.BillResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotBills == nil {
		return &rs.PaginationResponse[any, rs.BillResponse]{
			Data: []rs.BillResponse{},
		}, nil
	}

	if len(*gotBills) < 1 {
		return &rs.PaginationResponse[any, rs.BillResponse]{
			Data: []rs.BillResponse{},
		}, nil
	}

	var datares []rs.BillResponse
	err = copier.Copy(&datares, gotBills)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.BillResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}

func (s *impService) EditOneBill(c context.Context, newBill *rq.BillRequest) error {

	parsedBillID, err := serviceutil.GetUUIDFromStringWithValidation("Bill  ID", newBill.ID)
	if err != nil {
		return err
	}

	layout := "2006-01-02 15:04"
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return err
	}
	parsedDeadline, err := time.ParseInLocation(layout, newBill.Deadline, location)
	if err != nil {
		return err
	}

	// Fetch the existing bill data from the database
	existingBill, err := s.r.Midtrans().GetBillByID(parsedBillID.String())
	if err != nil {
		return err
	}

	// Check if EnablePayments is empty, and if it is, use the existing bill's EnablePayments
	if len(newBill.EnablePayments) == 0 {
		if err := json.Unmarshal([]byte(existingBill.EnablePayments), &newBill.EnablePayments); err != nil {
			return err
		}
	}

	// Convert the EnablePayments slice into a JSON string
	enablePaymentsJSON, err := json.Marshal(newBill.EnablePayments)
	if err != nil {
		return err
	}

	// Create a new bill object and assign values from the existing bill
	bill := model.Bill{
		Base:            model.Base{ID: *parsedBillID},
		AccountID:       existingBill.AccountID,
		AdminFee:        newBill.AdminFee,
		Deadline:        parsedDeadline,
		GrossAmount:     existingBill.GrossAmount,     // Keep the existing GrossAmount
		PurchasedAmount: existingBill.PurchasedAmount, // Keep the existing PurchasedAmount
		RemainingAmount: existingBill.RemainingAmount, // Keep the existing RemainingAmount
		EnablePayments:  string(enablePaymentsJSON),
		Description:     &newBill.Description,
	}

	err = s.r.Transaction(func(tx *gorm.DB) error {
		_, err := s.r.Midtrans().UpdateOneBill(&bill)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) GetBillsByAccountID(c context.Context, accountID string, params *rq.PaginationParams[model.Bill]) (*rs.PaginationResponse[any, rs.BillResponse], error) {
	parsedAccountID, err := serviceutil.GetUUIDFromStringWithValidation("Account ID", &accountID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.AccountID = *parsedAccountID

	return s.GetAllBills(c, &newParams)
}

func (s *impService) GetBillByID(c context.Context, billID string) (*model.Bill, error) {
	bill, err := s.r.Midtrans().GetBillByID(billID)
	if err != nil {
		return nil, err
	}

	return bill, nil
}

// func (s *impService) filterCompletedTransactions(tokens []string) []string {
// 	var filteredTokens []string

// 	for _, token := range tokens {
// 		transactionResp, err := s.UpdateTransactionDetailByToken(token)
// 		if err != nil {
// 			fmt.Printf("Error updating transaction with token %s: %v\n", token, err)
// 			// Handle the error as needed
// 			continue
// 		}

// 		if transactionResp.TransactionStatus != "settlement" {
// 			// Add the token to the filtered list if not settled
// 			filteredTokens = append(filteredTokens, token)
// 		}
// 	}

// 	return filteredTokens
// }
