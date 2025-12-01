package cmd

import (
	"fmt"
	"kubez/internal/database"
	"os"

	"github.com/spf13/cobra"
)

var (
	dbURL          string
	migrationsPath string
)

// migrateCmd 数据库迁移命令
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "数据库迁移管理",
	Long:  `管理数据库迁移，包括执行、回滚和查看迁移状态`,
}

// migrateUpCmd 执行迁移
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行数据库迁移",
	Long:  `执行所有待执行的数据库迁移`,
	RunE: func(cmd *cobra.Command, args []string) error {
		migrator, err := database.NewMigrator(dbURL, migrationsPath)
		if err != nil {
			return fmt.Errorf("创建迁移器失败: %w", err)
		}
		defer migrator.Close()

		fmt.Println("开始执行数据库迁移...")
		if err := migrator.Up(); err != nil {
			return fmt.Errorf("执行迁移失败: %w", err)
		}

		version, _, _ := migrator.Version()
		fmt.Printf("✓ 数据库迁移完成，当前版本: %d\n", version)
		return nil
	},
}

// migrateDownCmd 回滚迁移
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚数据库迁移",
	Long:  `回滚最后一次数据库迁移`,
	RunE: func(cmd *cobra.Command, args []string) error {
		migrator, err := database.NewMigrator(dbURL, migrationsPath)
		if err != nil {
			return fmt.Errorf("创建迁移器失败: %w", err)
		}
		defer migrator.Close()

		fmt.Println("开始回滚数据库迁移...")
		if err := migrator.Down(); err != nil {
			return fmt.Errorf("回滚失败: %w", err)
		}

		version, _, _ := migrator.Version()
		fmt.Printf("✓ 数据库回滚完成，当前版本: %d\n", version)
		return nil
	},
}

// migrateStatusCmd 查看迁移状态
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	Long:  `查看当前数据库迁移的版本和状态`,
	RunE: func(cmd *cobra.Command, args []string) error {
		migrator, err := database.NewMigrator(dbURL, migrationsPath)
		if err != nil {
			return fmt.Errorf("创建迁移器失败: %w", err)
		}
		defer migrator.Close()

		version, dirty, err := migrator.Version()
		if err != nil {
			return fmt.Errorf("获取迁移状态失败: %w", err)
		}

		fmt.Printf("当前迁移版本: %d\n", version)
		if dirty {
			fmt.Println("状态: dirty (不一致)")
		} else {
			fmt.Println("状态: clean (正常)")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// 添加全局参数
	migrateCmd.PersistentFlags().StringVar(&dbURL, "db-url", getEnv("KUBEZ_DB_URL", "postgres://postgres:postgres@localhost:5432/kubez?sslmode=disable"), "数据库连接字符串")
	migrateCmd.PersistentFlags().StringVar(&migrationsPath, "migrations-path", getEnv("KUBEZ_MIGRATIONS_PATH", "./migrations"), "迁移脚本目录路径")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
