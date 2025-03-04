// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package main

import (
	"io"
	"net/http"

	"github.com/drone/drone-go/plugin/converter"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/meltwater/drone-convert-pathschanged/plugin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// spec provides the plugin settings.
type spec struct {
	Bind   string `envconfig:"DRONE_BIND"`
	Debug  bool   `envconfig:"DRONE_DEBUG"`
	Text   bool   `envconfig:"DRONE_LOGS_TEXT"`
	Secret string `envconfig:"DRONE_SECRET"`

	Provider          string `envconfig:"PROVIDER"`
	Token             string `envconfig:"TOKEN"`
	BitBucketAddress  string `envconfig:"BB_ADDRESS"`
	BitBucketUser     string `envconfig:"BITBUCKET_USER"`
	BitBucketPassword string `envconfig:"BITBUCKET_PASSWORD"`
	GithubServer      string `envconfig:"GITHUB_SERVER"`
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func main() {
	spec := new(spec)
	err := envconfig.Process("", spec)
	if err != nil {
		logrus.Fatal(err)
	}

	if spec.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if spec.Text {
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	if spec.Secret == "" {
		logrus.Fatalln("missing secret key")
	}
	if spec.Provider == "" {
		logrus.Fatalln("missing provider")
	} else {
		providers := []string{"bitbucket", "bitbucket-server", "github"}
		if !contains(providers, spec.Provider) {
			logrus.Fatalln("invalid provider:", spec.Provider)
		}
	}
	if spec.Token == "" && (spec.Provider == "github" || spec.Provider == "bitbucket-server") {
		logrus.Fatalln("missing token")
	}
	if spec.BitBucketUser == "" && spec.Provider == "bitbucket" {
		logrus.Fatalln("missing bitbucket user")
	}
	if spec.BitBucketPassword == "" && spec.Provider == "bitbucket" {
		logrus.Fatalln("missing bitbucket password")
	}
	if spec.BitBucketAddress == "" && spec.Provider == "bitbucket-server" {
		logrus.Fatalln("missing bitbucket server address")
	}
	if spec.Bind == "" {
		spec.Bind = ":3000"
	}

	handler := converter.Handler(
		plugin.New(
			spec.Token,
			spec.Provider,
			spec.GithubServer,
			spec.BitBucketUser,
			spec.BitBucketPassword,
		),
		spec.Secret,
		logrus.StandardLogger(),
	)

	logrus.Infof("server listening on address %s", spec.Bind)

	http.Handle("/", handler)
	http.HandleFunc("/healthz", healthz)
	http.Handle("/metrics", promhttp.Handler())
	logrus.Fatal(http.ListenAndServe(spec.Bind, nil))
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "OK")
}
