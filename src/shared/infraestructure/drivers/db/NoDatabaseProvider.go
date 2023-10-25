package db

import "git.tnschile.com/sistemas/tnsgo/raidark/src/shared/domain"

/*
 * Represents a fake Database connection
 */
type NoDatabaseProvider struct {
}

var _ domain.DatabaseProvider = &NoDatabaseProvider{}

// Connect implements domain.DatabaseProvider.
func (*NoDatabaseProvider) Connect() error {
	return nil
}