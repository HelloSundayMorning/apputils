package app

type ApplicationID string


const (
	AppVersionEnv     = "APP_VERSION"
	AppEnvironmentEnv = "APP_ENVIRONMENT"
	AwsXrayDisableEnv = "AWS_XRAY_SDK_DISABLED"
	AwsXrayHostEnv = "AWS_XRAY_HOST"

	StagingEnvironment = "staging"
	ProductionEnvironment = "production"
	LocalEnvironment = "local"
)