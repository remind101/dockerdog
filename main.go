package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/fsouza/go-dockerclient"
)

// config represents a config file that controls what events and actions to
// track.
type config struct {
	Attributes map[string]bool `json:"attributes"`

	Events map[string]struct {
		Actions map[string]struct {
			Attributes map[string]bool `json:"attributes"`
		} `json:"actions"`
	} `json:"events"`
}

func (c *config) attributes(event, action string) map[string]bool {
	attributes := make(map[string]bool)
	for k, v := range c.Attributes {
		attributes[k] = v
	}
	if e, ok := c.Events[event]; ok {
		if a, ok := e.Actions[action]; ok {
			for k, v := range a.Attributes {
				attributes[k] = v
			}
		}
	}
	return attributes
}

func loadConfig(r io.Reader) (*config, error) {
	var c config
	err := json.NewDecoder(r).Decode(&c)
	return &c, err
}

var supportedEventTypes = map[string]bool{
	"container": true,
	"image":     true,
	"volume":    false,
	"network":   false,
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		statsdAddr = flag.String("statsd", "localhost:8126", "Address of statsd")
	)
	flag.Parse()
	args := flag.Args()
	var r io.Reader = os.Stdin
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	config, err := loadConfig(r)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	s, err := statsd.New(*statsdAddr)
	if err != nil {
		return fmt.Errorf("could not connect to statsd: %v", err)
	}
	defer s.Close()

	d, err := docker.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("could not connect to Docker daemon: %v", err)
	}

	return watch(config, d, s)
}

func watch(config *config, c *docker.Client, s *statsd.Client) error {
	events := make(chan *docker.APIEvents)
	if err := c.AddEventListener(events); err != nil {
		return fmt.Errorf("could not subscribe event listener: %v", err)
	}

	for event := range events {
		if _, ok := config.Events[event.Type]; !ok {
			continue
		}

		enabledAttributes := config.attributes(event.Type, event.Action)

		var tags []string
		for k, v := range event.Actor.Attributes {
			if enabledAttributes[k] {
				tags = append(tags, fmt.Sprintf("%s:%s", k, v))
			}
		}

		s.Count(fmt.Sprintf("docker.events.%s.%s", event.Type, event.Action), 1, tags, 1)
	}

	return nil
}
