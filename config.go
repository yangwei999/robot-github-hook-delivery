package main

import (
	"github.com/opensourceways/kafka-lib/mq"
)

type configuration struct {
	Config mq.MQConfig `json:"config" required:"true"`
}

func (c *configuration) Validate() error {
	return nil
}

func (c *configuration) SetDefault() {
}
