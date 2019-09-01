package metadata

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/utils"
)

var _ Storage = (*storageSQLite)(nil)

type storageSQLite struct {
	db     *sql.DB
	logger zerolog.Logger
}

func (s *storageSQLite) GetServices(ctx context.Context) ([]string, error) {
	var result []string

	callback := func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, "SELECT name FROM services")
		if err != nil {
			return errors.Wrap(err, "get services: select services")
		}
		defer func() {
			if txErr := rows.Close(); txErr != nil {
				s.logger.Error().Err(err).Msg("get services: close rows")
			}
		}()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return errors.Wrap(err, "get services: scan rows")
			}
			result = append(result, name)
		}
		return nil
	}

	if err := s.wrapTx(ctx, callback); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *storageSQLite) GetInstances(ctx context.Context, serviceName string) ([]*schema.InstanceDescription, error) {
	var result []*schema.InstanceDescription

	callback := func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(
			ctx,
			"SELECT name FROM instances WHERE service_id = (SELECT id FROM services WHERE name = ?)",
			serviceName,
		)
		if err != nil {
			return errors.Wrap(err, "get instances: select instances")
		}
		defer func() {
			if txErr := rows.Close(); txErr != nil {
				s.logger.Error().Err(err).Msg("get instances: close rows")
			}
		}()
		for rows.Next() {
			var instanceName string
			if err := rows.Scan(&instanceName); err != nil {
				return errors.Wrap(err, "get instances: scan rows")
			}
			desc := &schema.InstanceDescription{
				ServiceName:  serviceName,
				InstanceName: instanceName,
			}
			result = append(result, desc)
		}
		return nil
	}

	if err := s.wrapTx(ctx, callback); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *storageSQLite) GetSessions(
	ctx context.Context,
	instanceDesc *schema.InstanceDescription,
) ([]*schema.Session, error) {
	var result []*schema.Session

	callback := func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(
			ctx,
			"SELECT id, started_at, finished_at FROM sessions WHERE instance_id = ("+
				"SELECT id FROM instances WHERE name = ? AND service_id = (SELECT id FROM services WHERE name = ?))",
			instanceDesc.GetInstanceName(), instanceDesc.GetServiceName(),
		)
		if err != nil {
			return errors.Wrap(err, "get sessions: select instances")
		}
		defer func() {
			if txErr := rows.Close(); txErr != nil {
				s.logger.Error().Err(err).Msg("get sessions: close rows")
			}
		}()
		for rows.Next() {
			var (
				sessionID  int64
				startedAt  sql.NullString
				finishedAt sql.NullString
			)
			if err := rows.Scan(&sessionID, &startedAt, &finishedAt); err != nil {
				return errors.Wrap(err, "get sessions: scan rows")
			}
			startedAtTstamp, err := parseSQLiteTimeToTimestamp(startedAt)
			if err != nil {
				return errors.Wrap(err, "get sessions: convert started_at time to timestamp")
			}
			finishedAtTstamp, err := parseSQLiteTimeToTimestamp(finishedAt)
			if err != nil {
				return errors.Wrap(err, "get sessions: convert finished_at time to timestamp")
			}
			session := &schema.Session{
				Description: &schema.SessionDescription{
					InstanceDescription: instanceDesc,
					Id:                  sessionID,
				},
				Metadata: &schema.SessionMetadata{
					StartedAt:  startedAtTstamp,
					FinishedAt: finishedAtTstamp,
				},
			}
			result = append(result, session)
		}
		return nil
	}

	if err := s.wrapTx(ctx, callback); err != nil {
		return nil, err
	}

	return result, nil
}

const timeFormatSQLite = "2006-01-02T15:04:05Z"

func parseSQLiteTimeToTimestamp(src sql.NullString) (*timestamp.Timestamp, error) {
	if !src.Valid {
		return nil, nil
	}

	parsed, err := time.Parse(timeFormatSQLite, src.String)
	if err != nil {
		return nil, errors.Wrap(err, "parse SQL timestamp")
	}

	result, err := ptypes.TimestampProto(parsed)
	if err != nil {
		return nil, errors.Wrap(err, "convert time to protobuf timestamp")
	}

	return result, nil
}

func (s *storageSQLite) StartSession(
	ctx context.Context,
	instanceDesc *schema.InstanceDescription,
) (*schema.SessionDescription, error) {
	var sessionDesc *schema.SessionDescription

	callback := func(tx *sql.Tx) error {
		// create service if not exists
		_, err := tx.ExecContext(
			ctx,
			`INSERT OR IGNORE INTO services (name) VALUES (?)`,
			instanceDesc.GetServiceName(),
		)
		if err != nil {
			return errors.Wrap(err, "start session: insert service")
		}
		var serviceID int64
		err = tx.QueryRowContext(
			ctx,
			`SELECT id FROM services WHERE name == ? `,
			instanceDesc.GetServiceName(),
		).Scan(&serviceID)
		if err != nil {
			return errors.Wrap(err, "start session: select service id")
		}

		// create instance if not exists
		_, err = tx.ExecContext(
			ctx,
			`INSERT OR IGNORE INTO instances (name, service_id) VALUES (?, ?)`,
			instanceDesc.GetInstanceName(), serviceID,
		)
		if err != nil {
			return errors.Wrap(err, "start session: insert instance")
		}
		var instanceID int64
		err = tx.QueryRowContext(
			ctx,
			`SELECT id FROM instances WHERE name == ? AND service_id = ?`,
			instanceDesc.GetInstanceName(), serviceID).Scan(&instanceID)
		if err != nil {
			return errors.Wrap(err, "start session: select instance id")
		}

		// create new session
		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO sessions (started_at, instance_id) VALUES (CURRENT_TIMESTAMP, ?)`,
			instanceID,
		)
		if err != nil {
			return errors.Wrap(err, "start session: insert session")
		}
		sessionID, err := result.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "start session: extract session id")
		}

		sessionDesc = &schema.SessionDescription{InstanceDescription: instanceDesc, Id: sessionID}
		return nil
	}

	if err := s.wrapTx(ctx, callback); err != nil {
		return nil, err
	}
	return sessionDesc, nil
}

func (s *storageSQLite) StopSession(ctx context.Context, description *schema.SessionDescription) error {
	callback := func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`UPDATE sessions SET finished_at = CURRENT_TIMESTAMP WHERE id = ?`,
			description.Id,
		)
		if err != nil {
			return errors.Wrap(err, "stop session")
		}
		return nil
	}
	return s.wrapTx(ctx, callback)
}

func (s *storageSQLite) Quit() {
	if err := s.db.Close(); err != nil {
		s.logger.Error().Err(err).Msg("close metadata storage")
	}
}

func (s *storageSQLite) wrapTx(ctx context.Context, txFunc func(*sql.Tx) error) (err error) {
	// FIXME: must isolation level be parametrized?
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSnapshot, ReadOnly: false})
	if err != nil {
		err = errors.Wrap(err, "begin tx")
		return
	}

	defer func() {
		if err != nil {
			s.logger.Error().Err(err).Msg("SQL error")
			if txErr := tx.Rollback(); txErr != nil {
				s.logger.Error().Err(err).Msg("Transaction rollback error")
			}
		} else if txErr := tx.Commit(); txErr != nil {
			s.logger.Error().Err(err).Msg("Transaction commit error")
		}
	}()
	err = txFunc(tx)
	return
}

func NewStorageSQLite(logger *zerolog.Logger, cfg *config.MetadataStorageConfig) (Storage, error) {
	if !utils.FileExists(cfg.DataDir) {
		if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
			return nil, errors.Wrap(err, "create dir")
		}
	}

	fileName := filepath.Join(cfg.DataDir, "metadata.db")
	fileExists := utils.FileExists(fileName)

	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, errors.Wrap(err, "open metadata storage")
	}

	// initialize database if necessary
	if !fileExists {

		sqlInitStmt := `
		CREATE TABLE services (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE instances (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			name TEXT NOT NULL UNIQUE,
			service_id INTEGER,
			FOREIGN KEY (service_id)
				REFERENCES services (id)
					ON DELETE CASCADE
		);
		CREATE TABLE sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			started_at datetime NOT NULL,
			finished_at datetime,
			instance_id INTEGER,
			FOREIGN KEY (instance_id)
				REFERENCES instances (id)
					ON DELETE CASCADE
		);
	`

		_, err := db.Exec(sqlInitStmt)
		if err != nil {
			return nil, errors.Wrap(err, "initialize database")
		}
	}

	return &storageSQLite{db: db}, nil
}
