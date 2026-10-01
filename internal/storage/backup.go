package storage

import (
	"database/sql"
	"fmt"
	"io"
	"os"
)

// WriteBackup streams a consistent SQLite snapshot using VACUUM INTO.
func (s *MetricStore) WriteBackup(w io.Writer) error {
	tmp, err := os.CreateTemp("", "sysmon-backup-*.db")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpPath)

	if err := s.exportBackupFile(tmpPath); err != nil {
		return err
	}
	f, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

func (s *MetricStore) exportBackupFile(destPath string) error {
	return s.withDB(func(db *sql.DB) error {
		if _, err := db.Exec("PRAGMA wal_checkpoint(FULL)"); err != nil {
			return err
		}
		_, err := db.Exec("VACUUM INTO ?", destPath)
		if err != nil {
			return fmt.Errorf("vacuum into backup: %w", err)
		}
		return nil
	})
}
