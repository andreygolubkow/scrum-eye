package config

type AzureDevOpsConfig struct {
	Organization string `yaml:"organization"`
	Token        string `yaml:"token"`
}

type GlobalConfig struct {
	AzureDevOps AzureDevOpsConfig `yaml:"azure"`
}
