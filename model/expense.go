package model

import (
	"time"

	"gorm.io/gorm"
)

// User defined expense category, per tenant
type ExpenseCategory struct {
	Id        int        `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	TenantId  int        `json:"tenant_id" gorm:"column:tenant_id"`
	Name      string     `json:"name" gorm:"column:name"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
}

func (expenseCategory *ExpenseCategory) TableName() string {
	return "expense_category"
}

/*
Manual business expense. Stock purchases are not recorded here.
  - StoreId nil means the expense belongs to the whole business (tenant)
  - ExpenseCategoryId nil means uncategorized (-1 at the search filter)
*/
type Expense struct {
	Id                          int            `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	TenantId                    int            `json:"tenant_id" gorm:"column:tenant_id"`
	StoreId                     *int           `json:"store_id" gorm:"column:store_id"`
	ExpenseCategoryId           *int           `json:"expense_category_id" gorm:"column:expense_category_id"`
	Amount                      int            `json:"amount" gorm:"column:amount"`
	Description                 string         `json:"description" gorm:"column:description"`
	ExpenseDate                 time.Time      `json:"expense_date" gorm:"column:expense_date"`
	PaymentType                 PaymentType    `json:"payment_type" gorm:"column:payment_type"`
	ResponsibleUserId           int            `json:"responsible_user_id" gorm:"column:responsible_user_id;<-:create"`
	ResponsibleUserNameSnapshot string         `json:"responsible_user_name_snapshot" gorm:"column:responsible_user_name_snapshot;<-:create"`
	CreatedAt                   *time.Time     `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
	UpdatedAt                   *time.Time     `json:"updated_at,omitempty" gorm:"column:updated_at"`
	DeletedAt                   gorm.DeletedAt `json:"-"` // Soft delete

	// Read only, joined from expense_category
	ExpenseCategoryName *string `json:"expense_category_name" gorm:"->;column:expense_category_name"`
}

func (expense *Expense) TableName() string {
	return "expense"
}
