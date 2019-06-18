package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"log"
	"somethings-funny/service/boot"
	"somethings-funny/service/route"
)

type envSpec struct {
	HostedZone       string `envconfig:"hosted_zone" required:"true"`
	ServiceDomain    string `envconfig:"service_domain" required:"true"`
	PiDomain         string `envconfig:"pi_domain" required:"true"`
	Bucket           string `envconfig:"bucket" required:"true"`
	AwsRegion        string `envconfig:"aws_region"`
	AwsDefaultRegion string `envconfig:"aws_default_region"`
}

var env envSpec

func init() {
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	region := env.AwsRegion
	if region == "" {
		region = env.AwsDefaultRegion
	}
	if region == "" {
		logrus.Fatal("no region specified")
	}
	if err := boot.RegisterOwnDNS(&env.ServiceDomain, &env.HostedZone); err != nil {
		logrus.Fatal(err.Error())
	}
	router := route.NewRouter(&env.PiDomain, &env.HostedZone, &region, &env.Bucket)
	if err := router.Run(":8080"); err != nil {
		logrus.Fatal(err.Error())
	}
}
