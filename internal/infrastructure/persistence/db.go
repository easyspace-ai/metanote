package persistence

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 封装 GORM 连接，供仓储使用
type DB struct {
	*gorm.DB
}

// New 创建数据库连接，启动时执行迁移（检查 schema_migrations，将未执行的 .up.sql 按版本顺序落地到库）
func New(databasePath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := RunMigrations(sqlDB); err != nil {
		return nil, err
	}

	log.Println("✅ GORM connected to SQLite at", databasePath)
	return &DB{DB: db}, nil
}

// Close 关闭连接
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
