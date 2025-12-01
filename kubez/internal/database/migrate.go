package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Migrator 数据库迁移器
type Migrator struct {
	db      *sql.DB
	migrate *migrate.Migrate
}

// NewMigrator 创建新的迁移器
// dbURL: 数据库连接字符串，格式为 postgres://user:password@host:port/dbname?sslmode=disable
// migrationsPath: 迁移脚本所在目录路径，格式为 file://path/to/migrations
func NewMigrator(dbURL, migrationsPath string) (*Migrator, error) {
	// 打开数据库连接
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 创建 postgres 驱动实例
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建数据库驱动失败: %w", err)
	}

	// 确保迁移路径以 file:// 开头
	if migrationsPath[:7] != "file://" {
		migrationsPath = "file://" + migrationsPath
	}

	// 创建迁移实例
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建迁移实例失败: %w", err)
	}

	return &Migrator{
		db:      db,
		migrate: m,
	}, nil
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up() error {
	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行迁移失败: %w", err)
	}
	return nil
}

// Down 回滚最后一次迁移
func (m *Migrator) Down() error {
	err := m.migrate.Steps(-1)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("回滚迁移失败: %w", err)
	}
	return nil
}

// Version 获取当前迁移版本和状态
// 返回值：version - 当前版本号, dirty - 是否处于不一致状态, error - 错误信息
func (m *Migrator) Version() (version uint, dirty bool, err error) {
	version, dirty, err = m.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("获取迁移版本失败: %w", err)
	}
	return version, dirty, nil
}

// Close 关闭迁移器和数据库连接
func (m *Migrator) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
