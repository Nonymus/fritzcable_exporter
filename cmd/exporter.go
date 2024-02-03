package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"fritzcable_exporter/internal/adapter"
	"fritzcable_exporter/internal/client"
)

var (
	flagHost         = flag.String("host", "fritz.box", "hostname of IP address of target device")
	flagUsername     = flag.String("username", "", "username for login (if any)")
	flagPassword     = flag.String("password", "", "password for device login")
	flagPasswordFile = flag.String("passwordFile", "", "path to file containing the password")
	flagListen       = flag.String("listen", ":8080", "net/http listen string")
)

func main() {
	flag.Parse()
	if *flagPassword == "" && *flagPasswordFile == "" {
		log.Fatal("either password or passwordFile must be provided")
	}
	password := *flagPassword
	if *flagPasswordFile != "" {
		pwBytes, err := os.ReadFile(*flagPasswordFile)
		if err != nil {
			log.Fatal("failed to read password from file: ", err)
		}
		password = strings.TrimRight(string(pwBytes), "\n")
	}

	c := client.NewClient(*flagHost, *flagUsername, password)
	if err := c.Login(); err != nil {
		log.Fatal(err)
	}

	reg := prometheus.NewPedanticRegistry()

	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
		collectors.NewGoCollector(),
	)

	bc := adapter.BoxCollector{Client: c}
	reg.MustRegister(bc)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Printf("starting server on %s/metrics", *flagListen)
	log.Fatal(http.ListenAndServe(*flagListen, nil))
}
