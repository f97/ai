package common

import (
	"database/sql"
	"fmt"
	"github.com/songquanpeng/one-api/common/env"
	"github.com/songquanpeng/one-api/common/logger"
)

// SQLite optimization settings for single-user gateway workload
// These settings prioritize low latency and performance over strict durability
const (
	// Default SQLite PRAGMA settings for single user
	DefaultJournalMode   = "WAL"           // WAL mode for better concurrency
	DefaultSynchronous   = "NORMAL"        // NORMAL instead of FULL for better performance
	DefaultCacheSize     = -64000          // 64MB cache (negative = KB)
	DefaultMMapSize      = 268435456       // 256MB mmap
	DefaultTempStore     = "MEMORY"        // Store temp tables in memory
	DefaultBusyTimeout   = 5000            // 5 seconds
	DefaultPageSize      = 4096            // 4KB pages
	DefaultAutoVacuum    = "INCREMENTAL"   // Incremental auto vacuum
	DefaultWALAutoCheckpoint = 1000        // Checkpoint every 1000 pages
)

// SQLiteOptimizerConfig holds SQLite optimization settings
type SQLiteOptimizerConfig struct {
	Enabled           bool
	JournalMode       string
	Synchronous       string
	CacheSize         int
	MMapSize          int64
	TempStore         string
	BusyTimeout       int
	PageSize          int
	AutoVacuum        string
	WALAutoCheckpoint int
	ForeignKeys       bool
}

// GetSQLiteOptimizerConfig returns the SQLite optimizer configuration from environment
func GetSQLiteOptimizerConfig() *SQLiteOptimizerConfig {
	enabled := env.Bool("SQLITE_OPTIMIZE_ENABLED", true)
	
	return &SQLiteOptimizerConfig{
		Enabled:           enabled,
		JournalMode:       env.String("SQLITE_JOURNAL_MODE", DefaultJournalMode),
		Synchronous:       env.String("SQLITE_SYNCHRONOUS", DefaultSynchronous),
		CacheSize:         env.Int("SQLITE_CACHE_SIZE", DefaultCacheSize),
		MMapSize:          int64(env.Int("SQLITE_MMAP_SIZE", int(DefaultMMapSize))),
		TempStore:         env.String("SQLITE_TEMP_STORE", DefaultTempStore),
		BusyTimeout:       env.Int("SQLITE_BUSY_TIMEOUT", DefaultBusyTimeout),
		PageSize:          env.Int("SQLITE_PAGE_SIZE", DefaultPageSize),
		AutoVacuum:        env.String("SQLITE_AUTO_VACUUM", DefaultAutoVacuum),
		WALAutoCheckpoint: env.Int("SQLITE_WAL_AUTO_CHECKPOINT", DefaultWALAutoCheckpoint),
		ForeignKeys:       env.Bool("SQLITE_FOREIGN_KEYS", false),
	}
}

// ApplySQLiteOptimizations applies optimization PRAGMAs to SQLite database
func ApplySQLiteOptimizations(sqlDB *sql.DB, config *SQLiteOptimizerConfig) error {
	if !config.Enabled {
		logger.SysLog("SQLite optimizations disabled")
		return nil
	}

	logger.SysLog("applying SQLite optimizations for single-user workload")

	pragmas := []struct {
		name  string
		value interface{}
	}{
		{"journal_mode", config.JournalMode},
		{"synchronous", config.Synchronous},
		{"cache_size", config.CacheSize},
		{"mmap_size", config.MMapSize},
		{"temp_store", config.TempStore},
		{"busy_timeout", config.BusyTimeout},
		{"auto_vacuum", config.AutoVacuum},
		{"wal_autocheckpoint", config.WALAutoCheckpoint},
	}

	// Apply foreign_keys separately as it's boolean
	if config.ForeignKeys {
		if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return fmt.Errorf("failed to enable foreign_keys: %w", err)
		}
	} else {
		if _, err := sqlDB.Exec("PRAGMA foreign_keys = OFF"); err != nil {
			return fmt.Errorf("failed to disable foreign_keys: %w", err)
		}
	}

	for _, pragma := range pragmas {
		query := fmt.Sprintf("PRAGMA %s = %v", pragma.name, pragma.value)
		if _, err := sqlDB.Exec(query); err != nil {
			return fmt.Errorf("failed to set %s: %w", pragma.name, err)
		}
		logger.SysLog(fmt.Sprintf("SQLite PRAGMA: %s = %v", pragma.name, pragma.value))
	}

	// Log current settings for verification
	var journalMode, synchronous string
	sqlDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	sqlDB.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	
	logger.SysLog(fmt.Sprintf("SQLite optimizations applied - journal_mode=%s, synchronous=%s", journalMode, synchronous))
	logger.SysLog(fmt.Sprintf("⚠️  TRADE-OFF: Using synchronous=%s may risk losing recent logs/usage on crash for better performance", config.Synchronous))

	return nil
}

// GetOptimalConnectionPoolSettings returns optimal connection pool settings for SQLite
func GetOptimalConnectionPoolSettings() (maxOpen, maxIdle int) {
	// For SQLite, we should use only 1 write connection to avoid SQLITE_BUSY errors
	// Multiple read connections are OK with WAL mode
	if env.Bool("SQLITE_OPTIMIZE_ENABLED", true) {
		maxOpen = env.Int("SQLITE_MAX_OPEN_CONNS", 5)  // Allow some reads in parallel
		maxIdle = env.Int("SQLITE_MAX_IDLE_CONNS", 2)  // Keep 2 connections idle
	} else {
		// Default values from model/main.go
		maxOpen = env.Int("SQL_MAX_OPEN_CONNS", 1000)
		maxIdle = env.Int("SQL_MAX_IDLE_CONNS", 100)
	}
	
	return maxOpen, maxIdle
}
