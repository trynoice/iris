package config

import "github.com/spf13/viper"

type Viper interface {
	SetDefault(key string, value interface{})
	ReadInConfig() error
	Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error
	SafeWriteConfig() error
}
