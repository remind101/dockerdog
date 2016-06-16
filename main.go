package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/fsouza/go-dockerclient"
)

var supportedEventTypes = map[string]bool{
	"container": true,
	"image":     true,
	"volume":    false,
	"network":   false,
}

func main() {
	var (
		statsdAddr = flag.String("statsd", "localhost:8126", "Address of statsd")
	)

	s, err := statsd.New(*statsdAddr)
	if err != nil {
		log.Fatal(fmt.Errorf("could not connect to statsd: %v", err))
	}
	defer s.Close()

	d, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(fmt.Errorf("could not connect to Docker daemon: %v", err))
	}

	if err := watch(d, s); err != nil {
		log.Fatal(err)
	}
}

func watch(c *docker.Client, s *statsd.Client) error {
	events := make(chan *docker.APIEvents)
	if err := c.AddEventListener(events); err != nil {
		return fmt.Errorf("could not subscribe event listener: %v", err)
	}

	for event := range events {
		if !supportedEventTypes[event.Type] {
			continue
		}

		var tags []string
		for k, v := range event.Actor.Attributes {
			tags = append(tags, fmt.Sprintf("%s:%s", k, v))
		}

		s.Count(fmt.Sprintf("docker.events.%s.%s", event.Type, event.Action), 1, tags, 1)
	}

	return nil
}
