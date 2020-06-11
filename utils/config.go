package utils

import (
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type config struct {
	Domains    []string `mapstructure:"Domains" validate:"required"`
	Subdomains []string `mapstructure:"Subdomains" validate:"required"`
	Token      string   `mapstructure:"Token" validate:"required"`
}

// Cfg unparsed string config file
var Cfg string

// Config parsed config file
var Config config

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	// initialize viper
	if Cfg != "" {
		// Use config file from the flag.
		viper.SetConfigFile(Cfg)

	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			Log.Fatal(err)
			os.Exit(1)
		}

		// enable config type as yml only
		viper.SetConfigType("yaml")

		// enable environment variables

		// set config and alternative config paths
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
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
		Log.WithField("component", "config").Debugln("Using config file:", viper.ConfigFileUsed())
	} else {
		Log.WithField("component", "config").Debugln("Can not find config file at known locations. Trying for environment variables.")
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		Log.WithField("component", "config").Fatalln("Unable to decode config file, %v", err)
	}

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
			Log.Errorln(err)
			return
		}

		for _, err := range err.(validator.ValidationErrors) {
			Log.Errorf("%s field can not be empty and has to be an %s.\n", err.Field(), err.Type())
		}

		os.Exit(127)
	}
}
