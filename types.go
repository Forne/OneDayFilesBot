package main

// Configuration is a main config
type Configuration struct {
	Telegram Telegram `json:"telegram"`
	Swift    Swift    `json:"swift"`
}

// Telegram API settings
type Telegram struct {
	Token   string  `json:"token"`
	Webhook Webhook `json:"webhook"`
}

type Webhook struct {
	Set    string `json:"set"`
	Listen string `json:"listen"`
	Serve  string `json:"serve"`
}

type Swift struct {
	UserName    string
	ApiKey      string
	AuthUrl     string
	FrontendUrl string
	Container   string
	PathToFile  string
}
