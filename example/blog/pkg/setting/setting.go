package setting

import (
	"embed"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type Setting struct {
	vp *viper.Viper
}

func NewSetting(configDirs embed.FS) (*Setting, error) {
	d,err:=configDirs.ReadFile("configs/config.yaml")
	if os.IsNotExist(err) {
		return nil, err
	}
	vp := viper.New()
	/*vp.SetConfigName("config")
	vp.AddConfigPath("configs/")*/
	vp.SetConfigType("yaml")
	if os.IsNotExist(err) {
		return nil, err
	}
	err = vp.ReadConfig(strings.NewReader(string(d)))
	if err != nil {
		return nil, err
	}
	return &Setting{vp: vp}, nil
}
