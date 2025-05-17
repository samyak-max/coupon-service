package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/farmako/coupon-service/internal/handler"
	"github.com/farmako/coupon-service/internal/repository"
	"github.com/farmako/coupon-service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Coupon Service API
// @version 1.0
// @description This is a coupon management service for the medicine ordering platform.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost user=postgres password=postgres dbname=farmako_coupons port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Coupon{}, &models.CouponUsage{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize dependencies
	couponRepo := repository.NewCouponRepository(db)
	couponService := service.NewCouponService(couponRepo)
	couponHandler := handler.NewCouponHandler(couponService)

	// Setup Gin router
	router := gin.Default()
	router.Use(gin.Recovery())

	// API versioning
	v1 := router.Group("/api/v1")
	{
		coupons := v1.Group("/coupons")
		{
			coupons.POST("", couponHandler.CreateCoupon)
			coupons.GET("/applicable", couponHandler.GetApplicableCoupons)
			coupons.POST("/validate", couponHandler.ValidateCoupon)
		}
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 