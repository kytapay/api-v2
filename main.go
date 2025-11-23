package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kytapay/api-v2/config"
	"github.com/kytapay/api-v2/controllers"
	"github.com/kytapay/api-v2/controllers/payment"
	"github.com/kytapay/api-v2/controllers/payout"
	"github.com/kytapay/api-v2/middleware"
	"github.com/kytapay/api-v2/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Gin router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Initialize controllers
	tokenController := controllers.NewTokenController(db)
	paymentChannelsController := controllers.NewPaymentChannelsController(db)
	merchantInformationController := controllers.NewMerchantInformationController(db)
	vaPaymentController := payment.NewVAPaymentController(db)
	qrisPaymentController := payment.NewQRISPaymentController(db)
	ewalletPaymentController := payment.NewEWalletPaymentController(db)
	paymentDetailsController := payment.NewPaymentDetailsController(db)
	paymentCancelController := payment.NewPaymentCancelController(db)
	payoutInquiryController := payout.NewPayoutInquiryController(db)
	payoutTransfersController := payout.NewPayoutTransfersController(db)
	payoutDetailsController := payout.NewPayoutDetailsController(db)
	balanceController := controllers.NewBalanceController(db)

	// Setup routes
	routes.SetupRoutes(r, tokenController, paymentChannelsController, merchantInformationController,
		vaPaymentController, qrisPaymentController, ewalletPaymentController,
		paymentDetailsController, paymentCancelController,
		payoutInquiryController, payoutTransfersController, payoutDetailsController,
		balanceController)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

