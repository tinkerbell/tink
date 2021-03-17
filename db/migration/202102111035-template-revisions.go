package migration

import migrate "github.com/rubenv/sql-migrate"

// Get202102111035 introduces template revisions.
//
// Fixes: https://github.com/tinkerbell/tink/issues/413
//
// The migration introduces template revisions. A template can have multiple
// revisions, which are stored in the 'template_revisions' table.
//
// The migration also transforms the existing templates to the new
// template-revisions structure. Therefore ensuring that end users do not
// face any issues with the introduction of revisions.
func Get202102111035() *migrate.Migration {
	return &migrate.Migration{
		Id: "202102111035-template-revisions",
		Up: []string{`
CREATE TABLE IF NOT EXISTS template_revisions(
	template_id UUID
	, revision SMALLINT DEFAULT 1
	, data BYTEA
        , created_at TIMESTAMPTZ DEFAULT now()
        , deleted_at TIMESTAMPTZ
	, PRIMARY KEY(template_id, revision)
);

CREATE INDEX IF NOT EXISTS idx_template_revisions_template_id ON template_revisions (template_id);

ALTER TABLE template ADD COLUMN revision INT DEFAULT 1;
ALTER TABLE workflow ADD COLUMN revision INT DEFAULT 1;

CREATE OR REPLACE FUNCTION migrate_to_template_revisions()
RETURNS void AS $$
DECLARE
	template_count integer;
	revisions_count integer;
BEGIN
	SELECT COUNT(id) INTO template_count
	FROM template WHERE deleted_at IS NULL;

	INSERT INTO template_revisions(template_id, data, created_at, deleted_at)
	SELECT id, data, created_at, deleted_at
	FROM template;

	SELECT COUNT(template_id) INTO revisions_count
	FROM template_revisions WHERE deleted_at IS NULL;

	IF revisions_count != template_count THEN
		RAISE EXCEPTION 'failed to migrate templates';
	END IF;

	ALTER TABLE template DROP COLUMN data;
END;
$$ LANGUAGE plpgsql;

SELECT migrate_to_template_revisions();
		`},
	}
}
