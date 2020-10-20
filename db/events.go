package db

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/protos/events"
)

// Events fetches events for a given time frame, and
// sends them to over the stream
func (d TinkDB) Events(req *events.WatchRequest, fn func(n informers.Notification) error) error {
	rows, err := d.instance.Query(`
	SELECT id, resource_id, resource_type, event_type, created_at, data
	FROM events
	WHERE
		created_at >= $1;
	`, req.GetWatchEventsFrom().AsTime())

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		n := informers.Notification{}
		err = rows.Scan(&n.ID, &n.ResourceID, &n.ResourceType, &n.EventType, &n.CreatedAt, &n.Data)
		if err != nil {
			err = errors.Wrap(err, "SELECT")
			logger.Error(err)
			return err
		}
		n.Prefix()
		if informers.Filter(&n, informers.Reduce(req)) {
			continue
		}
		err = fn(n)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err == sql.ErrNoRows {
		err = nil
	}
	return err
}
