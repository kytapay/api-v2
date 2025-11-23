package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kytapay/api-v2/config"
	"github.com/kytapay/api-v2/helpers"
)

type PakaiLinkService struct {
	client *http.Client
	config *config.PakaiLinkConfig
}

func NewPakaiLinkService() *PakaiLinkService {
	return &PakaiLinkService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config.GetPakaiLinkConfig(),
	}
}

// GetAccessToken gets B2B access token from PakaiLink
func (pls *PakaiLinkService) GetAccessToken() (string, error) {
	url := fmt.Sprintf("%s/snap/v1.0/access-token/b2b", pls.config.BaseURL)

	// Generate asymmetric signature
	signature, timestamp, err := helpers.GenerateAsymmetricSignature(pls.config.ClientKey)
	if err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"grantType": "client_credentials",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-CLIENT-KEY", pls.config.ClientKey)
	req.Header.Set("X-SIGNATURE", signature)

	resp, err := pls.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check response code
	responseCode, _ := result["responseCode"].(string)
	if responseCode != "2007300" {
		responseMessage, _ := result["responseMessage"].(string)
		return "", fmt.Errorf("failed to get access token: %s - %s", responseCode, responseMessage)
	}

	accessToken, _ := result["accessToken"].(string)
	return accessToken, nil
}

// CreateVA creates a Virtual Account using PakaiLink
func (pls *PakaiLinkService) CreateVA(partnerRef, customerNo, virtualAccountName string, amount float64, expiredDate, bankCode, callbackURL string) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snap/v1.0/transfer-va/create-va", pls.config.BaseURL)

	// Prepare request body
	payload := map[string]interface{}{
		"partnerReferenceNo": partnerRef,
		"customerNo":         customerNo,
		"virtualAccountName": virtualAccountName,
		"totalAmount": map[string]interface{}{
			"value":    fmt.Sprintf("%.2f", amount),
			"currency": "IDR",
		},
		"additionalInfo": map[string]interface{}{
			"callbackUrl": callbackURL,
			"bankCode":    bankCode,
		},
	}

	if expiredDate != "" {
		payload["expiredDate"] = expiredDate
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/snap/v1.0/transfer-va/create-va"
	signature, err := helpers.GenerateSymmetricSignature("POST", path, accessToken, string(jsonData), timestamp, pls.config.ClientSecret)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", pls.config.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", partnerRef)
	req.Header.Set("CHANNEL-ID", pls.config.ChannelID)
	req.Header.Set("X-SIGNATURE", signature)

	resp, err := pls.client.Do(req)
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

// CreatePayout creates a payout/transfer using PakaiLink
func (pls *PakaiLinkService) CreatePayout(partnerRef, customerNo, accountName, accountNumber, bankCode string, amount float64, description string) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/pakailink/payout/create", pls.config.BaseURL)

	// Prepare request body
	payload := map[string]interface{}{
		"external_id":        partnerRef,
		"bank_code":          bankCode,
		"account_holder_name": accountName,
		"account_number":     accountNumber,
		"amount":             amount,
		"description":        description,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/pakailink/payout/create"
	signature, err := helpers.GenerateSymmetricSignature("POST", path, accessToken, string(jsonData), timestamp, pls.config.ClientSecret)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", pls.config.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", partnerRef)
	req.Header.Set("CHANNEL-ID", pls.config.ChannelID)
	req.Header.Set("X-SIGNATURE", signature)

	resp, err := pls.client.Do(req)
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

// GetBalance gets balance from PakaiLink
func (pls *PakaiLinkService) GetBalance(partnerRef, accountNo string) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snap/v1.0/balance-inquiry", pls.config.BaseURL)

	// Prepare request body
	payload := map[string]interface{}{
		"partnerReferenceNo": partnerRef,
		"balanceTypes":       []string{"Balance"},
	}

	if accountNo != "" {
		payload["accountNo"] = accountNo
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/snap/v1.0/balance-inquiry"
	signature, err := helpers.GenerateSymmetricSignature("POST", path, accessToken, string(jsonData), timestamp, pls.config.ClientSecret)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", pls.config.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", partnerRef)
	req.Header.Set("CHANNEL-ID", pls.config.ChannelID)
	req.Header.Set("X-SIGNATURE", signature)

	resp, err := pls.client.Do(req)
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

