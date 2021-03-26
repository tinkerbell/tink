package migration

import migrate "github.com/rubenv/sql-migrate"

// Get202103261030 removes the event system tables and triggers.
//
// Fixes: https://github.com/tinkerbell/tink/issues/464
//
// The event system in place relies on triggers.
// It causes many problems because of its 8k characters limitation.
// CAPT is suffering from this limitation.
func Get2021032610300() *migrate.Migration {
	return &migrate.Migration{
		Id: "2021032610300-drop-events-system",
		Up: []string{`
DROP TABLE IF EXISTS events CASCADE;
DROP TRIGGER IF EXISTS events_channel ON events;
DROP TRIGGER IF EXISTS hardware_event_trigger ON hardware;
DROP TRIGGER IF EXISTS template_event_trigger ON template;
DROP TRIGGER IF EXISTS workflow_event_trigger ON workflow;
DROP TYPE IF EXISTS resource_type;
DROP TYPE IF EXISTS event_type;
`},
	}
}
