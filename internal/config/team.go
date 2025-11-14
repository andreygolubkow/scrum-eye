package config

type AzureDevOpsTeam struct {
	Organisation string `yaml:"organization"`
	Token        string `yaml:"token"`
	ProjectId    string `yaml:"project"`
	TeamId       string `yaml:"team"`
	AreaPath     string `yaml:"area"`
}

type TeamConfig struct {
	AzureDevOps AzureDevOpsTeam `yaml:"azure"`
}
