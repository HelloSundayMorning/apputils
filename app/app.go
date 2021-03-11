package app

type ApplicationID string


const (
	AppVersionEnv = "APP_VERSION"
	AppEnvironmentEnv = "APP_ENVIRONMENT"

	StagingEnvironment = "staging"
	ProductionEnvironment = "production"
)