package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kytapay/api-v2/config"
)

type LinkQuService struct {
	client *http.Client
	config *config.LinkQuConfig
}

func NewLinkQuService() *LinkQuService {
	return &LinkQuService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config.GetLinkQuConfig(),
	}
}

// CreateQRIS creates a QRIS payment using LinkQu
func (lqs *LinkQuService) CreateQRIS(partnerRef, customerID, customerName string, amount int64, expired string, callbackURL string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/linkqu-partner/transaction/create/qris", lqs.config.BaseURL)

	payload := map[string]interface{}{
		"username":      lqs.config.Username,
		"pin":           lqs.config.PIN,
		"amount":        amount,
		"partner_reff":  partnerRef,
		"customer_id":   customerID,
		"customer_name": customerName,
		"expired":       expired,
		"url_callback":  callbackURL,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("client-id", lqs.config.ClientID)
	req.Header.Set("client-secret", lqs.config.ClientSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := lqs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateEWallet creates an E-Wallet payment using LinkQu
func (lqs *LinkQuService) CreateEWallet(partnerRef, customerID, customerName string, amount int64, expired string, retailCode, callbackURL string, ewalletPhone *string) (map[string]interface{}, error) {
	var url string
	payload := map[string]interface{}{
		"amount":        amount,
		"partner_reff":  partnerRef,
		"customer_id":   customerID,
		"customer_name": customerName,
		"expired":       expired,
		"username":      lqs.config.Username,
		"pin":           lqs.config.PIN,
		"retail_code":   retailCode,
		"url_callback":  callbackURL,
	}

	if retailCode == "PAYOVO" {
		url = fmt.Sprintf("%s/linkqu-partner/transaction/create/ovopush", lqs.config.BaseURL)
		if ewalletPhone != nil {
			payload["ewallet_phone"] = *ewalletPhone
		}
	} else {
		url = fmt.Sprintf("%s/linkqu-partner/transaction/create/paymentewallet", lqs.config.BaseURL)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("client-id", lqs.config.ClientID)
	req.Header.Set("client-secret", lqs.config.ClientSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := lqs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetBalance gets balance from LinkQu
func (lqs *LinkQuService) GetBalance() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/linkqu-partner/akun/resume?username=%s", lqs.config.BaseURL, lqs.config.Username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("client-id", lqs.config.ClientID)
	req.Header.Set("client-secret", lqs.config.ClientSecret)

	resp, err := lqs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

