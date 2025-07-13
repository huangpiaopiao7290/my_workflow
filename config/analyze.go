package config

import (
	"context"
	"fmt"
	"my_workflow/pkg/logger"
	"os"
	"sync"

	"github.com/spf13/viper"
)

// 加载配置

var (
	EnvConfigRootDir = "/config"				// 配置文件根目录
	BaseConfigPath = "/config/config.yaml"		// 基础配置文件路径
	configMutext sync.RWMutex
)
func init() {
	if err := LoadConfig(); err != nil {
		panic(fmt.Errorf("fatal error init config: %w", err))
	}
}

func LoadConfig() error {
	configMutext.Lock()
	defer configMutext.Unlock()

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	viper.SetConfigFile(cwd + BaseConfigPath)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

// 获取配置项字符串类型
// 参数：
// 	 - key 配置项的key
// 返回值：
// 	 - string 配置项的值
func GetString(key string) string {
	configValue := viper.GetString(key)
	if configValue == "" {
		logger.Error(context.Background(), "config not found", key)
	}
	return configValue
}

func GetInt(key string) int {
	configValue := viper.GetInt(key)
	if configValue == 0 {
		logger.Error(context.Background(), "config not found", key)
	}
	return configValue
}