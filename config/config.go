package config

var SQConfig *SQConfigStruct

type SQConfigStruct struct {
	Socket        string `yaml:"socket"`
	Image         string `yaml:"image"`
	LogLevel      string `yaml:"logLevel"`
	StartupImages struct {
		A2SServer string `yaml:"a2sServer"`
		Promtail  string `yaml:"promtail"`
	} `yaml:"startupImages"`
	Volumes struct {
		Pter string `yaml:"pter"`
	} `yaml:"volumes"`
}
