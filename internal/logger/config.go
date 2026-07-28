package logger

import "time"

// LoggerConfig holds configuration for the Logger.
// Zero values are replaced with sensible defaults by applyDefaults.
type LoggerConfig struct {
	// ProjectID is the Google Cloud Project ID used for Firestore.
	ProjectID string
	// ChannelBuffer is the buffer size for the in-process log channel.
	ChannelBuffer int
	// NumWorkers is the number of worker goroutines consuming the channel.
	NumWorkers int
	// CleanupInterval controls how often the FIFO cleanup routine runs.
	CleanupInterval time.Duration
	// MaxLogsPerProject caps the number of logs retained per project (FIFO).
	MaxLogsPerProject int
	// WriteTimeout caps each Firestore write call.
	WriteTimeout time.Duration
}

func (c *LoggerConfig) applyDefaults() {
	if c.ChannelBuffer <= 0 {
		c.ChannelBuffer = 100
	}
	if c.NumWorkers <= 0 {
		c.NumWorkers = 3
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = 5 * time.Minute
	}
	if c.MaxLogsPerProject <= 0 {
		c.MaxLogsPerProject = 100
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 5 * time.Second
	}
}
