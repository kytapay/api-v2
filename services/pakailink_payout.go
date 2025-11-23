package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kytapay/api-v2/helpers"
)

// BankAccountInquiry performs bank account inquiry for PakaiLink
func (pls *PakaiLinkService) BankAccountInquiry(partnerRef, accountNumber, bankCode string, amount float64) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snap/v1.0/emoney/bank-account-inquiry", pls.config.BaseURL)

	// Prepare request body
	payload := map[string]interface{}{
		"partnerReferenceNo":      partnerRef,
		"beneficiaryAccountNumber": accountNumber,
		"amount": map[string]interface{}{
			"value":    fmt.Sprintf("%.2f", amount),
			"currency": "IDR",
		},
		"additionalInfo": map[string]interface{}{
			"beneficiaryBankCode": bankCode,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/snap/v1.0/emoney/bank-account-inquiry"
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

// TransferBank performs bank transfer for PakaiLink
func (pls *PakaiLinkService) TransferBank(partnerRef, accountNumber, bankCode, sessionId string, amount float64, callbackURL, remark string) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snap/v1.0/emoney/transfer-bank", pls.config.BaseURL)

	// Prepare request body
	additionalInfo := map[string]interface{}{}
	if callbackURL != "" {
		additionalInfo["callbackUrl"] = callbackURL
	}
	if remark != "" {
		additionalInfo["remark"] = remark
	}

	payload := map[string]interface{}{
		"partnerReferenceNo":      partnerRef,
		"beneficiaryAccountNumber": accountNumber,
		"beneficiaryBankCode":      bankCode,
		"sessionId":                sessionId,
		"amount": map[string]interface{}{
			"value":    fmt.Sprintf("%.2f", amount),
			"currency": "IDR",
		},
		"additionalInfo": additionalInfo,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/snap/v1.0/emoney/transfer-bank"
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

// EWalletAccountInquiry performs e-wallet account inquiry for PakaiLink
func (pls *PakaiLinkService) EWalletAccountInquiry(partnerRef, customerNumber, productCode string, amount float64) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snap/v1.0/emoney/account-inquiry", pls.config.BaseURL)

	// Prepare request body
	payload := map[string]interface{}{
		"partnerReferenceNo": partnerRef,
		"customerNumber":     customerNumber,
		"amount": map[string]interface{}{
			"value":    fmt.Sprintf("%.2f", amount),
			"currency": "IDR",
		},
		"additionalInfo": map[string]interface{}{
			"productCode": productCode,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/snap/v1.0/emoney/account-inquiry"
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

// TopupEWallet performs e-wallet topup for PakaiLink
func (pls *PakaiLinkService) TopupEWallet(partnerRef, customerNumber, productCode, sessionId string, amount float64, callbackURL string) (map[string]interface{}, error) {
	// Get access token first
	accessToken, err := pls.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/snap/v1.0/emoney/topup", pls.config.BaseURL)

	// Prepare request body
	additionalInfo := map[string]interface{}{}
	if callbackURL != "" {
		additionalInfo["callbackUrl"] = callbackURL
	}

	payload := map[string]interface{}{
		"partnerReferenceNo": partnerRef,
		"customerNumber":     customerNumber,
		"productCode":         productCode,
		"sessionId":           sessionId,
		"amount": map[string]interface{}{
			"value":    fmt.Sprintf("%.2f", amount),
			"currency": "IDR",
		},
		"additionalInfo": additionalInfo,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Generate symmetric signature
	path := "/snap/v1.0/emoney/topup"
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

