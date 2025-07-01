package data

type Config struct {
	Db     DbConfig     `json:"db"`
	Server ServerConfig `json:"server"`
}

type SecretConfig struct {
	Password string `json:"password"`
}

type DbConfig struct {
	Machine  string `json:"machine"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type ServerConfig struct {
	Address       string `json:"address"`
	AuthServerUrl string `json:"authServerUrl"`
	Right         string `json:"right"`
	StorageFolder string `json:"storageFolder"`
	Context       string `json:"context"`
}
