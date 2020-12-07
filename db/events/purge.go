package events

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
)

const defaultEventsTTL = "60m"

// Purge periodically checks the events table and
// purges the events that have passed the defined EVENTS_TTL
func Purge(db *sql.DB, logger log.Logger) error {
	env := os.Getenv("EVENTS_TTL")
	if env == "" {
		env = defaultEventsTTL
	}
	val := strings.TrimRight(env, "m")
	ttl, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(ttl) * time.Minute)
	for range ticker.C {
		then := time.Now().Local().Add(time.Duration(int64(-ttl) * int64(time.Minute)))
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return errors.Wrap(err, "BEGIN transaction")
		}

		_, err = tx.Exec("DELETE FROM events WHERE created_at <= $1;", then)
		if err != nil {
			return errors.Wrap(err, "DELETE")
		}

		logger.Info("purging events")
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "COMMIT")
		}
	}
	return nil
}
