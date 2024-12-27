package main

import (
	"flag"
	"github.com/hazcod/sentinelpurger/config"
	msSentinel "github.com/hazcod/sentinelpurger/pkg/sentinel"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	//ctx := context.Background()

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	confFile := flag.String("config", "config.yml", "The YAML configuration file.")
	flag.Parse()

	conf := config.Config{}
	if err := conf.Load(*confFile); err != nil {
		logger.WithError(err).WithField("config", *confFile).Fatal("failed to load configuration")
	}

	if err := conf.Validate(); err != nil {
		logger.WithError(err).WithField("config", *confFile).Fatal("invalid configuration")
	}

	logrusLevel, err := logrus.ParseLevel(conf.Log.Level)
	if err != nil {
		logger.WithError(err).Error("invalid log level provided")
		logrusLevel = logrus.InfoLevel
	}
	logger.SetLevel(logrusLevel)

	//

	sentinel, err := msSentinel.New(logger, msSentinel.Credentials{
		TenantID:       conf.Microsoft.TenantID,
		ClientID:       conf.Microsoft.AppID,
		ClientSecret:   conf.Microsoft.SecretKey,
		SubscriptionID: conf.Microsoft.SubscriptionID,
		ResourceGroup:  conf.Microsoft.ResourceGroup,
		WorkspaceName:  conf.Microsoft.WorkspaceName,
	})
	if err != nil {
		logger.WithError(err).Fatal("could not create audit MS Sentinel client")
	}

	//
	now := time.Now()

	for _, entry := range conf.Tables {
		tableLogger := logger.WithFields(logrus.Fields{"table": entry.Name})

		duration, err := time.ParseDuration(entry.Retention)
		if err != nil {
			tableLogger.WithError(err).Fatal("could not parse duration")
		}

		tresholdDate := now.Add(duration * -1)

		tableLogger.WithField("treshold", tresholdDate.Format("2006-01-02")).Info("found table")

		if err := sentinel.PurgeLogs(
			tableLogger, conf.Microsoft.SubscriptionID, conf.Microsoft.ResourceGroup, conf.Microsoft.WorkspaceName,
			entry.Name, tresholdDate); err != nil {
			tableLogger.WithError(err).Fatal("failed to purge logs")
		}
	}

	logger.Info("finished ingesting")
}
