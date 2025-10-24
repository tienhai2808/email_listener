package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type serviceImpl struct {
	repo      repository
	oauthCfg  *oauth2.Config
	psClient  *pubsub.Client
	psTopicID string
	psSubID   string
	projectID string
}

func newService(repo repository, cfg *config, psClient *pubsub.Client) service {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  "postmessage",
		Scopes:       cfg.Google.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &serviceImpl{
		repo,
		oauthCfg,
		psClient,
		cfg.Google.TopicID,
		cfg.Google.SubscriptionID,
		cfg.Google.ProjectID,
	}
}

func (s *serviceImpl) login(ctx context.Context, code string) (*user, error) {
	token, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("đổi authorization code thất bại: %w", err)
	}

	client := s.oauthCfg.Client(ctx, token)
	gmailSvc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("tạo gmail service thất bại: %w", err)
	}

	profile, err := gmailSvc.Users.GetProfile("me").Do()
	if err != nil {
		return nil, fmt.Errorf("lấy profile người dùng thất bại: %w", err)
	}

	user, err := s.repo.findOrCreate(ctx, profile.EmailAddress, "Unknown Name", "")
	if err != nil {
		return nil, fmt.Errorf("lỗi xử lý người dùng trong DB: %w", err)
	}

	if err := s.repo.saveToken(ctx, user.Email, token); err != nil {
		return nil, err
	}

	if err := s.repo.saveHistoryID(ctx, user.Email, profile.HistoryId); err != nil {
		log.Printf("[Warning] Không lưu được historyId: %v", err)
	}

	go s.startWatching(user.Email, token)

	log.Printf("Xử lý đăng nhập thành công cho: %s", user.Email)

	return user, nil
}

func (s *serviceImpl) startWatching(email string, token *oauth2.Token) {
	ctx := context.Background()
	client := s.oauthCfg.Client(ctx, token)
	gmailSvc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("[Watch Error] %s: không thể tạo gmail service: %v", email, err)
		return
	}

	topicName := fmt.Sprintf("projects/%s/topics/%s", s.projectID, s.psTopicID)
	watchReq := &gmail.WatchRequest{
		LabelIds:  []string{"INBOX"},
		TopicName: topicName,
	}
	if _, err := gmailSvc.Users.Watch("me", watchReq).Do(); err != nil {
		log.Printf("[Watch Error] %s: không thể bật watch: %v", email, err)
		return
	}
	log.Printf("Đã bật watch thành công cho: %s", email)
}

func (s *serviceImpl) listenForNotifications(ctx context.Context) {
	sub := s.psClient.Subscription(s.psSubID)
	log.Printf("Bắt đầu Goroutine lắng nghe Pub/Sub subscription: %s", s.psSubID)

	if err := sub.Receive(ctx, func(cctx context.Context, msg *pubsub.Message) {
		msg.Ack()

		var notification struct {
			EmailAddress string `json:"emailAddress"`
			HistoryID    uint64 `json:"historyId"`
		}
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			log.Printf("[PubSub Error] Lỗi giải mã tin nhắn: %v", err)
			return
		}

		log.Printf(">>> Có email mới cho: %s | HistoryID: %d", notification.EmailAddress, notification.HistoryID)

		s.processNotification(notification.EmailAddress, notification.HistoryID)
	}); err != nil {
		log.Fatalf("Goroutine lắng nghe Pub/Sub đã dừng: %v", err)
	}
}

func (s *serviceImpl) processNotification(email string, newHistoryID uint64) {
	ctx := context.Background()

	token, err := s.repo.getToken(ctx, email)
	if err != nil {
		log.Printf("[Process Error] %s: không tìm thấy token: %v", email, err)
		return
	}

	lastHistoryID, err := s.repo.getHistoryID(ctx, email)
	if err != nil {
		log.Printf("[Process Error] %s: không lấy được historyId cũ: %v", email, err)
		return
	}

	client := s.oauthCfg.Client(ctx, token)
	gmailSvc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("[Process Error] %s: không thể tạo gmail service: %v", email, err)
		return
	}

	historyList, err := gmailSvc.Users.History.List("me").StartHistoryId(lastHistoryID).Do()
	if err != nil {
		log.Printf("[Process Error] %s: không thể lấy history: %v", email, err)
		return
	}

	for _, history := range historyList.History {
		for _, msgAdded := range history.MessagesAdded {
			msg, err := gmailSvc.Users.Messages.Get("me", msgAdded.Message.Id).Format("full").Do()
			if err != nil {
				continue
			}
			var subject string
			for _, h := range msg.Payload.Headers {
				if h.Name == "Subject" {
					subject = h.Value
					break
				}
			}

			var body string
			if msg.Payload.Parts == nil {
				decodedBody, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
				if err == nil {
					body = string(decodedBody)
				}
			} else {
				for _, part := range msg.Payload.Parts {
					if part.MimeType == "text/plain" {
						decodedBody, err := base64.URLEncoding.DecodeString(part.Body.Data)
						if err == nil {
							body = string(decodedBody)
							break
						}
					}
				}
			}

			log.Printf("--- EMAIL MỚI: [%s]", email)
      log.Printf("   Tiêu đề: %s", subject)
      log.Printf("   Nội dung: %s", body)
		}
	}

	if err := s.repo.saveHistoryID(ctx, email, newHistoryID); err != nil {
		log.Printf("[Warning] %s: không lưu được historyId mới: %v", email, err)
	}
}
