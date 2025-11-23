package config

import (
	"os"
)

type PakaiLinkConfig struct {
	BaseURL      string
	PartnerID    string
	ChannelID    string
	ClientKey    string
	ClientSecret string
	AccountNo    string
	CallbackURL  string
}

var pakaiLinkConfig *PakaiLinkConfig

func GetPakaiLinkConfig() *PakaiLinkConfig {
	if pakaiLinkConfig == nil {
		pakaiLinkConfig = &PakaiLinkConfig{
			BaseURL:      os.Getenv("PAKAILINK_BASE_URL"),
			PartnerID:    os.Getenv("PAKAILINK_PARTNER_ID"),
			ChannelID:    os.Getenv("PAKAILINK_CHANNEL_ID"),
			ClientKey:    os.Getenv("PAKAILINK_CLIENT_KEY"),
			ClientSecret: os.Getenv("PAKAILINK_CLIENT_SECRET"),
			AccountNo:    os.Getenv("PAKAILINK_ACCOUNT_NO"),
			CallbackURL:  os.Getenv("PAKAILINK_CALLBACK_URL"),
		}
	}
	return pakaiLinkConfig
}

