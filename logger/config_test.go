package logger

import (
	"testing"
	"time"
)

func TestLoggerConfig_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   LoggerConfig
		want LoggerConfig
	}{
		{
			name: "all zero -> all defaults",
			in:   LoggerConfig{},
			want: LoggerConfig{
				ChannelBuffer:     100,
				NumWorkers:        3,
				CleanupInterval:   5 * time.Minute,
				MaxLogsPerProject: 100,
				WriteTimeout:      5 * time.Second,
			},
		},
		{
			name: "negative values are treated as zero",
			in:   LoggerConfig{ChannelBuffer: -1, NumWorkers: -2},
			want: LoggerConfig{
				ChannelBuffer:     100,
				NumWorkers:        3,
				CleanupInterval:   5 * time.Minute,
				MaxLogsPerProject: 100,
				WriteTimeout:      5 * time.Second,
			},
		},
		{
			name: "explicit values are preserved",
			in: LoggerConfig{
				ChannelBuffer:     50,
				NumWorkers:        7,
				CleanupInterval:   10 * time.Second,
				MaxLogsPerProject: 200,
				WriteTimeout:      2 * time.Second,
			},
			want: LoggerConfig{
				ChannelBuffer:     50,
				NumWorkers:        7,
				CleanupInterval:   10 * time.Second,
				MaxLogsPerProject: 200,
				WriteTimeout:      2 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.in
			c.applyDefaults()
			if c != tt.want {
				t.Errorf("applyDefaults: want %+v, got %+v", tt.want, c)
			}
		})
	}
}
