package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kholmanskikh/home_sensors/zmq_api"

    "zmq_to_web/internal/config"
    "zmq_to_web/internal/publisher"
    "zmq_to_web/internal/publisher/web"
    "zmq_to_web/internal/publisher/mqtt"
)

var zmqPollTimeout = time.Second * 5

func main() {
	ret := 1
	defer func() {
		os.Exit(ret)
	}()

    configFile := flag.String("c", "", "Config file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Reads measurements from the ZMQ endpoint and posts them to HTTP or MQTT.

Usage: %s -c "<path_to_config_file>"

The config file is a JSON file of form:
%s
`, os.Args[0], config.Format)
	}

    flag.Parse()

    if *configFile == "" {
        flag.Usage()
        return
    }

    config, err := config.ParseFromFile(*configFile)
    if err != nil {
        log.Println(err)
        return
    }

	subscriber, err := zmq_api.NewSubscriber(config.ZMQEndpoint)
	if err != nil {
		log.Printf("NewSubscriber() failed: %v", err)
		return
	}
	defer subscriber.Destroy()
	log.Printf("ZMQ Endpoint: %s", subscriber.Endpoint)

    var publisher publisher.Publisher
    switch (config.Publisher) {
    case "web":
        publisher, err = web.NewWebPublisher(config.WebURL,
                                            time.Duration(config.WebUpdateTypesInterval) * time.Second)
    case "mqtt":
        publisher, err = mqtt.NewMQTTPublisher(config.MQTTBroker,
                                            config.MQTTUser, config.MQTTPassword,
                                            config.MQTTTopic)
    default:
        err = fmt.Errorf("unknown publisher type: %s", config.Publisher)
    }

	if err != nil {
		log.Printf("unable to create a publisher: %v", err)
		return
	}
    defer func() {
        if err := publisher.Destroy(); err != nil {
            log.Printf("Error while destroying the publisher: %v", err)
            ret = 1
        }
    }()

    log.Printf("Publisher: %s", publisher.Description())

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

    log.Printf("Begin operating")

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
			if config.Debug {
				log.Printf("Received %#v", *m)
			}

			err = publisher.PublishMeasurement(*m)
			if err != nil {
				log.Printf("PublishMeasurement() failed: %v", err)
			}
		}
	}

	log.Printf("Exiting")
}
