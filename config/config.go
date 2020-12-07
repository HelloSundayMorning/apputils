package config

import "github.com/HelloSundayMorning/apputils/app"

type (
	Configuration struct {
		AppID      app.ApplicationID
		CorsConfig Cors
	}
)

func NewConfig(appID app.ApplicationID) (config *Configuration) {

	return &Configuration{
		AppID: appID,
	}
}
