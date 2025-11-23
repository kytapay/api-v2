package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kytapay/api-v2/controllers"
	"github.com/kytapay/api-v2/controllers/payment"
	"github.com/kytapay/api-v2/controllers/payout"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(
	r *gin.Engine,
	tokenController *controllers.TokenController,
	paymentChannelsController *controllers.PaymentChannelsController,
	merchantInformationController *controllers.MerchantInformationController,
	vaPaymentController *payment.VAPaymentController,
	qrisPaymentController *payment.QRISPaymentController,
	ewalletPaymentController *payment.EWalletPaymentController,
	paymentDetailsController *payment.PaymentDetailsController,
	paymentCancelController *payment.PaymentCancelController,
	payoutInquiryController *payout.PayoutInquiryController,
	payoutTransfersController *payout.PayoutTransfersController,
	payoutDetailsController *payout.PayoutDetailsController,
	balanceController *controllers.BalanceController,
) {
	// Health check (support both GET and HEAD for Docker healthcheck)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "KytaPay API v2",
		})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.Status(200)
	})

	// Balance check (X-ADMIN-KEY required)
	r.GET("/balance", balanceController.GetBalance)

	// API v2 routes
	v2 := r.Group("/v2")
	{
		// Access Token (POST because it requires grant_type in body)
		v2.POST("/access-token", tokenController.VerifyClientCredentials)

		// Merchant routes (Bearer token required)
		merchants := v2.Group("/merchants")
		{
			merchants.GET("/payment-channels", paymentChannelsController.PaymentChannels)
			merchants.GET("/information", merchantInformationController.Informations)
		}

		// Payment routes (Bearer token required)
		payments := v2.Group("/payments")
		{
			payments.POST("/create/va", vaPaymentController.CreateVA)
			payments.POST("/create/qris", qrisPaymentController.CreateQRIS)
			payments.POST("/create/ewallet", ewalletPaymentController.CreateEWallet)
			payments.POST("/details", paymentDetailsController.GetDetail)
			payments.POST("/cancel", paymentCancelController.CancelPayment)
		}

		// Payout routes (Bearer token required)
		payouts := v2.Group("/payouts")
		{
			payouts.POST("/inquiry", payoutInquiryController.ProcessInquiry)
			payouts.POST("/transfers", payoutTransfersController.ProcessPayout)
			payouts.POST("/details", payoutDetailsController.GetDetail)
		}
	}
}

