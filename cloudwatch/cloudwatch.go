package cloudwatch

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/remind101/dockerdog/config"
)

const (
	// Source is the value that will be used in the "Source" of the
	// PutEventRequestEntry.
	Source = "docker"

	// The maximum number of PutEventsRequestEntries per PutEvents API
	// call.
	MaxEntries = 10

	// Number of PutEvents requests to buffer before applying back pressure.
	PutEventsBuffer = 100
)

// Detail represents the structure of the `Detail` field we send in the
// PutEvents request.
type Detail struct {
	// Event is the raw Docker event.
	Event *docker.APIEvents `json:"event"`
}

type cloudwatcheventsClient interface {
	PutEvents(*cloudwatchevents.PutEventsInput) (*cloudwatchevents.PutEventsOutput, error)
}

type Worker struct {
	config *config.Config

	// events is where the worker will pull events from.
	events chan *docker.APIEvents

	// cloudwatch is the CloudWatch Events client that will be used to
	// PutEvents.
	cloudwatch cloudwatcheventsClient
}

func NewWorker(config *config.Config, events chan *docker.APIEvents) *Worker {
	cloudwatch := cloudwatchevents.New(session.New())

	return &Worker{
		config:     config,
		events:     events,
		cloudwatch: cloudwatch,
	}
}

func (w *Worker) Run() error {
	// The Docker event consumer will send PutEventsInput's that are ready
	// to be flushed to this channel.
	var (
		flush  = make(chan *cloudwatchevents.PutEventsInput, PutEventsBuffer)
		closed bool
	)

	// Represents our current buffer of PutEventsRequestEntries. When this
	// has reached 10 Entries, it will be flushed to the channel above.
	var input *cloudwatchevents.PutEventsInput

	for {
		select {
		case input, ok := <-flush:
			if !ok {
				log.Println("cloudwatch: All inputs flushed")
				// Channel is fully flushed and closed, return.
				return nil
			}

			_, err := w.cloudwatch.PutEvents(input)
			if err != nil {
				return err
			}
		case event, ok := <-w.events:
			if !ok {
				if !closed {
					if input != nil {
						flush <- input
					}

					// If the docker events channel has been
					// closed, we close the flush channel so
					// that we can flush all buffered inputs
					// before exiting.
					close(flush)
					closed = true
					log.Println("cloudwatch: Closing flush channel")
				}
				continue
			}

			if _, ok := w.config.Events[event.Type]; !ok {
				continue
			}

			raw, err := json.Marshal(&Detail{
				Event: event,
			})
			if err != nil {
				return err
			}

			if input == nil {
				input = new(cloudwatchevents.PutEventsInput)
			}

			input.Entries = append(input.Entries, &cloudwatchevents.PutEventsRequestEntry{
				Source:     aws.String("docker"),
				DetailType: aws.String(fmt.Sprintf("%s.%s", event.Type, event.Action)),
				Detail:     aws.String(string(raw)),
				Time:       aws.Time(time.Unix(0, event.TimeNano)),
			})

			if len(input.Entries) == MaxEntries {
				select {
				case flush <- input:
				default:
					log.Println("cloudwatch: Dropping %d events", len(input.Entries))
				}
				input = nil
			}
		}
	}

	return nil
}

func Watch(config *config.Config, events chan *docker.APIEvents) error {
	return NewWorker(config, events).Run()
}
