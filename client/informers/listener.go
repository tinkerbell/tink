package informers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/tinkerbell/tink/protos/events"
)

const (
	eventsChannel = "events_channel"
	pgDatabase    = "PGDATABASE"
	pgUser        = "PGUSER"
	pgPassword    = "PGPASSWORD"
	pgSSLMode     = "PGSSLMODE"
)

var (
	connInfo = fmt.Sprintf("dbname=%s user=%s password=%s sslmode=%s",
		os.Getenv(pgDatabase),
		os.Getenv(pgUser),
		os.Getenv(pgPassword),
		os.Getenv(pgSSLMode),
	)
)

// Listener is the base listener interface
type Listener interface {
	Listen(func(e *events.Event) error) error
}

type sharedListener struct {
	req        *events.WatchRequest
	pqListener *pq.Listener
}

func (s *sharedListener) Listen(fn func(e *events.Event) error) error {
	eventTypes := MapEventType(s.req.EventTypes)
	resourceTypes := MapResourceType(s.req.ResourceTypes)

	for {
		pqNotification := <-s.pqListener.Notify
		var notification Notification
		err := json.Unmarshal([]byte(pqNotification.Extra), &notification)
		if err != nil {
			return err
		}

		if notification.Filter(s.req.ResourceId, eventTypes, resourceTypes) {
			continue
		}
		var event events.Event
		err = notification.Unmarshal(&event)
		if err != nil {
			return err
		}
		err = fn(&event)
		if err != nil {
			return err
		}
	}
}

func newPQListener() (*pq.Listener, error) {
	_, err := sql.Open("postgres", "")
	if err != nil {
		return nil, err
	}

	listener := pq.NewListener(connInfo, 5*time.Second, 15*time.Second, nil)
	err = listener.Listen(eventsChannel)
	if err != nil {
		return nil, err
	}
	return listener, nil
}
