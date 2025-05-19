package repository

import (
	"context"
	"time"

	"github.com/samyak-max/coupon-service/internal/models"
	"gorm.io/gorm"
)

type CouponRepository interface {
	Create(ctx context.Context, coupon *models.Coupon) error
	GetByCode(ctx context.Context, code string) (*models.Coupon, error)
	GetApplicableCoupons(ctx context.Context, orderTotal float64, timestamp time.Time) ([]models.Coupon, error)
	RecordUsage(ctx context.Context, usage *models.CouponUsage) error
	GetUserUsageCount(ctx context.Context, couponID uint, userID string) (int, error)
}

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

func (r *couponRepository) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = true", code).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetApplicableCoupons(ctx context.Context, orderTotal float64, timestamp time.Time) ([]models.Coupon, error) {
	var coupons []models.Coupon
	err := r.db.WithContext(ctx).Where("is_active = true AND expiry_date > ? AND min_order_value <= ?", timestamp, orderTotal).Find(&coupons).Error
	return coupons, err
}

func (r *couponRepository) RecordUsage(ctx context.Context, usage *models.CouponUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *couponRepository) GetUserUsageCount(ctx context.Context, couponID uint, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", couponID, userID).Count(&count).Error
	return int(count), err
}
