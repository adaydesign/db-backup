package models

type ServerConfig struct {
	Name          string   `json:"name"`
	User          string   `json:"user"`
	Password      string   `json:"password"`
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	Database      string   `json:"database"`
	Tables        []string `json:"tables"`
	IgnoredTables []string `json:"ignored_tables"`
}

type ResultMessage struct {
	ServerName string
	Success    bool
	Message    string
}
