package store

import (
	"log"
	"os"
	"os/user"
	"path"

	"github.com/spf13/viper"
)

const MaxAppendLength = 100

func Initialize() {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	viper.SetConfigType("json")
	configPath := ""
	// Use `$XDG_CONFIG_HOME` or `~/.config` as config dir
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configPath = path.Join(xdgConfigHome, "goproxie")
	} else {
		configPath = path.Join(user.HomeDir, ".config", "goproxie")
	}
	configFile := "store"

	// Make sure the dir structure exist
	os.MkdirAll(configPath, os.ModePerm)

	// Configure viper to use the settings
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configFile)

	err = viper.SafeWriteConfig()
	if err != nil {
		// If file already exists, its fine
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			log.Fatal(err)
		}
	}
	viper.ReadInConfig()
}

func Set(key string, value interface{}) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

func Get(key string) interface{} {
	return viper.Get(key)
}

func Append(key string, value interface{}) error {
	currentValue := []interface{}{}
	if viper.IsSet(key) {
		currentValue = viper.Get(key).([]interface{})
	}
	currentValue = append(currentValue, value)
	if len(currentValue) > MaxAppendLength {
		currentValue = currentValue[len(currentValue)-MaxAppendLength:]
	}
	return Set(key, currentValue)
}
