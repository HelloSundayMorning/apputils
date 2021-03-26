package app

type ApplicationID string


const (
	AppVersionEnv     = "APP_VERSION"
	AppEnvironmentEnv = "APP_ENVIRONMENT"
	AwsXrayDisableEnv = "AWS_XRAY_SDK_DISABLED"

	StagingEnvironment = "staging"
	ProductionEnvironment = "production"
	LocalEnvironment = "local"
)