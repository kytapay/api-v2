package config

import (
	"os"
)

type LinkQuConfig struct {
	BaseURL      string
	Username     string
	PIN          string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	CheckoutURL  string
}

var linkQuConfig *LinkQuConfig

func GetLinkQuConfig() *LinkQuConfig {
	if linkQuConfig == nil {
		linkQuConfig = &LinkQuConfig{
			BaseURL:     os.Getenv("LINKQU_BASE_URL"),
			Username:    os.Getenv("LINKQU_USERNAME"),
			PIN:         os.Getenv("LINKQU_PIN"),
			ClientID:    os.Getenv("LINKQU_CLIENT_ID"),
			ClientSecret: os.Getenv("LINKQU_CLIENT_SECRET"),
			CallbackURL: os.Getenv("LINKQU_CALLBACK_URL"),
			CheckoutURL: os.Getenv("CHECKOUT_URL"),
		}
	}
	return linkQuConfig
}

