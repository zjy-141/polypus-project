package config

import (
	"polypus-project/service/validator"
)

func init() {
	initConfig()
	initLogger()
	validator.InitValidator(Config.AppLanguage)
}
