package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/isobit/cli"

	dockertypes "github.com/docker/docker/api/types"
	dockerfilters "github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

type DockerDog struct {
	Debug         bool
	StatsdAddr    string
	AttributeTags []string `cli:"name=attribute-tag,short=a,repeatable"`

	globalAttributeTags map[string]string
	eventAttributeTags  map[string]map[string]string
}

func NewDockerDog() *DockerDog {
	return &DockerDog{
		StatsdAddr:          "localhost:8125",
		globalAttributeTags: map[string]string{},
		eventAttributeTags: map[string]map[string]string{
			"container.attach":      {},
			"container.create":      {},
			"container.destroy":     {},
			"container.detach":      {},
			"container.die":         {"exitCode": "exitCode"},
			"container.exec_create": {},
			"container.exec_detach": {},
			"container.exec_start":  {},
			"container.kill":        {"signal": "signal"},
			"container.oom":         {},
			"container.start":       {},
			"container.stop":        {},
			"image.delete":          {},
			"image.import":          {},
			"image.load":            {},
			"image.pull":            {},
			"image.push":            {},
			"image.save":            {},
		},
	}
}

func (cmd *DockerDog) Before() error {
	if cmd.AttributeTags != nil {
		for _, attrTag := range cmd.AttributeTags {
			if attr, tag, ok := strings.Cut(attrTag, ":"); ok {
				cmd.globalAttributeTags[attr] = tag
			} else {
				cmd.globalAttributeTags[attrTag] = attrTag
			}
		}
	}
	return nil
}

func (cmd *DockerDog) Run() error {
	statsd, err := statsd.New(cmd.StatsdAddr)
	if err != nil {
		return fmt.Errorf("statsd error: %w", err)
	}
	defer statsd.Close()

	docker, err := dockerclient.NewClientWithOpts()
	if err != nil {
		return fmt.Errorf("docker error: %w", err)
	}
	defer docker.Close()

	filters := dockerfilters.NewArgs()
	for k, _ := range cmd.eventAttributeTags {
		if typeName, actionName, ok := strings.Cut(k, "."); ok {
			filters.Add("type", typeName)
			filters.Add("event", actionName)
		}
	}

	fmt.Printf("listening to docker events, sending stats to %s\n", cmd.StatsdAddr)
	events, errs := docker.Events(
		context.Background(),
		dockertypes.EventsOptions{
			Filters: filters,
		},
	)
	for {
		select {
		case event := <-events:
			key := event.Type + "." + event.Action

			attributeTags, ok := cmd.eventAttributeTags[key]
			if !ok {
				fmt.Printf("ignoring %s\n", key)
				continue
			}

			tags := []string{}
			for attr, tag := range cmd.globalAttributeTags {
				if v, ok := event.Actor.Attributes[attr]; ok {
					tags = append(tags, fmt.Sprintf("%s:%s", tag, v))
				}
			}
			for attr, tag := range attributeTags {
				if v, ok := event.Actor.Attributes[attr]; ok {
					tags = append(tags, fmt.Sprintf("%s:%s", tag, v))
				}
			}

			metricName := "docker.events." + key
			if cmd.Debug {
				fmt.Printf("%s %v\n", metricName, tags)
			}
			statsd.Count(metricName, 1, tags, 1)

		case err := <-errs:
			return err
		}
	}
}

func main() {
	cli.New("dockerdog", NewDockerDog()).
		Parse().
		RunFatal()
}
