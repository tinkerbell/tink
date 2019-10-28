SET ROLE rover;

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

CREATE TABLE IF NOT EXISTS targets (
        id UUID UNIQUE
        , inserted_at TIMESTAMPTZ
        , deleted_at TIMESTAMPTZ
        , data JSONB
);

CREATE INDEX IF NOT EXISTS idx_rid ON targets (id);
CREATE INDEX IF NOT EXISTS idx_rdeleted_at ON targets (deleted_at NULLS FIRST);
CREATE INDEX IF NOT EXISTS idxgin_rtype ON targets USING GIN (data JSONB_PATH_OPS);

CREATE TABLE IF NOT EXISTS workflow (
	id UUID UNIQUE NOT NULL
	, template UUID NOT NULL
	, target UUID NOT NULL
	, state SMALLINT default 0
	, created_at TIMESTAMPTZ
	, updated_at TIMESTAMPTZ
	, deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_wid ON workflow (id);
CREATE INDEX IF NOT EXISTS idx_wdeleted_at ON workflow (deleted_at NULLS FIRST);
