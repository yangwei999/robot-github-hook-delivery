package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/opensourceways/kafka-lib/mq"
)

var reIpPort = regexp.MustCompile(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}:[1-9][0-9]*$`)

type Config struct {
	Address string `json:"address" required:"true"`
}

func (cfg *Config) Validate() error {
	if r := cfg.parseAddress(); len(r) == 0 {
		return errors.New("invalid mq address")
	}

	return nil
}

func (cfg *Config) SetDefault() {
}

func (cfg *Config) mqConfig() mq.MQConfig {
	return mq.MQConfig{
		Addresses: cfg.parseAddress(),
	}
}

func (cfg *Config) parseAddress() []string {
	v := strings.Split(cfg.Address, ",")
	r := make([]string, 0, len(v))
	for i := range v {
		if reIpPort.MatchString(v[i]) {
			r = append(r, v[i])
		}
	}

	return r
}
