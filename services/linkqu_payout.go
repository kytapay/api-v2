package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// BankWithdrawInquiry performs bank withdraw inquiry for LinkQu
func (lqs *LinkQuService) BankWithdrawInquiry(bankCode, accountNumber string, amount int64, partnerRef string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/linkqu-partner/transaction/withdraw/inquiry", lqs.config.BaseURL)

	payload := map[string]interface{}{
		"username":      lqs.config.Username,
		"pin":           lqs.config.PIN,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        amount,
		"partner_reff":  partnerRef,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Log request to LinkQu
	log.Printf("[LINKQU BANK WITHDRAW INQUIRY] Request URL: %s", url)
	log.Printf("[LINKQU BANK WITHDRAW INQUIRY] Request body: %s", string(jsonData))

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

// BankWithdrawPayment performs bank withdraw payment for LinkQu
func (lqs *LinkQuService) BankWithdrawPayment(bankCode, accountNumber string, amount int64, partnerRef, inquiryRef, callbackURL string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/linkqu-partner/transaction/withdraw/payment", lqs.config.BaseURL)

	payload := map[string]interface{}{
		"username":      lqs.config.Username,
		"pin":           lqs.config.PIN,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        amount,
		"partner_reff":  partnerRef,
		"inquiry_reff":  inquiryRef,
	}

	if callbackURL != "" {
		payload["url_callback"] = callbackURL
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Log request to LinkQu
	log.Printf("[LINKQU BANK WITHDRAW PAYMENT] Request URL: %s", url)
	log.Printf("[LINKQU BANK WITHDRAW PAYMENT] Request body: %s", string(jsonData))

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

// EWalletReloadInquiry performs e-wallet reload inquiry for LinkQu
func (lqs *LinkQuService) EWalletReloadInquiry(bankCode, accountNumber string, amount int64) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/linkqu-partner/transaction/reload/inquiry", lqs.config.BaseURL)

	payload := map[string]interface{}{
		"username":      lqs.config.Username,
		"pin":           lqs.config.PIN,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        amount,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Log request to LinkQu
	log.Printf("[LINKQU E-WALLET RELOAD INQUIRY] Request URL: %s", url)
	log.Printf("[LINKQU E-WALLET RELOAD INQUIRY] Request body: %s", string(jsonData))

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

// EWalletReloadPayment performs e-wallet reload payment for LinkQu
func (lqs *LinkQuService) EWalletReloadPayment(bankCode, accountNumber string, amount int64, partnerRef, inquiryRef, callbackURL string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/linkqu-partner/transaction/reload/payment", lqs.config.BaseURL)

	payload := map[string]interface{}{
		"username":      lqs.config.Username,
		"pin":           lqs.config.PIN,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        amount,
		"partner_reff":  partnerRef,
		"inquiry_reff":  inquiryRef,
	}

	if callbackURL != "" {
		payload["url_callback"] = callbackURL
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Log request to LinkQu
	log.Printf("[LINKQU E-WALLET RELOAD PAYMENT] Request URL: %s", url)
	log.Printf("[LINKQU E-WALLET RELOAD PAYMENT] Request body: %s", string(jsonData))

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

