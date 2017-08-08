package datadog

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/remind101/dockerdog/config"
)

func Watch(config *config.Config, events chan *docker.APIEvents, s *statsd.Client) error {
	for event := range events {
		if _, ok := config.Events[event.Type]; !ok {
			continue
		}

		enabledAttributes := config.AttributesFor(event.Type, event.Action)

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
