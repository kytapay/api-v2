package helpers

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

// GenerateAsymmetricSignature generates asymmetric signature for PakaiLink
func GenerateAsymmetricSignature(clientKey string) (string, string, error) {
	// Read private key from file
	privateKeyPath := "pkcs8_rsa_private_key.pem"
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private key: %v", err)
	}

	// Parse private key
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return "", "", fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key: %v", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", "", fmt.Errorf("not an RSA private key")
	}

	// Generate timestamp
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone data not available
		loc = time.UTC
	}
	timestamp := time.Now().In(loc).Format("2006-01-02T15:04:05+07:00")

	// Create string to sign: X-CLIENT-KEY|X-TIMESTAMP
	stringToSign := fmt.Sprintf("%s|%s", clientKey, timestamp)

	// Sign using SHA256withRSA
	hashed := sha256.Sum256([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", "", fmt.Errorf("failed to sign: %v", err)
	}

	// Encode to Base64
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	return signatureBase64, timestamp, nil
}

// GenerateSymmetricSignature generates symmetric signature for PakaiLink
func GenerateSymmetricSignature(method, path, accessToken, requestBody, timestamp, clientSecret string) (string, error) {
	// Minify request body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(requestBody), &bodyMap); err != nil {
		return "", fmt.Errorf("failed to parse request body: %v", err)
	}

	minifiedBody, err := json.Marshal(bodyMap)
	if err != nil {
		return "", fmt.Errorf("failed to minify body: %v", err)
	}

	// SHA-256 hash of minified body
	hash := sha256.Sum256(minifiedBody)
	hashHex := strings.ToLower(hex.EncodeToString(hash[:]))

	// Compose string to sign
	stringToSign := fmt.Sprintf("%s:%s:%s:%s:%s", method, path, accessToken, hashHex, timestamp)

	// HMAC SHA-512
	mac := hmac.New(sha512.New, []byte(clientSecret))
	mac.Write([]byte(stringToSign))
	hmacHash := mac.Sum(nil)

	// Encode to Base64
	signature := base64.StdEncoding.EncodeToString(hmacHash)

	return signature, nil
}

// GetBankCode converts bank name to PakaiLink bank code
func GetBankCode(bankName string) string {
	bankCodes := map[string]string{
		"BCA":      "014",
		"BJB":      "110",
		"BNI":      "009",
		"BRI":      "002",
		"BSI":      "451",
		"BSS":      "523",
		"CIMB":     "022",
		"MANDIRI":  "008",
		"PERMATA":  "013",
		"BTN":      "200",
		"DANAMON":  "011",
		"OCBC":     "028",
		"SEABANK":  "535",
		"ALLO":     "567",
		"BNC":      "490",
		"JAGO":     "542",
		"MEGA":     "426",
		"BUKOPIN":  "441",
	}

	code, exists := bankCodes[strings.ToUpper(bankName)]
	if !exists {
		return "002" // Default to BRI
	}
	return code
}

// GetPakaiLinkBankCode converts bank name to PakaiLink bank code (same as GetBankCode)
func GetPakaiLinkBankCode(bankName string) string {
	return GetBankCode(bankName)
}

// GetLinkQuBankCode converts bank name to LinkQu bank code
func GetLinkQuBankCode(bankName string) string {
	bankCodes := map[string]string{
		"BCA":      "014",
		"BJB":      "110",
		"BNI":      "009",
		"BRI":      "002",
		"BSI":      "451",
		"BSS":      "523",
		"CIMB":     "022",
		"MANDIRI":  "008",
		"PERMATA":  "013",
		"BTN":      "200",
		"DANAMON":  "011",
		"OCBC":     "028",
		"SEABANK":  "535",
		"ALLO":     "567",
		"BNC":      "490",
		"JAGO":     "542",
		"MEGA":     "426",
		"BUKOPIN":  "441",
	}

	code, exists := bankCodes[strings.ToUpper(bankName)]
	if !exists {
		return "002" // Default to BRI
	}
	return code
}

// GetLinkQuEWalletCode converts e-wallet name to LinkQu bankcode
func GetLinkQuEWalletCode(ewalletName string) string {
	ewalletCodes := map[string]string{
		"DANA":      "DANA",
		"OVO":       "OVO",
		"GOPAY":     "GOPAY",
		"LINKAJA":   "LINKAJA",
		"SHOPEEPAY": "SHOPEEPAY",
		"KASPRO":    "KASPRO",
	}

	code, exists := ewalletCodes[strings.ToUpper(ewalletName)]
	if !exists {
		return strings.ToUpper(ewalletName) // Return uppercase if not found
	}
	return code
}

// GetPakaiLinkProductCode converts e-wallet name to PakaiLink productCode
func GetPakaiLinkProductCode(ewalletName string) string {
	productCodes := map[string]string{
		"DANA":      "DANA",
		"OVO":       "OVO",
		"GOPAY":     "GOPAY",
		"LINKAJA":   "LINKAJA",
		"SHOPEEPAY": "SHOPEEPAY",
	}

	code, exists := productCodes[strings.ToUpper(ewalletName)]
	if !exists {
		return strings.ToUpper(ewalletName) // Return uppercase if not found
	}
	return code
}

