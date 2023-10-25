package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

/**
 * This model contains the base data for all models
 * @param ID the numeric incremental id of the model
 * @param Uuid the uuid of the model
 * @param CreatedAt the date when the model was created
 * @param UpdatedAt the date when the model was updated
 * @param DeletedAt the date when the model was deleted
 *
 * UUID identification over numeric incremental id is used to prevent
 * the user from knowing the number of elements in the table
 *
 */
type BaseModel struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
