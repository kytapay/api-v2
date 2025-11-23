package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatIDRAmount formats amount to IDR currency format
// Example: 1000000 -> "IDR 1.000.000"
func FormatIDRAmount(amount float64) string {
	// Convert to int to remove decimals
	amountInt := int64(amount)
	
	// Convert to string
	amountStr := strconv.FormatInt(amountInt, 10)
	
	// Add thousand separators
	var result strings.Builder
	length := len(amountStr)
	
	for i, char := range amountStr {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(char)
	}
	
	return fmt.Sprintf("IDR %s", result.String())
}

