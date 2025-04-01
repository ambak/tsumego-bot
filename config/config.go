package config

type Config struct {
	Token       string `json:"token"`
	LogFile     string `json:"log_file"`
	Database    string `json:"database"`
	Tsumego     string `json:"tsumego"`
	OgsUsername string `json:"ogs_username"`
	OgsPassword string `json:"ogs_password"`
	OgsUrl      string `json:"ogs_url"`
}
