package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UsageType string
type DiscountType string

const (
	OneTime   UsageType = "one_time"
	MultiUse  UsageType = "multi_use"
	TimeBased UsageType = "time_based"

	PercentageDiscount DiscountType = "percentage"
	FixedDiscount      DiscountType = "fixed"
)

// Coupon represents the coupon entity in the system
type Coupon struct {
	gorm.Model
	ID                    uint           `gorm:"primaryKey" json:"id"`
	Code                  string         `gorm:"uniqueIndex" json:"code"`
	ExpiryDate            time.Time      `json:"expiry_date"`
	UsageType             UsageType      `json:"usage_type"`
	ApplicableMedicineIDs pq.StringArray `gorm:"type:text[]" json:"applicable_medicine_ids,omitempty"`
	ApplicableCategories  pq.StringArray `gorm:"type:text[]" json:"applicable_categories,omitempty"`
	MinOrderValue         float64        `json:"min_order_value"`
	ValidTimeWindow       *TimeWindow    `gorm:"embedded" json:"valid_time_window,omitempty"`
	TermsAndConditions    string         `json:"terms_and_conditions"`
	DiscountType          DiscountType   `json:"discount_type"`
	DiscountValue         float64        `json:"discount_value"`
	MaxUsagePerUser       int            `json:"max_usage_per_user"`
	IsActive              bool           `json:"is_active" gorm:"default:true"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// TimeWindow represents the valid time window for time-based coupons
type TimeWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// CouponUsage tracks the usage of coupons by users
type CouponUsage struct {
	gorm.Model
	CouponID uint      `json:"coupon_id"`
	UserID   string    `json:"user_id"`
	UsedAt   time.Time `json:"used_at"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// ValidationRequest represents the input for coupon validation
type ValidationRequest struct {
	ID         uint       `json:"id"`
	CouponCode string     `json:"coupon_code"`
	CartItems  []CartItem `json:"cart_items"`
	OrderTotal float64    `json:"order_total"`
	Timestamp  time.Time  `json:"timestamp"`
	UserID     string     `json:"user_id"`
}

// ValidationResponse represents the output of coupon validation
type ValidationResponse struct {
	ID       uint      `json:"id"`
	IsValid  bool      `json:"is_valid"`
	Discount *Discount `json:"discount,omitempty"`
	Message  string    `json:"message,omitempty"`
	Reason   string    `json:"reason,omitempty"`
}

// Discount represents the discount breakdown
type Discount struct {
	ItemsDiscount   float64 `json:"items_discount"`
	ChargesDiscount float64 `json:"charges_discount"`
	TotalDiscount   float64 `json:"total_discount"`
}
