package main

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

const uptimeQuery = "SELECT pg_postmaster_start_time(), now() - pg_postmaster_start_time()"

type UptimeInfo struct {
	StartTime time.Time
	Uptime    time.Duration
}

func LogUptime(logger zerolog.Logger) {
	uptime, err := GetUptime(logger)
	if err != nil {
		return
	}
	logger.Debug().Interface("uptime", uptime).Msg("PostgreSQL uptime")
}

func GetUptime(logger zerolog.Logger) (uptime UptimeInfo, err error) {
	err = Connection.QueryRow(context.Background(), uptimeQuery).Scan(&uptime.StartTime, &uptime.Uptime)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to select pg_postmaster_start_time()")
	}
	return
}
