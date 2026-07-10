package main

import (
	"context"
	"fmt"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

const (
	query      = "SELECT now() - pg_stat_activity.query_start AS duration, * FROM pg_stat_activity"
	moduleName = "process"
)

type activityInfo struct {
	duration time.Duration
	fields   map[string]string
}

func CheckActivity(logger zerolog.Logger) {
	logger.Info().Msg("Checking PostgreSql processes...")

	activities := make([]map[string]string, 0, 30)
	activeActivities := make([]map[string]string, 0, 10)

	rows, err := Connection.Query(context.Background(), query)
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
		row := make(map[string]string)
		for i, fd := range rows.FieldDescriptions() {
			columnStr := fmt.Sprint(columns[i])
			if len(columnStr) > 150 {
				columnStr = columnStr[:147] + "..."
			}
			row[fd.Name] = columnStr
		}
		rows.Scan()
		activities = append(activities, row)

		if row["state"] == "active" {
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
