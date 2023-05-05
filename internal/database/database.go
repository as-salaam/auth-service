package database

import (
	"fmt"
	"github.com/as-salaam/auth-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBInit initializes a database connection and gorm.DB object
func DBInit(host, name, user, password string, port uint, timezone string, ssl bool) (*gorm.DB, error) {
	sslMode := "disable"
	if ssl {
		sslMode = "enable"
	}
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s port=%d timezone=%s sslmode=%s", host, name, user, password, port, timezone, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
