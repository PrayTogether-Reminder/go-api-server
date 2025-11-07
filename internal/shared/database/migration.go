package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"

	"gorm.io/gorm"
)

// Migrate executes database migration based on configuration
func Migrate(db *gorm.DB, cfg *config.Config) error {
	if !cfg.Database.IsAutoMigrate {
		slog.Info("⏭️  데이터베이스 마이그레이션 비활성화됨",
			"auto_migrate", false, "env", cfg.App.Env,
		)
		return nil
	}

	slog.Warn("🔧 데이터베이스 마이그레이션 시작 - 모든 테이블이 삭제되고 재생성됩니다!",
		"auto_migrate", true, "env", cfg.App.Env,
	)

	// Safety check: prevent accidental data loss in production
	if cfg.App.Env == "prod" || cfg.App.Env == "production" {
		return fmt.Errorf("🚨 PRODUCTION 환경에서는 DB_AUTO_MIGRATE=true를 사용할 수 없습니다! 데이터 손실 방지를 위해 차단됨")
	}

	// Get all models in dependency order (중요: FK 관계 역순으로 삭제)
	models := getAllModelsInReverseOrder()

	// Step 1: Drop all tables
	slog.Info("🗑️  기존 테이블 삭제 중...")
	for _, model := range models {
		if err := db.Migrator().DropTable(model); err != nil {
			// Ignore "table does not exist" errors
			slog.Debug("테이블 삭제 시도", "model", fmt.Sprintf("%T", model), "error", err)
		}
	}

	// Step 2: Create tables with IsAutoMigrate
	slog.Info("📦 새 테이블 생성 중...")
	if err := runAutoMigrate(db); err != nil {
		return fmt.Errorf("테이블 생성 실패: %w", err)
	}

	// Step 3: Create sequences (Oracle specific)
	if err := createOracleSequences(db); err != nil {
		return fmt.Errorf("시퀀스 생성 실패: %w", err)
	}

	slog.Info("✅ 마이그레이션 완료!")
	return nil
}

// runAutoMigrate creates tables based on model definitions
func runAutoMigrate(db *gorm.DB) error {
	// 중요: 의존성 순서대로 생성 (FK 참조 순서)
	// 1. 독립 테이블 먼저
	// 2. FK 참조하는 테이블은 나중에
	models := []interface{}{
		// Independent tables (no foreign keys)
		&model.Member{},
		&model.Room{},

		// Join tables (have foreign keys)
		&model.MemberRoom{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("%T 마이그레이션 실패: %w", m, err)
		}
		slog.Debug("테이블 생성됨", "model", fmt.Sprintf("%T", m))
	}

	return nil
}

// getAllModelsInReverseOrder returns all models in reverse dependency order
// Used for dropping tables (FK constraints require reverse order)
func getAllModelsInReverseOrder() []interface{} {
	return []interface{}{
		&model.MemberRoom{},

		// Drop independent tables last
		&model.Room{},
		&model.Member{},
	}
}

// createOracleSequences creates Oracle sequences for auto-increment IDs
func createOracleSequences(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sequences := []struct {
		name       string
		startValue int64
	}{
		{"MEMBER_SEQ", 1},
		{"ROOM_SEQ", 1},
		{"MEMBER_ROOM_SEQ", 1},
	}

	for _, seq := range sequences {
		// Drop sequence if exists
		dropSQL := fmt.Sprintf("BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE %s'; EXCEPTION WHEN OTHERS THEN NULL; END;", seq.name)
		if err := db.WithContext(ctx).Exec(dropSQL).Error; err != nil {
			slog.Debug("시퀀스 삭제 시도", "sequence", seq.name, "error", err)
		}

		// Create sequence
		createSQL := fmt.Sprintf("CREATE SEQUENCE %s START WITH %d INCREMENT BY 1 NOCACHE", seq.name, seq.startValue)
		if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
			return fmt.Errorf("시퀀스 %s 생성 실패: %w", seq.name, err)
		}
		slog.Debug("시퀀스 생성됨", "sequence", seq.name)
	}

	return nil
}
