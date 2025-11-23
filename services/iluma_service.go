package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type IlumaService struct {
	client *http.Client
	apiKey string
}

func NewIlumaService() *IlumaService {
	apiKey := os.Getenv("ILUMA_API_KEY")
	if apiKey == "" {
		apiKey = "iluma_live_uf4HfFeR46gQAutlQh1jEFINVwIq0bX1jrVkpOpm5kTiyiWc8LyFAQk33osL1t"
	}

	return &IlumaService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// ValidateBankAccount validates bank account using Iluma API
func (is *IlumaService) ValidateBankAccount(accountNumber, bankCode, referenceID string) (map[string]interface{}, error) {
	url := "https://api.iluma.ai/v1.2/identity/bank_account_validation_details"

	payload := map[string]interface{}{
		"bank_account_number": accountNumber,
		"bank_code":           fmt.Sprintf("ID_%s", bankCode),
		"reference_id":       referenceID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set Authorization header with Basic Auth
	auth := base64.StdEncoding.EncodeToString([]byte(is.apiKey + ":"))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	req.Header.Set("Content-Type", "application/json")

	resp, err := is.client.Do(req)
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

