package database

import (
	"database/sql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

func FromSQL(db *sql.DB) (*gorm.DB, error) {
	return gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

// Writer keeps ORM writes on the connection that owns the MariaDB advisory lock.
func Writer(conn *sql.Conn) (*gorm.DB, error) {
	return gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
}

func Migrate(db *gorm.DB, models ...any) error {
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin").AutoMigrate(models...)
}
