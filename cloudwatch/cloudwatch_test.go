package cloudwatch

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/remind101/dockerdog/config"
	"github.com/stretchr/testify/assert"
)

func TestWorker_ClosedChannel(t *testing.T) {
	config := newConfig(t)
	events := make(chan *docker.APIEvents)
	close(events)
	c := new(mockCloudWatchEventsClient)
	w := &Worker{
		config:     config,
		events:     events,
		cloudwatch: c,
	}

	err := w.Run()
	assert.NoError(t, err)
}

func TestWorker_ClosedWithFlush(t *testing.T) {
	config := newConfig(t)
	events := make(chan *docker.APIEvents, 1)
	c := &mockCloudWatchEventsClient{
		putEvents: make(chan *cloudwatchevents.PutEventsInput, 1),
	}
	w := &Worker{
		config:     config,
		events:     events,
		cloudwatch: c,
	}

	events <- &docker.APIEvents{Type: "image", Action: "pull"}
	close(events)

	err := w.Run()
	assert.NoError(t, err)
	assert.Equal(t, 1, len((<-c.putEvents).Entries))
}

const testConfig = `{
  "events": {
    "image": {
      "actions": {
        "pull": {}
      }
    }
  }
}`

func newConfig(t testing.TB) *config.Config {
	config, err := config.Load(strings.NewReader(testConfig))
	if err != nil {
		t.Fatal(err)
	}
	return config
}

type mockCloudWatchEventsClient struct {
	putEvents chan *cloudwatchevents.PutEventsInput
}

func (m *mockCloudWatchEventsClient) PutEvents(input *cloudwatchevents.PutEventsInput) (*cloudwatchevents.PutEventsOutput, error) {
	m.putEvents <- input
	return nil, nil
}
