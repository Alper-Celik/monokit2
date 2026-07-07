package main

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func ConnectPSQL(logger zerolog.Logger) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), lib.DBConfig.PostgreSql.ConnectionString)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to PostgreSql")
		return nil, err
	}

	return conn, nil
}
