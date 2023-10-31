package handler_midtrans

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go/snap"
)

type SnapHandler interface {
	CreateTransaction(c *gin.Context)
	CreateOneBill(c *gin.Context)
	EditOneBill(c *gin.Context)
	CreateMultipleBill(c *gin.Context)
	GetAllOwnBills(c *gin.Context)
	GetAllBills(c *gin.Context)
	GetBillByID(c *gin.Context)
	SaveMidtransCredentials(c *gin.Context)
	GetMidtransCredentials(c *gin.Context)
	GetMidtransFrontendCredentials(c *gin.Context)
	GetAllTransactions(c *gin.Context)
	UpdateTransactionStatusAndBill(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) SnapHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) CreateTransaction(c *gin.Context) {

	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	var request *snap.Request

	if err := c.ShouldBindJSON(&request); err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	billID := c.DefaultQuery("bill_id", "")
	if billID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "billID is required"})
		return
	}

	// Retrieve the Bill information from the database
	bill, err := h.s.Midtrans().GetBillByID(c, billID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Bill information"})
		return
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("Error loading time zone:", err)
		return
	}

	currentLocalTime := time.Now().In(loc)

	fmt.Println(currentLocalTime)
	fmt.Println(bill.Deadline)

	if bill.Deadline.Before(currentLocalTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot pay after deadline"})
		return
	}

	if bill.RemainingAmount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "your bill is already paid completely"})
		return
	}

	if request.TransactionDetails.GrossAmt < bill.AdminFee {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gross amount should be greater than admin fee"})
		return
	}

	if request.TransactionDetails.GrossAmt-bill.AdminFee > bill.RemainingAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gross amount should be less than remaining fee"})
		return
	}

	// Create the transaction using the service
	response, err := h.s.Midtrans().CreateTransaction(c, request, billID, gotAccount.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Snap transaction"})
		return
	}

	// Serialize the response to JSON and send it to the client
	c.JSON(http.StatusOK, response)
}

func (h *impHandler) EditOneBill(c *gin.Context) {
	id := c.Param("id")

	var request rq.BillRequest = rq.BillRequest{ID: &id}
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Midtrans().EditOneBill(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) CreateOneBill(c *gin.Context) {
	var request rq.BillRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Midtrans().CreateOneBill(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateMultipleBill(c *gin.Context) {
	var requests []*rq.BillRequest
	err := c.ShouldBindJSON(&requests)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	for _, request := range requests {
		err = h.s.Midtrans().CreateOneBill(c, request)
		if err != nil {
			rs.ErrorResponse(c, err)
			return
		}
	}
	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) GetAllOwnBills(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchDescription := c.DefaultQuery("description", "")

	params := rq.PaginationParams[model.Bill]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Bill{
			Description: &searchDescription,
		},
	}

	res, err := h.s.Midtrans().GetBillsByAccountID(c, gotAccount.ID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllBills(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchAccountID := c.DefaultQuery("account_id", "")
	parsedAccountID, _ := uuid.Parse(searchAccountID)

	searchDescription := c.DefaultQuery("description", "")

	filters := model.Bill{
		AccountID:   parsedAccountID,
		Description: &searchDescription,
	}

	if len(searchAccountID) >= 2 {
		parsedAccountID, _ := uuid.Parse(searchAccountID)
		filters.AccountID = parsedAccountID
	}

	params := rq.PaginationParams[model.Bill]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	res, err := h.s.Midtrans().GetAllBills(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetBillByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.Midtrans().GetBillByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) SaveMidtransCredentials(c *gin.Context) {
	var request rq.MidtransCredentials

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.s.Midtrans().SetMidtransCredentials(c, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save Midtrans credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Midtrans credentials saved successfully",
	})
}

func (h *impHandler) GetMidtransCredentials(c *gin.Context) {
	// Handle getting Midtrans credentials here
	credentials, err := h.s.Midtrans().GetMidtransCredentials()

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, credentials, http.StatusOK)
}

func (h *impHandler) GetMidtransFrontendCredentials(c *gin.Context) {
	// Handle getting Midtrans credentials here
	frontendCredentials, err := h.s.Midtrans().GetMidtransFrontendCredentials()

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, frontendCredentials, http.StatusOK)
}

func (h *impHandler) GetAllTransactions(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchTransactionTime := c.DefaultQuery("transaction_time", "")

	params := rq.PaginationParams[any]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Transaction{
			TransactionTime: searchTransactionTime,
		},
	}

	res, err := h.s.Midtrans().GetAllTransactions(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) UpdateTransactionStatusAndBill(c *gin.Context) {
	err := h.s.Midtrans().UpdateDatabaseJob()
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}
	rs.SuccessResponse(c, "Database update job completed successfully", http.StatusOK)
}
