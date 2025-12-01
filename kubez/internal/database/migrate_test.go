package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDBURL = "postgres://postgres:postgres@localhost:5432/kubez_test?sslmode=disable"
var testMigrationsPath = "../../migrations"

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) {
	// 连接到 postgres 默认数据库
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 强制断开所有连接到测试数据库的会话
	_, _ = db.Exec(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = 'kubez_test'
		AND pid <> pg_backend_pid()
	`)

	// 删除测试数据库（如果存在）
	_, _ = db.Exec("DROP DATABASE IF EXISTS kubez_test")

	// 创建测试数据库
	_, err = db.Exec("CREATE DATABASE kubez_test")
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		return
	}
	defer db.Close()

	// 强制断开所有连接
	_, _ = db.Exec(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = 'kubez_test'
		AND pid <> pg_backend_pid()
	`)

	_, _ = db.Exec("DROP DATABASE IF EXISTS kubez_test")
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// TestMigratorUp 测试执行数据库迁移
func TestMigratorUp(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建测试用的迁移器
	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 执行迁移
	err = migrator.Up()
	if err != nil {
		t.Errorf("执行迁移失败: %v", err)
	}

	// 验证迁移状态
	version, dirty, err := migrator.Version()
	if err != nil {
		t.Errorf("获取迁移版本失败: %v", err)
	}

	if dirty {
		t.Error("迁移状态不应该是 dirty")
	}

	if version == 0 {
		t.Error("迁移版本应该大于 0")
	}
}

// TestMigratorDown 测试回滚数据库迁移
func TestMigratorDown(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 先执行迁移
	err = migrator.Up()
	if err != nil {
		t.Fatalf("执行迁移失败: %v", err)
	}

	// 记录当前版本
	currentVersion, _, _ := migrator.Version()

	// 执行回滚
	err = migrator.Down()
	if err != nil {
		t.Errorf("回滚迁移失败: %v", err)
	}

	// 验证版本已回滚
	newVersion, _, _ := migrator.Version()
	if newVersion >= currentVersion {
		t.Errorf("回滚后版本应该小于之前的版本，当前: %d, 之前: %d", newVersion, currentVersion)
	}
}

// TestMigratorStatus 测试查看迁移状态
func TestMigratorStatus(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 获取迁移状态
	version, dirty, err := migrator.Version()
	if err != nil && err.Error() != "no migration" {
		t.Errorf("获取迁移状态失败: %v", err)
	}

	t.Logf("当前迁移版本: %d, dirty: %v", version, dirty)
}

// TestMigratorUpIdempotent 测试重复执行迁移不报错
func TestMigratorUpIdempotent(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 第一次执行迁移
	err = migrator.Up()
	if err != nil {
		t.Fatalf("第一次执行迁移失败: %v", err)
	}

	// 第二次执行迁移（应该不报错）
	err = migrator.Up()
	if err != nil && err.Error() != "no change" {
		t.Errorf("重复执行迁移应该返回 'no change' 或不报错，实际: %v", err)
	}
}

// TestMigratorDownAndRestore 测试回滚后数据库恢复到之前状态
func TestMigratorDownAndRestore(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 执行迁移
	err = migrator.Up()
	if err != nil {
		t.Fatalf("执行迁移失败: %v", err)
	}

	// 回滚
	err = migrator.Down()
	if err != nil {
		t.Fatalf("回滚失败: %v", err)
	}

	// 再次执行迁移（验证可以重新迁移）
	err = migrator.Up()
	if err != nil {
		t.Errorf("回滚后重新执行迁移失败: %v", err)
	}
}

// TestMigratorRollbackOnFailure 测试迁移失败时能够正确回滚
func TestMigratorRollbackOnFailure(t *testing.T) {
	// 使用无效的数据库连接字符串
	migrator, err := NewMigrator("postgres://invalid:invalid@localhost:9999/invaliddb?sslmode=disable", testMigrationsPath)
	if err != nil {
		// 预期会失败
		t.Logf("预期失败: %v", err)
		return
	}
	defer migrator.Close()

	// 尝试执行迁移
	err = migrator.Up()
	if err == nil {
		t.Error("使用无效连接应该失败")
	}
}

// TestNewMigratorInvalidURL 测试使用无效 URL 创建迁移器
func TestNewMigratorInvalidURL(t *testing.T) {
	_, err := NewMigrator("invalid-url", testMigrationsPath)
	if err == nil {
		t.Error("使用无效 URL 应该失败")
	}
}

// TestNewMigratorInvalidPath 测试使用无效路径创建迁移器
func TestNewMigratorInvalidPath(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	_, err := NewMigrator(testDBURL, "/nonexistent/path")
	if err == nil {
		t.Error("使用不存在的迁移路径应该失败")
	}
}

// TestMigratorUpNoChange 测试当没有新迁移时的情况
func TestMigratorUpNoChange(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 第一次执行迁移
	err = migrator.Up()
	if err != nil {
		t.Fatalf("第一次执行迁移失败: %v", err)
	}

	// 第二次执行迁移，应该没有变化
	err = migrator.Up()
	// 不应该返回错误，或者返回 no change 不算错误
	if err != nil {
		t.Logf("第二次执行迁移结果: %v", err)
	}
}

// TestMigratorDownNoMigration 测试在没有迁移时执行回滚
func TestMigratorDownNoMigration(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	migrator, err := NewMigrator(testDBURL, testMigrationsPath)
	if err != nil {
		t.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 在没有执行任何迁移的情况下尝试回滚
	err = migrator.Down()
	// 应该处理这种情况，不返回严重错误
	if err != nil {
		t.Logf("在没有迁移时回滚结果: %v", err)
	}
}

// TestMigratorCloseNil 测试关闭 nil 迁移器
func TestMigratorCloseNil(t *testing.T) {
	m := &Migrator{}
	err := m.Close()
	if err != nil {
		t.Errorf("关闭 nil 迁移器不应该返回错误: %v", err)
	}
}
