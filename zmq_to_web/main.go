package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

var zmqPollTimeout = time.Second * 5
var updateSupportedTypesInterval = 60 * time.Minute

func isEqualStringSlices(a, b []string) bool {
	if (a == nil) && (b == nil) {
		return true
	}

	if (a == nil) || (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func updateSupportedTypes(api *WebApi, sleepInterval time.Duration) {
	for {
		time.Sleep(sleepInterval)

		origTypes := api.SupportedTypes()
		err := api.UpdateMtypeIds()
		if err != nil {
			log.Printf("Unable to update supported types: %v", err)
			continue
		}

		newTypes := api.SupportedTypes()

		if !isEqualStringSlices(origTypes, newTypes) {
			log.Printf("The list of supported types was updated to: %s",
				strings.Join(newTypes, ", "))
		}
	}
}

func main() {
	ret := 1
	defer func() {
		os.Exit(ret)
	}()

	debugMode := flag.Bool("debug", false, "Enable debug output")
	zmqEndpoint := flag.String("endpoint", "", "ZMQ endpoint (tcp://1.2.3.4:5555)")
	webApiBaseUrl := flag.String("apiurl", "", "The web API url (http://1.2.3.4/api)")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Reads measurements from the ZMQ endpoint and posts them to the HTTP API.

Usage: %s --endpoint "..." --apiurl "..." [ --debug ]

`, os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "hello\n\n")
		flag.PrintDefaults()
	}

	if (*zmqEndpoint == "") || (*webApiBaseUrl == "") {
		flag.Usage()
		return
	}

	subscriber, err := NewSubscriber(*zmqEndpoint)
	if err != nil {
		log.Printf("NewSubscriber() failed: %v", err)
		return
	}
	defer subscriber.Destroy()
	log.Printf("Connected to %s", subscriber.Endpoint)

	api, err := NewWebApi(*webApiBaseUrl)
	if err != nil {
		log.Printf("NewWebApi() failed: %v", err)
		return
	}
	log.Printf("Posting messages to %s", api.BaseUrl)
	log.Printf("Supported measurement types: %s", strings.Join(api.SupportedTypes(), ", "))

	go updateSupportedTypes(api, updateSupportedTypesInterval)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	ret = 0
	for {
		if len(stopChan) != 0 {
			break
		}

		measurements, err := subscriber.RecvMeasurement(zmqPollTimeout)
		if err != nil {
			log.Printf("RecvMeasurement() failed: %v", err)
			continue
		}

		if len(measurements) == 0 {
			continue
		}

		for _, m := range measurements {
			if *debugMode {
				log.Printf("Received %#v", *m)
			}

			err = api.PublishMeasurement(*m)
			if err != nil {
				log.Printf("PublishMeasurement() failed: %v", err)
			}
		}
	}

	log.Printf("Exiting")
}
