package main

import (
	"context"
	"fmt"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

const query = `
	SELECT
  datid::TEXT              ,-- oid
  datname::TEXT            ,-- name
  client_addr::TEXT        ,-- inet
  client_port::TEXT        ,-- integer
  backend_start::TEXT      ,-- timestamp with time zone
  xact_start::TEXT         ,-- timestamp with time zone
  query_start::TEXT        ,-- timestamp with time zone
  state_change::TEXT       ,-- timestamp with time zone
  backend_xid::TEXT        ,-- xid
  backend_xmin::TEXT       ,-- xid
  query_id::TEXT           ,-- bigint
  pid::TEXT                ,-- integer
  leader_pid::TEXT         ,-- integer
  usesysid::TEXT           ,-- oid
  client_hostname::TEXT    ,-- text
  wait_event_type::TEXT    ,-- text
  state::TEXT              ,-- text
  backend_type::TEXT       ,-- text
  wait_event::TEXT         ,-- text
  application_name::TEXT   ,-- text
  query::TEXT              ,-- text
  usename::TEXT            ,-- name
  datname::TEXT             -- name
	FROM pg_stat_activity
	`

func CheckProcess(logger zerolog.Logger) {
	moduleName := "process"
	logger.Info().Msg("Checking PostgreSql processes...")

	activities := make([]map[string]string, 0, 30)

	rows, err := Connection.Query(context.Background(), query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query pg_stat_activity")
		return
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		columns, err := rows.Values()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scann a row of pg_stat_activity")
			return
		}
		row := make(map[string]string)
		for i, fd := range rows.FieldDescriptions() {
			row[fd.Name] = fmt.Sprint(columns[i])
		}
		activities = append(activities, row)

		rowCount++
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error occurred during rows iteration")
		return
	}

	logger.Info().Msgf("Successfully retrieved PostgreSql processes. %d processes found.", rowCount)
	logger.Debug().Interface("activities", activities).Msg("PostgreSql process details")

	activityThreshold := lib.DBConfig.PostgreSql.ActivityLimit

	// Down alarm if process count is above threshold
	if lib.DBConfig.PostgreSql.Alarm.Enabled {
		if rowCount > activityThreshold {
			alarmMessage := fmt.Sprintf("[%s] - %s - PostgreSql activity count has been more than the set limit %d, (%d)", pluginName, lib.GlobalConfig.Hostname, activityThreshold, rowCount)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			}

		}

		// UP alarm if process count is below threshold
		if rowCount < activityThreshold {
			alarmMessage := fmt.Sprintf("[%s] - %s - PostgreSql activity count is back to normal (%d)", pluginName, lib.GlobalConfig.Hostname, rowCount)

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
