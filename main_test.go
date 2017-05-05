package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Attributes(t *testing.T) {
	config := testConfig(t)

	assert.Equal(t, map[string]bool{"image": true}, config.attributes("container", "start"))
	assert.Equal(t, map[string]bool{"image": false}, config.attributes("container", "create"))
	assert.Equal(t, map[string]bool{"image": true, "exitCode": true}, config.attributes("container", "die"))
	assert.Equal(t, map[string]bool{"image": true, "signal": true}, config.attributes("container", "kill"))
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

func testConfig(t testing.TB) *config {
	config, err := loadConfig(strings.NewReader(testConfigJson))
	if err != nil {
		t.Fatal(err)
	}
	return config
}
