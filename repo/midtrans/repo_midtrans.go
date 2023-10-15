package repo_midtrans

import (
	"errors"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MidtransRepo interface {
	GetAll(params *rq.PaginationParams[any]) (*[]model.Transaction, int, int, error)
	CreateOne(tx *gorm.DB, newCreation *model.Transaction) error
	UpdateOne(tx *gorm.DB, updateTx *model.Transaction) error
	GetTokensFromDatabase() ([]string, error)
	FindBillByAccountId(accountId string) (*model.Bill, error)
	CreateOneBill(tx *gorm.DB, newBill *model.Bill) error
	UpdateOneBill(bill *model.Bill) (*model.Bill, error)
	GetBillByEmail(email string) (*model.Bill, error)
	GetBillByUserId(accountId string) (*model.Bill, error)
	GetAllBills(params *rq.PaginationParams[model.Bill]) (*[]model.Bill, int, int, error)
	GetBillByID(id string) (*model.Bill, error)
	GetSettledTransactionsByBillID(billID string) ([]model.Transaction, error)
	UpdateBillAmounts(billID string, purchasedAmount int64, remainingAmount int64) error
	GetBillIDByToken(token string) (uuid.UUID, error)
	SaveMidtransCredentials(newCredential *rq.MidtransCredentials) error
	GetMidtransCredentials() (*model.MidtransCredentials, error)
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) MidtransRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params *rq.PaginationParams[any]) (*[]model.Transaction, int, int, error) {
	var entries []model.Transaction

	chain := r.db

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&entries), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&entries)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &entries, maxPage, rowCount, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newCreation *model.Transaction) error {
	if err := tx.Create(newCreation).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, updateTx *model.Transaction) error {
	if updateTx == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Article"}
	}

	result := tx.Model(updateTx).Updates(updateTx)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) GetTokensFromDatabase() ([]string, error) {
	var tokens []string
	result := r.db.Model(&model.Transaction{}).Pluck("token", &tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

func (r *impRepo) FindBillByAccountId(accountId string) (*model.Bill, error) {
	var bill *model.Bill

	err := r.db.Where("account_id = ?", accountId).Find(&bill).Error
	if err != nil {
		return bill, err
	}

	return bill, nil
}

func (r *impRepo) CreateOneBill(tx *gorm.DB, newBill *model.Bill) error {
	if err := tx.Create(newBill).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOneBill(bill *model.Bill) (*model.Bill, error) {
	err := r.db.Save(&bill).Error
	if err != nil {
		return bill, err
	}

	return bill, nil
}

func (r *impRepo) GetBillByEmail(email string) (*model.Bill, error) {
	var bill model.Bill
	if err := r.db.
		Joins("JOIN accounts ON bills.account_id = accounts.id").
		Where("accounts.email = ?", email).
		First(&bill).Error; err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *impRepo) GetBillByID(id string) (*model.Bill, error) {
	var bill model.Bill

	chain := r.db.Preload("Account")
	if err := chain.Where("id = ?", id).Find(&bill).Error; err != nil {
		return nil, err
	}

	return &bill, nil
}

func (r *impRepo) GetBillByUserId(accountId string) (*model.Bill, error) {
	var bill model.Bill

	chain := r.db.Preload("Account")
	if err := chain.Where("account_id = ?", accountId).Find(&bill).Error; err != nil {
		return nil, err
	}

	return &bill, nil
}

func (r *impRepo) GetAllBills(params *rq.PaginationParams[model.Bill]) (*[]model.Bill, int, int, error) {
	var bills []model.Bill

	chain := r.db

	if params.Data.AccountID != uuid.Nil {
		chain = chain.Where(r.db.Where("account_id = ?", params.Data.AccountID.String()))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&bills), params.Limit)

	validColumDescription := []string{
		"description",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumDescription)).Find(&bills)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &bills, maxPage, rowCount, nil
}

func (r *impRepo) GetSettledTransactionsByBillID(billID string) ([]model.Transaction, error) {
	var transactions []model.Transaction

	if err := r.db.Where("bill_id = ? AND transaction_status = ?", billID, "settlement").Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *impRepo) UpdateBillAmounts(billID string, purchasedAmount int64, remainingAmount int64) error {
	if err := r.db.Model(&model.Bill{}).Where("id = ?", billID).Updates(map[string]interface{}{
		"purchased_amount": purchasedAmount,
		"remaining_amount": remainingAmount,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) GetBillIDByToken(token string) (uuid.UUID, error) {
	var transaction model.Transaction
	if err := r.db.Where("token = ?", token).First(&transaction).Error; err != nil {
		return uuid.Nil, err
	}
	return transaction.BillID, nil
}

func (r *impRepo) SaveMidtransCredentials(newCredential *rq.MidtransCredentials) error {
	// Check if the record exists based on the serverKey
	existingCredentials := model.MidtransCredentials{}

	// Attempt to retrieve the record with the given serverKey
	err := r.db.Where("server_key IS NOT NULL").First(&existingCredentials).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If the record doesn't exist, create a new one
			newCredentials := model.MidtransCredentials{
				ServerKey:      newCredential.ServerKey,
				ClientKey:      newCredential.ClientKey,
				Environment:    newCredential.Environment,
				TransactionAPI: newCredential.TransactionAPI,
				SnapJSUrl:      newCredential.SnapJSUrl,
			}

			if err := r.db.Create(&newCredentials).Error; err != nil {
				return err // Return any database error
			}
		} else {
			return err // Return any database error other than "not found"
		}
	} else {
		// If the record exists, update its serverKey and environment fields with a WHERE condition
		if err := r.db.Where("server_key IS NOT NULL").Model(&existingCredentials).Updates(model.MidtransCredentials{
			ServerKey:      newCredential.ServerKey,
			ClientKey:      newCredential.ClientKey,
			Environment:    newCredential.Environment,
			TransactionAPI: newCredential.TransactionAPI,
			SnapJSUrl:      newCredential.SnapJSUrl,
		}).Error; err != nil {
			return err // Return any database error
		}
	}

	return nil
}

func (r *impRepo) GetMidtransCredentials() (*model.MidtransCredentials, error) {
	// Example: Retrieving Midtrans credentials from a database table.
	var credentials model.MidtransCredentials

	if err := r.db.First(&credentials).Error; err != nil {
		return nil, err
	}

	return &credentials, nil
}
