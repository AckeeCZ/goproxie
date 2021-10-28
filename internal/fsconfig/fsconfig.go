package fsconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"time"
)

// MaxCommandsLength defines max stored commands threshold
const MaxCommandsLength = 100

type Config struct {
	History History `json:"history"`
}

type History struct {
	Commands []string `json:"commands"`
}

const configFileName = "store"

func getConfigDir() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	configDir := ""
	// Use `$XDG_CONFIG_HOME` or `~/.config` as config dir
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configDir = path.Join(xdgConfigHome, "goproxie")
	} else {
		configDir = path.Join(user.HomeDir, ".config", "goproxie")
	}
	return configDir
}

func getConfigPath() string {
	configDir := getConfigDir()
	configFile := fmt.Sprintf("%s.json", configFileName)
	file := path.Join(configDir, configFile)
	return file
}

func makeSureConfigDirectoryExists() {
	os.MkdirAll(getConfigDir(), os.ModePerm)
}

func configFileExists() bool {
	configFile := getConfigPath()
	// 💡 errors.Is to check error is of some type
	// Other way to check it is to cast it:
	// @example: if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {..}
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func CreateBackup() bool {
	if !configFileExists() {
		return false
	}
	file := readConfigFile()
	backupConfig := path.Join(getConfigDir(), fmt.Sprintf("%s-%d.json", configFileName, time.Now().Unix()))
	if err := os.WriteFile(backupConfig, file, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	return true
}

func readConfigFile() []byte {
	file, err := os.ReadFile(getConfigPath())
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func writeConfigToFile(c *Config) {
	data, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(getConfigPath(), data, os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func GetConfig() *Config {
	file := readConfigFile()
	config := &Config{}
	if err := json.Unmarshal(file, config); err != nil {
		log.Fatal(err)
	}
	return config
}

func AppendHistoryCommand(command string) {
	c := GetConfig()
	commands := append(c.History.Commands, command)
	if len(commands) > MaxCommandsLength {
		c.History.Commands = c.History.Commands[len(c.History.Commands)-MaxCommandsLength:]
	}
	c.History.Commands = deduplicate(commands)
	writeConfigToFile(c)
}

func deduplicate(strings []string) []string {
	stringToSelf := make(map[string]string)
	// Gotta have a separate struct for results to maintain ordering https://blog.golang.org/maps#TOC_7.
	uniqueStrings := []string{}
	for _, str := range strings {
		if stringToSelf[str] == "" {
			uniqueStrings = append(uniqueStrings, str)
			stringToSelf[str] = str
		}
	}
	return uniqueStrings
}

func Initialize() {
	makeSureConfigDirectoryExists()
	if !configFileExists() {
		writeConfigToFile(&Config{})
	}
}
