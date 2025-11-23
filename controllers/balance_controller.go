package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kytapay/api-v2/config"
	"github.com/kytapay/api-v2/helpers"
	"github.com/kytapay/api-v2/services"
)

type BalanceController struct {
	db              *sql.DB
	linkQuService   *services.LinkQuService
	pakaiLinkService *services.PakaiLinkService
}

func NewBalanceController(db *sql.DB) *BalanceController {
	return &BalanceController{
		db:              db,
		linkQuService:   services.NewLinkQuService(),
		pakaiLinkService: services.NewPakaiLinkService(),
	}
}

// GetBalance handles GET /balance
func (bc *BalanceController) GetBalance(c *gin.Context) {
	// Check X-ADMIN-KEY
	adminKey := c.GetHeader("X-ADMIN-KEY")
	if adminKey != "VLADMIN" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response_code":    "4011200",
			"response_message": "Unauthorized",
		})
		return
	}

	// Get LinkQu balance
	linkQuResponse, err := bc.linkQuService.GetBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001201",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Parse LinkQu response
	linkQuBalance := 0.0
	linkQuPending := 0.0

	rc, _ := linkQuResponse["rc"].(string)
	if rc == "00" {
		// Get balance
		if balance, ok := linkQuResponse["balance"].(float64); ok {
			linkQuBalance = balance
		} else if balanceStr, ok := linkQuResponse["balance"].(string); ok {
			if balanceFloat, err := strconv.ParseFloat(balanceStr, 64); err == nil {
				linkQuBalance = balanceFloat
			}
		}

		// Get unsettle_amount from data
		if data, ok := linkQuResponse["data"].(map[string]interface{}); ok {
			if unsettleAmount, ok := data["unsettle_amount"].(float64); ok {
				linkQuPending = unsettleAmount
			}
		}
	}

	// Get PakaiLink balance
	// Generate partner reference
	partnerRef := uuid.New().String()
	
	// Get account number from config
	pakaiLinkConfig := config.GetPakaiLinkConfig()
	accountNo := pakaiLinkConfig.AccountNo

	pakaiLinkResponse, err := bc.pakaiLinkService.GetBalance(partnerRef, accountNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response_code":    "5001201",
			"response_message": "Internal Server Error",
		})
		return
	}

	// Parse PakaiLink response
	pakaiLinkBalance := 0.0
	pakaiLinkPending := 0.0

	responseCode, _ := pakaiLinkResponse["responseCode"].(string)
	if responseCode == "2001100" {
		if accountInfo, ok := pakaiLinkResponse["accountInfo"].([]interface{}); ok && len(accountInfo) > 0 {
			if firstAccount, ok := accountInfo[0].(map[string]interface{}); ok {
				// Get activeBalance
				if activeBalance, ok := firstAccount["activeBalance"].(map[string]interface{}); ok {
					if valueStr, ok := activeBalance["value"].(string); ok {
						if valueFloat, err := strconv.ParseFloat(valueStr, 64); err == nil {
							pakaiLinkBalance = valueFloat
						}
					}
				}

				// Get pendingBalance
				if pendingBalance, ok := firstAccount["pendingBalance"].(map[string]interface{}); ok {
					if valueStr, ok := pendingBalance["value"].(string); ok {
						if valueFloat, err := strconv.ParseFloat(valueStr, 64); err == nil {
							pakaiLinkPending = valueFloat
						}
					}
				}
			}
		}
	}

	// Calculate totals
	totalBalance := linkQuBalance + pakaiLinkBalance
	totalPending := linkQuPending + pakaiLinkPending

	// Format to IDR text
	linkQuBalanceText := helpers.FormatIDRAmount(linkQuBalance)
	linkQuPendingText := helpers.FormatIDRAmount(linkQuPending)
	pakaiLinkBalanceText := helpers.FormatIDRAmount(pakaiLinkBalance)
	pakaiLinkPendingText := helpers.FormatIDRAmount(pakaiLinkPending)

	loc, _ := time.LoadLocation("Asia/Jakarta")
	requestTime := time.Now().In(loc).Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"response_code":    "2001200",
		"response_message": "Successful",
		"response_data": gin.H{
			"balance": totalBalance,
			"pending": totalPending,
			"linkqu": gin.H{
				"balance": linkQuBalance,
				"pending": linkQuPending,
			},
			"pakailink": gin.H{
				"balance": pakaiLinkBalance,
				"pending": pakaiLinkPending,
			},
			"text_balance": gin.H{
				"linkqu":    linkQuBalanceText,
				"pakailink": pakaiLinkBalanceText,
			},
			"text_pending": gin.H{
				"linkqu":    linkQuPendingText,
				"pakailink": pakaiLinkPendingText,
			},
			"request_time": requestTime,
		},
	})
}

