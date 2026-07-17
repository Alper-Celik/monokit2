package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

const (
	query      = "SELECT clock_timestamp() - pg_stat_activity.query_start AS duration, * FROM pg_stat_activity"
	moduleName = "process"
)

type activityInfo struct {
	Duration *time.Duration
	Fields   map[string]string
}

func CheckActivity(logger zerolog.Logger) {
	logger.Info().Msg("Checking PostgreSql processes...")

	activities := make([]activityInfo, 0, 30)
	activeActivities := make([]activityInfo, 0, 10)

	rows, err := Connection.Query(context.Background(), activityQuery)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query pg_stat_activity")
		return
	}
	defer rows.Close()

	for rows.Next() {
		columns, err := rows.Values()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scann a row of pg_stat_activity")
			return
		}
		row := activityInfo{
			Fields: map[string]string{},
		}
		for i, fd := range rows.FieldDescriptions() {
			columnStr := fmt.Sprint(columns[i])
			if len(columnStr) > 150 {
				columnStr = columnStr[:147] + "..."
			}
			row.Fields[fd.Name] = columnStr
		}

		if columns[0] != nil {
			dur := ToDuration(columns[0].(pgtype.Interval))
			row.Duration = &dur
		}

		activities = append(activities, row)

		if row.Fields["state"] == "active" {
			activeActivities = append(activeActivities, row)
		}

	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error occurred during rows iteration")
		return
	}

	logger.Info().Msgf("Successfully retrieved PostgreSql processes. %d processes found.", len(activities))
	logger.Debug().Interface("activities", activities).Msg("PostgreSql process details")

	checkThreshold(len(activeActivities), logger)
}

func ToDuration(i pgtype.Interval) time.Duration {
	if !i.Valid {
		return -1
	}
	const usecPerDay = 24 * 3600 * 1_000_000
	totalUsec := i.Microseconds +
		(int64(i.Days) * usecPerDay) +
		(int64(i.Months) * 30 * usecPerDay)

	return time.Duration(totalUsec) * time.Microsecond
}

func checkLongRunningQueries(activeActivities []activityInfo, logger zerolog.Logger) {
	// Down alarm if there is long running queries
	if lib.DBConfig.PostgreSql.Alarm.Enabled &&
		lib.DBConfig.PostgreSql.Alarm.LongQuery.Enabled {

		longRunningActivities := make([]activityInfo, len(activeActivities))

		for _, activity := range activeActivities {
			if activity.Duration.Seconds() > float64(lib.DBConfig.PostgreSql.Alarm.LongQuery.DurationSeconds) {
				longRunningActivities = append(longRunningActivities, activity)
			}
		}

		if len(longRunningActivities) > 0 {
			alarmMessage := fmt.Sprintf("[%s] - %s - PostgreSql has %d query(ies) running longer than %d seconds", pluginName, lib.GlobalConfig.Hostname, len(longRunningActivities), lib.DBConfig.PostgreSql.Alarm.LongQuery.DurationSeconds)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			}

		}

		// UP alarm if process count is below threshold
		if len(longRunningActivities) == 0 {
			alarmMessage := fmt.Sprintf("[%s] - %s - PostgreSql long running queries ended", pluginName, lib.GlobalConfig.Hostname)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to get last Zulip alarm")
				}

				if lastAlarm.Status == down {
					lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, up)
				}
			}
		}
	}
}

func checkThreshold(activeActivityCount int, logger zerolog.Logger) {
	activityThreshold := lib.DBConfig.PostgreSql.ActivityLimit

	// Down alarm if process count is above threshold
	if lib.DBConfig.PostgreSql.Alarm.Enabled {
		if activeActivityCount > activityThreshold {
			alarmMessage := fmt.Sprintf("[%s] - %s - PostgreSql activity count has been more than the set limit %d, (%d)", pluginName, lib.GlobalConfig.Hostname, activityThreshold, activeActivityCount)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			}

		}

		// UP alarm if process count is below threshold
		if activeActivityCount < activityThreshold {
			alarmMessage := fmt.Sprintf("[%s] - %s - PostgreSql activity count is back to normal (%d)", pluginName, lib.GlobalConfig.Hostname, activeActivityCount)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to get last Zulip alarm")
				}

				if lastAlarm.Status == down {
					lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, up)
				}
			}
		}
	}
}
