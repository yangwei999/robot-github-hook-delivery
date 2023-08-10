package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/opensourceways/kafka-lib/mq"
)

var reIpPort = regexp.MustCompile(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}:[1-9][0-9]*$`)

type configuration struct {
	Address string `json:"address" required:"true"`
}

func (cfg *configuration) Validate() error {
	if r := cfg.parseAddress(); len(r) == 0 {
		return errors.New("invalid mq address")
	}

	return nil
}

func (cfg *configuration) SetDefault() {
}

func (cfg *configuration) mqConfig() mq.MQConfig {
	return mq.MQConfig{
		Addresses: cfg.parseAddress(),
	}
}

func (cfg *configuration) parseAddress() []string {
	v := strings.Split(cfg.Address, ",")
	r := make([]string, 0, len(v))
	for i := range v {
		if reIpPort.MatchString(v[i]) {
			r = append(r, v[i])
		}
	}

	return r
}
