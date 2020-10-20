package listener

import (
	sha "crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/protos/events"
)

const eventsChannel = "events_channel"

type listener struct {
	mu         sync.Mutex
	pqListener *pq.Listener
	handlers   map[string]*handler
	listening  bool
}

type handler struct {
	req *events.WatchRequest
	fn  func(e *events.Event) error
}

var sharedListener *listener

// Init initializes a PostgreSQL listener.
func Init(connInfo string) error {
	pql := pq.NewListener(connInfo, 5*time.Second, 15*time.Second, nil)
	sharedListener = &listener{
		pqListener: pql,
		handlers:   map[string]*handler{},
	}
	return nil
}

// Listen listens for incoming events.
func Listen(req *events.WatchRequest, fn func(e *events.Event) error) error {
	key, err := getKey(req)
	if err != nil {
		return err
	}
	sharedListener.mu.Lock()
	sharedListener.handlers[key] = &handler{
		req: req,
		fn:  fn,
	}
	sharedListener.mu.Unlock()

	if !sharedListener.listening {
		err := sharedListener.pqListener.Listen(eventsChannel)
		if err != nil {
			return err
		}
		sharedListener.listening = true
	}

	for {
		pqNotification := <-sharedListener.pqListener.Notify
		var n informers.Notification
		err := json.Unmarshal([]byte(pqNotification.Extra), &n)
		if err != nil {
			return err
		}

		n.Prefix()
		for _, handler := range sharedListener.handlers {
			if informers.Filter(&n, informers.Reduce(handler.req)) {
				continue
			}

			event, err := n.ToEvent()
			if err != nil {
				return err
			}

			err = handler.fn(event)
			if err != nil {
				return err
			}
		}
	}
}

// RemoveHandlers removes the registered handlers for given watch request
func RemoveHandlers(req *events.WatchRequest) error {
	key, err := getKey(req)
	if err != nil {
		return err
	}
	sharedListener.mu.Lock()
	delete(sharedListener.handlers, key)
	sharedListener.mu.Unlock()
	return nil
}

func getKey(req *events.WatchRequest) (string, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	h := sha.New()
	return base64.StdEncoding.EncodeToString(h.Sum(reqBytes)), nil
}
