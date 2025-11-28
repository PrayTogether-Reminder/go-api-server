package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/config"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/logger"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/fcm_token"
	"gorm.io/gorm"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMGateway handles Firebase Cloud Messaging operations
type FCMGateway struct {
	client       *messaging.Client
	db           *gorm.DB
	cfg          *config.Config
	fcmTokenRepo *fcm_token.FcmTokenRepository
}

// NewFCMGateway creates a new FCM gateway instance
func NewFCMGateway(cfg *config.Config, db *gorm.DB, fcmTokenRepo *fcm_token.FcmTokenRepository) (*FCMGateway, error) {
	slog.Info("FCM 활성화 상태", "enabled", cfg.FCM.Enabled)
	if !cfg.FCM.Enabled {
		return &FCMGateway{
			client:       nil,
			db:           db,
			cfg:          cfg,
			fcmTokenRepo: fcmTokenRepo,
		}, nil
	}

	if cfg.FCM.CredentialsPath == "" {
		return nil, fmt.Errorf("FCM_CREDENTIALS_PATH가 설정되지 않았습니다")
	}

	opt := option.WithCredentialsFile(cfg.FCM.CredentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase 앱 초기화 실패: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fcm 클라이언트 초기화 실패: %w", err)
	}

	slog.Info("FCM Gateway 초기화 완료", "credentials_path", cfg.FCM.CredentialsPath)

	return &FCMGateway{
		client:       client,
		db:           db,
		cfg:          cfg,
		fcmTokenRepo: fcmTokenRepo,
	}, nil
}

// SendPrayerCompletionNotification sends push notification to room members asynchronously
func (g *FCMGateway) SendPrayerCompletionNotification(
	ctx context.Context,
	recipientIDs []int64,
	roomID int64,
	prayerTitleID int64,
	prayerTitleText string,
	title string,
	body string,
) {
	if !g.cfg.FCM.Enabled || g.client == nil {
		logger.FromContext(ctx).Warn("FCM이 비활성화되어 있어 푸시 알림을 전송하지 않습니다")
		return
	}

	// create a detached context so DB/FCM calls keep request metadata but aren't canceled immediately
	notificationCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	notificationCtx = logger.WithLogger(notificationCtx, logger.FromContext(ctx))
	if memberID := ctx.Value(sharedHttp.ContextKey()); memberID != nil {
		notificationCtx = context.WithValue(notificationCtx, sharedHttp.ContextKey(), memberID)
	}

	// Run asynchronously using goroutine
	go func(asyncCtx context.Context) {
		defer cancel()

		log := logger.FromContext(asyncCtx)

		tokens, err := g.fcmTokenRepo.FindTokensByMemberIDs(asyncCtx, g.db, recipientIDs)
		if err != nil {
			log.Error("FCM 토큰 조회 실패", "error", err)
			return
		}

		if len(tokens) == 0 {
			log.Warn("전송할 FCM 토큰이 없습니다", "recipient_ids", recipientIDs)
			return
		}

		message := &messaging.MulticastMessage{
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: map[string]string{
				"Title":         title,
				"Body":          body,
				"roomId":        fmt.Sprintf("%d", roomID),
				"prayerTitle":   prayerTitleText,
				"prayerTitleId": fmt.Sprintf("%d", prayerTitleID),
			},
			Tokens: tokens,
		}

		response, err := g.client.SendEachForMulticast(asyncCtx, message)
		if err != nil {
			log.Error("FCM 메시지 전송 실패", "error", err)
			return
		}

		log.Info("FCM 푸시 알림 전송 완료",
			"success_count", response.SuccessCount,
			"failure_count", response.FailureCount,
		)

		// Handle failed tokens (delete invalid tokens)
		if response.FailureCount > 0 {
			g.handleFailedTokens(asyncCtx, tokens, response)
		}
	}(notificationCtx)
}

// handleFailedTokens removes invalid FCM tokens from database
func (g *FCMGateway) handleFailedTokens(ctx context.Context, tokens []string, response *messaging.BatchResponse) {
	log := logger.FromContext(ctx)

	for idx, resp := range response.Responses {
		if !resp.Success {
			token := tokens[idx]
			log.Warn("FCM 토큰 전송 실패 - 토큰 삭제 시도",
				"token", token, "error", resp.Error,
			)

			if err := g.fcmTokenRepo.DeleteByToken(ctx, g.db, token); err != nil {
				log.Error("실패한 FCM 토큰 삭제 실패", "token", token, "error", err)
			} else {
				log.Info("유효하지 않은 FCM 토큰 삭제 완료", "token", token)
			}
		}
	}
}
