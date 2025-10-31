package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CommandLog - structure for command log entry
type CommandLog struct {
	ID             int
	Timestamp      time.Time
	UserID         int
	Username       string
	Command        string
	Arguments      string
	OutputPreview  string
	ExecutionTime  int64
	Success        bool
}

// Database - wrapper for SQLite database
type Database struct {
	db *sql.DB
}

// InitDatabase - initialize database and create tables
func InitDatabase(dbPath string) (*Database, error) {
	// Create directory if needed
	dir := getOsUserHomeDir() + string(os.PathSeparator) + ".config"
	createDirIfNeed(dir)

	if dbPath == "" {
		dbPath = dir + string(os.PathSeparator) + "shell2telegram.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS command_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		command TEXT NOT NULL,
		arguments TEXT,
		output_preview TEXT,
		execution_time_ms INTEGER,
		success BOOLEAN NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON command_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_user_id ON command_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_command ON command_logs(command);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	log.Printf("Database initialized: %s", dbPath)
	return &Database{db: db}, nil
}

// Close - close database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// LogCommand - log command execution to database
func (d *Database) LogCommand(userID int, username, command, arguments string, output []byte, executionTime int64, success bool) error {
	if d.db == nil {
		return nil
	}

	// Limit output preview to 500 characters
	outputPreview := string(output)
	if len(outputPreview) > 500 {
		outputPreview = outputPreview[:500] + "..."
	}

	insertSQL := `
	INSERT INTO command_logs (timestamp, user_id, username, command, arguments, output_preview, execution_time_ms, success)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(insertSQL, time.Now(), userID, username, command, arguments, outputPreview, executionTime, success)
	if err != nil {
		log.Printf("Failed to log command to database: %v", err)
		return err
	}

	return nil
}

// GetRecentLogs - get recent N log entries
func (d *Database) GetRecentLogs(limit int) ([]CommandLog, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
	SELECT id, timestamp, user_id, username, command, arguments, output_preview, execution_time_ms, success
	FROM command_logs
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []CommandLog
	for rows.Next() {
		var log CommandLog
		err := rows.Scan(&log.ID, &log.Timestamp, &log.UserID, &log.Username, &log.Command, &log.Arguments, &log.OutputPreview, &log.ExecutionTime, &log.Success)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// SearchLogs - search logs by keyword
func (d *Database) SearchLogs(keyword string, limit int) ([]CommandLog, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
	SELECT id, timestamp, user_id, username, command, arguments, output_preview, execution_time_ms, success
	FROM command_logs
	WHERE command LIKE ? OR arguments LIKE ? OR output_preview LIKE ?
	ORDER BY timestamp DESC
	LIMIT ?
	`

	searchPattern := "%" + keyword + "%"
	rows, err := d.db.Query(query, searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []CommandLog
	for rows.Next() {
		var log CommandLog
		err := rows.Scan(&log.ID, &log.Timestamp, &log.UserID, &log.Username, &log.Command, &log.Arguments, &log.OutputPreview, &log.ExecutionTime, &log.Success)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetStats - get statistics from logs
func (d *Database) GetStats() (map[string]interface{}, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	stats := make(map[string]interface{})

	// Total commands
	var total int
	err := d.db.QueryRow("SELECT COUNT(*) FROM command_logs").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_commands"] = total

	// Success rate
	var successful int
	err = d.db.QueryRow("SELECT COUNT(*) FROM command_logs WHERE success = 1").Scan(&successful)
	if err != nil {
		return nil, err
	}
	stats["successful_commands"] = successful
	if total > 0 {
		stats["success_rate"] = float64(successful) / float64(total) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	// Average execution time
	var avgTime sql.NullFloat64
	err = d.db.QueryRow("SELECT AVG(execution_time_ms) FROM command_logs WHERE success = 1").Scan(&avgTime)
	if err != nil {
		return nil, err
	}
	if avgTime.Valid {
		stats["avg_execution_time_ms"] = avgTime.Float64
	} else {
		stats["avg_execution_time_ms"] = 0.0
	}

	// Most used command
	var mostUsedCmd sql.NullString
	var mostUsedCount sql.NullInt64
	err = d.db.QueryRow(`
		SELECT command, COUNT(*) as cnt
		FROM command_logs
		GROUP BY command
		ORDER BY cnt DESC
		LIMIT 1
	`).Scan(&mostUsedCmd, &mostUsedCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if mostUsedCmd.Valid {
		stats["most_used_command"] = mostUsedCmd.String
		stats["most_used_count"] = mostUsedCount.Int64
	}

	return stats, nil
}
