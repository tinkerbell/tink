package migration

import migrate "github.com/rubenv/sql-migrate"

// Get2020121691335 updates the primary key on events table.
//
// Fixes: https://github.com/tinkerbell/tink/issues/379
//
// We can have multiple events generated at a given time, therefore the value of
// 'created_at' field will be same for each event. This violates the unique constraint
// on "events_pkey", when these events are add to the events table.
//
// The migration changes the primary key on events table from 'created_at' to `id`,
// which will always be unique for each event generated at any point in time.
func Get2020121691335() *migrate.Migration {
	return &migrate.Migration{
		Id: "2020121691335-update-events-primary-key",
		Up: []string{`
			ALTER TABLE events DROP CONSTRAINT events_pkey;
			ALTER TABLE events ADD PRIMARY KEY (id);
			CREATE INDEX IF NOT EXISTS idx_events_created_at ON events (created_at);
		`},
	}
}
