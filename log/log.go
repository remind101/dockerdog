package log

import (
	"fmt"
	"os"
	"text/template"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/remind101/dockerdog/config"
)

type Event struct {
	*docker.APIEvents
	Attributes map[string]string
}

func Watch(config *config.Config, events chan *docker.APIEvents, t *template.Template) error {
	fmt.Fprintln(os.Stderr, "Logging events to stdout")

	w := os.Stdout
	for event := range events {
		if _, ok := config.Events[event.Type]; !ok {
			continue
		}
		enabledAttributes := config.AttributesFor(event.Type, event.Action)

		attrs := make(map[string]string)
		for k, v := range event.Actor.Attributes {
			if enabledAttributes[k] {
				attrs[k] = v
			}
		}

		t.Execute(w, &Event{APIEvents: event, Attributes: attrs})
	}

	return nil
}
