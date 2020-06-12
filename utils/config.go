package utils

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type config struct {
	Domains    []string `mapstructure:"domains" validate:"required,unique"`
	Subdomains []string `mapstructure:"subdomains" validate:"required,unique"`
	Token      string   `mapstructure:"token" validate:"required"`
	Interval   int      `mapstructure:"repeat"`
}

// Cfg unparsed string config file
var Cfg string

// Config parsed config file
var Config config

// InitConfig initialize the config variable
func InitConfig() {
	// initialize viper
	if Cfg != "" {
		// Use config file from the flag.
		viper.SetConfigFile(Cfg)

	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			Log.WithField("component", "CONFIG").Fatal(err)
			os.Exit(1)
		}

		// enable config type as yml only
		viper.SetConfigType("yaml")

		// enable environment variables

		// set config and alternative config paths
		viper.AddConfigPath(".")
		viper.AddConfigPath(fmt.Sprintf("%s/.config/do-dyndns/", home))
		viper.AddConfigPath("/etc/do-dyndns/")

		// Search config in given directories with name ".do-dyndns" (without extension).
		viper.SetConfigName(".do-dyndns")
	}

	// set environment variables
	viper.SetEnvPrefix("dyndns")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Log.WithField("component", "CONFIG").Debugln("Using config file:", viper.ConfigFileUsed())
	} else {
		Log.WithField("component", "CONFIG").Debugln("Can not find config file at known locations. Trying for environment variables.")
	}
}

// LoadConfig reads in config file and ENV variables if set.
func LoadConfig() {
	err := viper.Unmarshal(&Config)
	if err != nil {
		Log.WithField("component", "CONFIG").Fatalln("Unable to decode config file, %v", err)
	}

	// set default values
	viper.SetDefault("Repeat", 3600)

	// bind environments and get the config into struct
	bindEnvs(viper.GetViper(), Config)
	viper.Unmarshal(&Config)

	// check struct
	checkConfig(Config)
}

func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		fieldv := ifv.Field(i)
		t := ift.Field(i)
		name := strings.ToLower(t.Name)
		tag, ok := t.Tag.Lookup("mapstructure")
		if ok {
			name = tag
		}
		path := append(parts, name)
		switch fieldv.Kind() {
		case reflect.Struct:
			bindEnvs(v, fieldv.Interface(), path...)
		default:
			v.BindEnv(strings.Join(path, "."))
		}
	}
}

func checkConfig(iface config) {
	validate := validator.New()
	err := validate.Struct(iface)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			Log.WithField("component", "CONFIG").Errorln(err)
			return
		}

		for _, err := range err.(validator.ValidationErrors) {
			var parsedError string

			if err.ActualTag() == "required" {
				parsedError = "is required"
			} else if err.ActualTag() == "unique" {
				parsedError = "has to be unique"
			} else {
				parsedError = fmt.Sprintf("has to be %s", err.ActualTag())
			}

			Log.WithField("component", "CONFIG").Errorf("%s %s and has to be %s.\n", err.Field(), parsedError, err.Type())
		}

		os.Exit(127)
	}
}
