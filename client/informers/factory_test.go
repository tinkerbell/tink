package informers

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/protos/events"
)

func TestListenerFactory(t *testing.T) {
	t.Run("test listener factory", func(t *testing.T) {
		factoryOne := Factory()
		factoryTwo := Factory()

		assert.Equal(t, factoryOne, factoryTwo)
	})
}

func TestGetKey(t *testing.T) {
	testCases := map[string]*events.WatchRequest{
		"returns valid key for nil event list":           watchRequest(withEventTypes(nil)),
		"returns valid key for nil resource list":        watchRequest(withResourceTypes(nil)),
		"returns valid key for all event/resource types": watchRequest(withAllEventTypes(), withAllResourceTypes()),
	}

	for name, req := range testCases {
		t.Run("return valid key "+name, func(t *testing.T) {
			key, err := getKey(req)
			assert.NoError(t, err)
			assert.NotEmpty(t, key)

			_, err = base64.StdEncoding.DecodeString(key)
			assert.NoError(t, err)
		})
	}
}
