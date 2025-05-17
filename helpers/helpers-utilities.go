package helpers

import (
	"os"

)

const CONF_DIR = ".xampress"
const CONF_FILE = "configuration"
const CONF_EXT = "yaml"

func GetUserDir() string {
	userDir, _ := os.UserConfigDir()
	return userDir
}
