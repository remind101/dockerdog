package config

import (
	"encoding/json"
	"io"
)

// Config represents a config file that controls what events and actions to
// track.
type Config struct {
	// Attributes defines any global attributes to include across all events
	// and actions.
	Attributes map[string]bool `json:"attributes"`

	// Events configures the events that should be tracked.
	Events map[string]struct {
		// Actions configures the actions that should be tracked.
		Actions map[string]struct {
			// Attributes configures the attributes in the action
			// that should be included.
			Attributes map[string]bool `json:"attributes"`
		} `json:"actions"`
	} `json:"events"`
}

// AttributesFor returns a map of the attributes that should be included for a
// given action.
func (c *Config) AttributesFor(event, action string) map[string]bool {
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

// Load parses the given json config file in r and returns a parsed
// config.
func Load(r io.Reader) (*Config, error) {
	var c Config
	err := json.NewDecoder(r).Decode(&c)
	return &c, err
}
