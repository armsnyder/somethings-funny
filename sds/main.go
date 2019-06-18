package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme"
	"log"
	"somethings-funny/sds/cache"
	"somethings-funny/sds/server"
)

type envSpec struct {
	Bucket           string `envconfig:"bucket" required:"true"`
	AwsRegion        string `envconfig:"aws_region"`
	AwsDefaultRegion string `envconfig:"aws_default_region"`
	Domain           string `envconfig:"domain" required:"true"`
	Staging          bool   `envconfig:"staging"`
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
	letsEncryptUrl := acme.LetsEncryptURL
	if env.Staging {
		letsEncryptUrl = "https://acme-staging.api.letsencrypt.org/directory"
	}
	server.Run(cache.New(&region, &env.Bucket, letsEncryptUrl), letsEncryptUrl, env.Domain)
}
