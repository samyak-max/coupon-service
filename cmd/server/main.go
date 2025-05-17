package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/samyak-max/coupon-service/internal/handler"
	"github.com/samyak-max/coupon-service/internal/models"
	"github.com/samyak-max/coupon-service/internal/repository"
	"github.com/samyak-max/coupon-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Coupon Service API
// @version 1.0
// @description This is a coupon management service for the medicine ordering platform.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost user=postgres password=postgres dbname=coupons port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Coupon{}, &models.CouponUsage{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	couponRepo := repository.NewCouponRepository(db)
	couponService := service.NewCouponService(couponRepo)
	couponHandler := handler.NewCouponHandler(couponService)

	router := gin.Default()
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		coupons := v1.Group("/coupons")
		{
			coupons.POST("", couponHandler.CreateCoupon)
			coupons.GET("/applicable", couponHandler.GetApplicableCoupons)
			coupons.POST("/validate", couponHandler.ValidateCoupon)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
