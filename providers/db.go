package providers

import (
	"log"

	"github.com/samber/do"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"joints-be/config"
	sessionEntity "joints-be/modules/session/entity"
)

// ProvideDB registers *gorm.DB into the DI injector.
func ProvideDB(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (*gorm.DB, error) {
		cfg := config.AppConfig

		gormCfg := &gorm.Config{}
		if cfg.AppEnv == "development" {
			gormCfg.Logger = logger.Default.LogMode(logger.Info)
		} else {
			gormCfg.Logger = logger.Default.LogMode(logger.Silent)
		}

		db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormCfg)
		if err != nil {
			return nil, err
		}

		log.Println("[database] connected to PostgreSQL")

		// Auto-migrate entities — only used in development.
		// Replace with proper migration system before going to production.
		if cfg.AppEnv == "development" {
			if err := db.AutoMigrate(
				&sessionEntity.Session{},
			); err != nil {
				return nil, err
			}
			log.Println("[database] auto-migration completed")
		}

		return db, nil
	})
}

