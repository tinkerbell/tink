SET ROLE tinkerbell;

CREATE TABLE IF NOT EXISTS hardware (
	id UUID UNIQUE
	, inserted_at TIMESTAMPTZ
	, deleted_at TIMESTAMPTZ
	, data JSONB
);

CREATE INDEX IF NOT EXISTS idx_id ON hardware (id);
CREATE INDEX IF NOT EXISTS idx_deleted_at ON hardware (deleted_at NULLS FIRST);
CREATE INDEX IF NOT EXISTS idxgin_type ON hardware USING GIN (data JSONB_PATH_OPS);

CREATE TABLE IF NOT EXISTS template (
        id UUID UNIQUE NOT NULL
        , name VARCHAR(200) NOT NULL
        , created_at TIMESTAMPTZ
        , updated_at TIMESTAMPTZ
        , deleted_at TIMESTAMPTZ
        , data BYTEA

        CONSTRAINT CK_name CHECK (name ~ '^[a-zA-Z0-9_-]*$')
);

CREATE INDEX IF NOT EXISTS idx_tid ON template (id);
CREATE INDEX IF NOT EXISTS idx_tdeleted_at ON template (deleted_at NULLS FIRST);

CREATE TABLE IF NOT EXISTS workflow (
	id UUID UNIQUE NOT NULL
	, template UUID NOT NULL
	, devices JSONB NOT NULL
	, created_at TIMESTAMPTZ
	, updated_at TIMESTAMPTZ
	, deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_wid ON workflow (id);
CREATE INDEX IF NOT EXISTS idx_wdeleted_at ON workflow (deleted_at NULLS FIRST);

CREATE TABLE IF NOT EXISTS workflow_state (
        workflow_id UUID UNIQUE NOT NULL
        , current_task_name VARCHAR(200)
        , current_action_name VARCHAR(200)
        , current_action_state SMALLINT
        , current_worker VARCHAR(200)
        , action_list JSONB
        , current_action_index int
        , total_number_of_actions INT
);

CREATE INDEX IF NOT EXISTS idx_wfid ON workflow_state (workflow_id);

CREATE TABLE IF NOT EXISTS workflow_event (
        workflow_id UUID NOT NULL
        , worker_id UUID  NOT NULL
        , task_name VARCHAR(200)
        , action_name VARCHAR(200)
        , execution_time int
	, message VARCHAR(200)
	, status SMALLINT
        , created_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_event ON workflow_event (created_at);

CREATE TABLE IF NOT EXISTS workflow_worker_map (
        workflow_id UUID NOT NULL
        , worker_id UUID NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_data (
        workflow_id UUID NOT NULL
        , version INT
        , metadata JSONB
        , data JSONB
);
