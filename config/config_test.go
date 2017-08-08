package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_AttributesFor(t *testing.T) {
	config := testConfig(t)

	assert.Equal(t, map[string]bool{"image": true}, config.AttributesFor("container", "start"))
	assert.Equal(t, map[string]bool{"image": false}, config.AttributesFor("container", "create"))
	assert.Equal(t, map[string]bool{"image": true, "exitCode": true}, config.AttributesFor("container", "die"))
	assert.Equal(t, map[string]bool{"image": true, "signal": true}, config.AttributesFor("container", "kill"))
}

const testConfigJson = `{
  "attributes": {
    "image": true
  },
  "events": {
    "image": {
      "actions": {
        "pull": {},
        "delete": {}
      }
    },
    "container": {
      "actions": {
        "create": {
          "attributes": {
            "image": false
          }
	},
        "start": {},
        "die": {
          "attributes": {
            "exitCode": true
          }
        },
        "kill": {
          "attributes": {
            "signal": true
          }
        }
      }
    }
  }
}`

func testConfig(t testing.TB) *Config {
	config, err := Load(strings.NewReader(testConfigJson))
	if err != nil {
		t.Fatal(err)
	}
	return config
}
