package migration

import migrate "github.com/rubenv/sql-migrate"

func Get202010071530() *migrate.Migration {
	return &migrate.Migration{
		Id: "202010071530-init-events-table-and-triggers",
		Up: []string{`
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$ BEGIN
    CREATE TYPE resource_type AS ENUM ('HARDWARE', 'TEMPLATE', 'WORKFLOW');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE event_type AS ENUM ('CREATED', 'UPDATED', 'DELETED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS events (
	id UUID UNIQUE DEFAULT uuid_generate_v4()
	, resource_id UUID NOT NULL
	, resource_type resource_type NOT NULL
	, event_type event_type NOT NULL
	, created_at TIMESTAMPTZ PRIMARY KEY DEFAULT now()
	, data JSONB
);

CREATE INDEX IF NOT EXISTS idx_events_event_type ON events (event_type);
CREATE INDEX IF NOT EXISTS idx_events_resource_id ON events (resource_id);
CREATE INDEX IF NOT EXISTS idx_events_resource_type ON events (resource_type);

CREATE OR REPLACE FUNCTION events_notify_changes()
RETURNS trigger AS $$
BEGIN
	PERFORM pg_notify('events_channel', row_to_json(NEW)::text);
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS events_channel ON events;

CREATE TRIGGER events_channel
AFTER INSERT ON events
FOR EACH ROW EXECUTE PROCEDURE events_notify_changes();

CREATE OR REPLACE FUNCTION insert_hardware_event()
RETURNS trigger AS $$
BEGIN
	IF (TG_OP = 'INSERT') THEN
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'HARDWARE', 'CREATED', row_to_json(new));
	ELSE
	IF new.deleted_at IS NULL THEN
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'HARDWARE', 'UPDATED', row_to_json(new));
	ELSE
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'HARDWARE', 'DELETED', row_to_json(new));
	END IF;
	END IF;

	RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS hardware_event_trigger ON hardware;

CREATE TRIGGER hardware_event_trigger
AFTER INSERT OR UPDATE ON hardware
FOR EACH ROW EXECUTE PROCEDURE insert_hardware_event();

CREATE OR REPLACE FUNCTION insert_template_event()
RETURNS trigger AS $$
BEGIN
	IF (TG_OP = 'INSERT') THEN
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'TEMPLATE', 'CREATED', row_to_json(new));
	ELSE
	IF new.deleted_at IS NULL THEN
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'TEMPLATE', 'UPDATED', row_to_json(new));
	ELSE
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'TEMPLATE', 'DELETED', row_to_json(new));
	END IF;
	END IF;

	RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS template_event_trigger ON template;

CREATE TRIGGER template_event_trigger
AFTER INSERT OR UPDATE ON template
FOR EACH ROW EXECUTE PROCEDURE insert_template_event();

CREATE OR REPLACE FUNCTION insert_workflow_event()
RETURNS trigger AS $$
BEGIN
	IF (TG_OP = 'INSERT') THEN
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'WORKFLOW', 'CREATED', row_to_json(new));
	ELSE
	IF new.deleted_at IS NULL THEN
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'WORKFLOW', 'UPDATED', row_to_json(new));
	ELSE
	INSERT INTO events(resource_id, resource_type, event_type, data) VALUES (new.id, 'WORKFLOW', 'DELETED', row_to_json(new));
	END IF;
	END IF;

	RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS workflow_event_trigger ON workflow;

CREATE TRIGGER workflow_event_trigger
AFTER INSERT OR UPDATE ON workflow
FOR EACH ROW EXECUTE PROCEDURE insert_workflow_event();
`},
	}
}
