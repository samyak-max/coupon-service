package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyak-max/coupon-service/internal/models"
	"github.com/samyak-max/coupon-service/internal/service"
)

type CouponHandler struct {
	service service.CouponService
}

func NewCouponHandler(service service.CouponService) *CouponHandler {
	return &CouponHandler{
		service: service,
	}
}

// @Summary Create a new coupon
// @Description Create a new coupon with the provided details
// @Tags coupons
// @Accept json
// @Produce json
// @Param coupon body models.Coupon true "Coupon details"
// @Success 201 {object} models.Coupon
// @Failure 400 {object} map[string]string
// @Router /coupons [post]
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var coupon models.Coupon
	err := c.ShouldBindJSON(&coupon)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.CreateCoupon(c.Request.Context(), &coupon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

// @Summary Get applicable coupons
// @Description Get all applicable coupons for the given cart
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body models.ValidationRequest true "Cart details"
// @Success 200 {array} models.Coupon
// @Failure 400 {object} map[string]string
// @Router /coupons/applicable [get]
func (h *CouponHandler) GetApplicableCoupons(c *gin.Context) {
	var req models.ValidationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupons, err := h.service.GetApplicableCoupons(
		c.Request.Context(),
		req.CartItems,
		req.OrderTotal,
		req.Timestamp,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applicable_coupons": coupons,
	})
}

// @Summary Validate a coupon
// @Description Validate a coupon for the given cart
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body models.ValidationRequest true "Validation request"
// @Success 200 {object} models.ValidationResponse
// @Failure 400 {object} map[string]string
// @Router /coupons/validate [post]
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var req models.ValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.ValidateCoupon(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
} 