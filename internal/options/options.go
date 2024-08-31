package options

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
)

const configDir = ".config/port-jump"
const configFile = "config.yml"

type Options struct {
	LogDebug bool

	Jumps []*PortJump `mapstructure:"jumps"`
}

type PortJump struct {
	Enabled      bool   `mapstructure:"enabled"`
	DstPort      int    `mapstructure:"dstport"`
	Interval     int64  `mapstructure:"interval"`
	SharedSecret string `mapstructure:"sharedsecret"`
}

// NewOptions returns fresh Options
func NewOptions() *Options {
	return &Options{}
}

// NewPortJump returns a new port jumping configuration
func NewPortJump(dst int, secret string, interval int64, enabled bool) (*PortJump, error) {
	if dst == 0 {
		return nil, errors.New("dst cant be 0")
	}

	if secret == "" {
		return nil, errors.New("secret cant be empty")
	}

	if interval == 0 {
		return nil, errors.New("interval has to be more than 0")
	}

	return &PortJump{
		Enabled:      enabled,
		DstPort:      dst,
		SharedSecret: secret,
		Interval:     interval,
	}, nil
}

// configPath returns the path where configuration should live.
// if the config file does not exist, it will be created.
func (o *Options) configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	p := filepath.Join(home, configDir, configFile)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return "", fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	return p, nil
}

// Load loads configuration from the config file
func (o *Options) Load() error {
	configPath, err := o.configPath()
	if err != nil {
		return err
	}

	viper.SetConfigFile(configPath)

	// Check if the file exists before trying to read it
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If the config file does not exist, just return without an error
		// as this is expected on the first run.
		return nil
	}

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		// Handle other potential errors reading the config file
		return fmt.Errorf("failed to read config file: %v", err)
	}

	if err := viper.Unmarshal(o); err != nil {
		return fmt.Errorf("failed to unmarshal config into options struct: %v", err)
	}

	return nil
}

// Save saves configuration to the config file
func (o *Options) Save() error {
	configPath, err := o.configPath()
	if err != nil {
		return err
	}

	// Use reflection to loop through the struct fields and set them in Viper
	val := reflect.ValueOf(o).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Tag.Get("mapstructure")
		fieldValue := val.Field(i).Interface()

		if fieldName != "" {
			viper.Set(fieldName, fieldValue)
		}
	}

	// Write the options to the YAML file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	// Ensure file permissions are set after writing the file
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %v", err)
	}

	return nil
}
