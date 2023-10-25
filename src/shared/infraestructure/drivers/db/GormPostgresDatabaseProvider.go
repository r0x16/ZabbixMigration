package db

import (
	"os"

	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/domain"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers/db/connection"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
 * Represents a mysql database provider connector using gorm
 */
type GormPostgresDatabaseProvider struct {
	Connection *gorm.DB
}

var _ domain.DatabaseProvider = &GormPostgresDatabaseProvider{}

/*
 * Creates a new dsn string for the mysql driver
 * using the connection struct and the environment variables
 */
func (g *GormPostgresDatabaseProvider) Connect() error {
	dsn := connection.GormPostgresConnection{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
	}

	var err error
	g.Connection, err = gorm.Open(postgres.Open(dsn.GetDsn()), &gorm.Config{})
	return err
}
