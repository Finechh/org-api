package database
 
import (
	"embed"
 
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
 
//go:embed *.sql
var migrationsFS embed.FS
 
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
 
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
 
	goose.SetBaseFS(migrationsFS)
 
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
 
	if err := goose.Up(sqlDB, "."); err != nil {
		return nil, err
	}
 
	return db, nil
}
 