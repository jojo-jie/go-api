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
)

var (
	Logger *logger.Logger
)


