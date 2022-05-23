package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/interrupts"
	"github.com/opensourceways/community-robot-lib/kafka"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	"github.com/opensourceways/community-robot-lib/mq"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
)

const component = "robot-github-hook-delivery"

type options struct {
	service        liboptions.ServiceOptions
	hmacSecretFile string
	topic          string
}

func (o *options) Validate() error {
	if o.topic == "" {
		return fmt.Errorf("please set topic")
	}

	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.StringVar(&o.hmacSecretFile, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing the HMAC secret.")
	fs.StringVar(&o.topic, "topic", "", "The topic to which github webhook messages need to be published ")

	_ = fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(component)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	configAgent := config.NewConfigAgent(func() config.Config {
		return new(configuration)
	})
	if err := configAgent.Start(o.service.ConfigFile); err != nil {
		logrus.WithError(err).Fatal("Error starting config agent.")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.hmacSecretFile}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	getHmac := secretAgent.GetTokenGenerator(o.hmacSecretFile)

	d := delivery{hmac: getHmac, topic: o.topic}

	if err := initBroker(configAgent); err != nil {
		logrus.WithError(err).Fatal("Error init broker.")
	}

	defer interrupts.WaitForGracefulShutdown()
	interrupts.OnInterrupt(func() {
		configAgent.Stop()

		_ = kafka.Disconnect()

		d.wait()
	})

	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	// For /hook, handle a webhook normally.
	http.Handle("/github-hook", &d)

	httpServer := &http.Server{Addr: ":" + strconv.Itoa(o.service.Port)}

	interrupts.ListenAndServe(httpServer, o.service.GracePeriod)
}

func initBroker(agent config.ConfigAgent) error {
	cfg := &configuration{}
	_, c := agent.GetConfig()

	if v, ok := c.(*configuration); ok {
		cfg = v
	}

	tlsConfig, err := cfg.Config.TLSConfig.TLSConfig()
	if err != nil {
		return err
	}

	err = kafka.Init(
		mq.Addresses(cfg.Config.Addresses...),
		mq.SetTLSConfig(tlsConfig),
		mq.Log(logrus.WithField("module", "broker")),
	)

	if err != nil {
		return err
	}

	return kafka.Connect()
}
