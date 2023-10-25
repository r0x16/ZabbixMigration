package db

import (
	"os"

	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/domain"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers/db/connection"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
 * Represents a mysql database provider connector using gorm
 */
type GormMysqlDatabaseProvider struct {
	Connection *gorm.DB
}

var _ domain.DatabaseProvider = &GormMysqlDatabaseProvider{}

/*
 * Creates a new dsn string for the mysql driver
 * using the connection struct and the environment variables
 */
func (g *GormMysqlDatabaseProvider) Connect() error {
	dsn := connection.GormMysqlConnection{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
	}

	var err error
	g.Connection, err = gorm.Open(mysql.Open(dsn.GetDsn()), &gorm.Config{})
	return err
}
