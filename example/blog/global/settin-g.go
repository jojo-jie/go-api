package global

import (
	"blog/pkg/logger"
	"blog/pkg/setting"
)

var (
	ServerSetting *setting.ServerSettingS
	AppSetting *setting.AppSettingS
	DatabaseSetting *setting.DataBaseSettingS
	JWTSetting *setting.JWTSettingS
	EmailSetting *setting.EmailSettingS
)

var (
	Logger *logger.Logger
)


