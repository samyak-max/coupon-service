package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/samyak-max/coupon-service/internal/models"
	"github.com/samyak-max/coupon-service/internal/repository"
)

var (
	ErrCouponExpired       = errors.New("coupon has expired")
	ErrCouponNotApplicable = errors.New("coupon not applicable to cart items")
	ErrMinOrderValueNotMet = errors.New("minimum order value not met")
	ErrMaxUsageExceeded    = errors.New("maximum usage per user exceeded")
	ErrInvalidTimeWindow   = errors.New("coupon not valid in current time window")
)

type CouponService interface {
	CreateCoupon(ctx context.Context, coupon *models.Coupon) error
	ValidateCoupon(ctx context.Context, req *models.ValidationRequest) (*models.ValidationResponse, error)
	GetApplicableCoupons(ctx context.Context, cartItems []models.CartItem, orderTotal float64, timestamp time.Time) ([]models.Coupon, error)
}

type couponService struct {
	repo  repository.CouponRepository
	cache *cache.Cache
	mu    sync.RWMutex
}

func NewCouponService(repo repository.CouponRepository) CouponService {
	return &couponService{
		repo:  repo,
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (s *couponService) CreateCoupon(ctx context.Context, coupon *models.Coupon) error {
	return s.repo.Create(ctx, coupon)
}

func (s *couponService) ValidateCoupon(ctx context.Context, req *models.ValidationRequest) (*models.ValidationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cache the coupon
	var coupon *models.Coupon
	if cachedCoupon, found := s.cache.Get(req.CouponCode); found {
		coupon = cachedCoupon.(*models.Coupon)
	} else {
		coupon, err := s.repo.GetByCode(ctx, req.CouponCode)
		if err != nil {
			return &models.ValidationResponse{
				IsValid: false,
				Reason:  "coupon not found",
			}, nil
		}
		s.cache.Set(req.CouponCode, coupon, cache.DefaultExpiration)
	}

	// Check expiration
	if coupon.ExpiryDate.Before(req.Timestamp) {
		return &models.ValidationResponse{
			IsValid: false,
			Reason:  ErrCouponExpired.Error(),
		}, nil
	}

	// Check minimum order value
	if req.OrderTotal < coupon.MinOrderValue {
		return &models.ValidationResponse{
			IsValid: false,
			Reason:  ErrMinOrderValueNotMet.Error(),
		}, nil
	}

	// Check time window for time-based coupons
	if coupon.UsageType == models.TimeBased && coupon.ValidTimeWindow != nil {
		if req.Timestamp.Before(coupon.ValidTimeWindow.StartTime) || req.Timestamp.After(coupon.ValidTimeWindow.EndTime) {
			return &models.ValidationResponse{
				IsValid: false,
				Reason:  ErrInvalidTimeWindow.Error(),
			}, nil
		}
	}

	// Check usage limits
	if coupon.UsageType == models.OneTime {
		count, err := s.repo.GetUserUsageCount(ctx, coupon.ID, req.UserID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return &models.ValidationResponse{
				IsValid: false,
				Reason:  ErrMaxUsageExceeded.Error(),
			}, nil
		}
	}

	// Check if coupon is applicable to cart items
	if !s.isCouponApplicableToCart(*coupon, req.CartItems) {
		return &models.ValidationResponse{
			IsValid: false,
			Reason:  ErrCouponNotApplicable.Error(),
		}, nil
	}

	// Calculate discount
	discount := s.calculateDiscount(coupon, req.CartItems, req.OrderTotal)

	// Record usage
	usage := &models.CouponUsage{
		CouponID: coupon.ID,
		UserID:   req.UserID,
		UsedAt:   req.Timestamp,
	}

	err := s.repo.RecordUsage(ctx, usage)
	if err != nil {
		return nil, err
	}

	return &models.ValidationResponse{
		IsValid:  true,
		Discount: discount,
		Message:  "coupon applied successfully",
	}, nil
}

func hashCartSignature(sig string) string {
	sum := sha256.Sum256([]byte(sig))
	return hex.EncodeToString(sum[:])
}

func cartSignature(cartItems []models.CartItem) string {
	var sb strings.Builder
	sort.Slice(cartItems, func(i, j int) bool {
		return cartItems[i].ID < cartItems[j].ID
	})
	for _, item := range cartItems {
		sb.WriteString(fmt.Sprintf("%s|", item.ID))
	}
	return sb.String()
}

func (s *couponService) GetApplicableCoupons(ctx context.Context, cartItems []models.CartItem, orderTotal float64, timestamp time.Time) ([]models.Coupon, error) {
	// Create a cache key based on the parameters
	cartSig := cartSignature(cartItems)
	cacheKey := fmt.Sprintf("applicable_%s_%.2f_%d", hashCartSignature(cartSig), orderTotal, timestamp.Unix())

	// Try to get from cache first
	if cachedCoupons, found := s.cache.Get(cacheKey); found {
		return cachedCoupons.([]models.Coupon), nil
	}

	coupons, err := s.repo.GetApplicableCoupons(ctx, orderTotal, timestamp)
	if err != nil {
		return nil, err
	}

	// Filter coupons based on cart items and categories
	var applicableCoupons []models.Coupon
	for _, coupon := range coupons {
		if s.isCouponApplicableToCart(coupon, cartItems) {
			applicableCoupons = append(applicableCoupons, coupon)
		}
	}

	s.cache.Set(cacheKey, applicableCoupons, cache.DefaultExpiration)
	return applicableCoupons, nil
}

func (s *couponService) calculateDiscount(coupon *models.Coupon, items []models.CartItem, orderTotal float64) *models.Discount {
	var itemsDiscount, chargesDiscount float64

	if coupon.DiscountType == models.PercentageDiscount {
		itemsDiscount = orderTotal * (coupon.DiscountValue / 100)
	} else {
		itemsDiscount = coupon.DiscountValue
	}

	totalDiscount := itemsDiscount + chargesDiscount
	return &models.Discount{
		ItemsDiscount:   itemsDiscount,
		ChargesDiscount: chargesDiscount,
		TotalDiscount:   totalDiscount,
	}
}

func (s *couponService) isCouponApplicableToCart(coupon models.Coupon, items []models.CartItem) bool {
	if len(coupon.ApplicableMedicineIDs) == 0 && len(coupon.ApplicableCategories) == 0 {
		return true
	}

	for _, item := range items {
		// Check if item ID is in applicable medicines
		for _, medicineID := range coupon.ApplicableMedicineIDs {
			if item.ID == medicineID {
				return true
			}
		}

		// Check if item category is in applicable categories
		for _, category := range coupon.ApplicableCategories {
			if item.Category == category {
				return true
			}
		}
	}

	return false
}
