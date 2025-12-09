package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podvibe/internal/config"
	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Podcast{},
		&models.Episode{},
		&models.EpisodeTag{},
		&models.Follow{},
		&models.EpisodeLike{},
		&models.EpisodeComment{},
		&models.Playlist{},
		&models.PlaylistItem{},
		&models.RefreshToken{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createUser(t *testing.T, repo *repositories.UserRepository, email, username string) *models.User {
	t.Helper()
	u := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: "hash",
		DisplayName:  username,
		Role:         "user",
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func createPodcast(t *testing.T, repo *repositories.PodcastRepository, ownerID uint, title string) *models.Podcast {
	t.Helper()
	p := &models.Podcast{
		OwnerID: ownerID,
		Title:   title,
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create podcast: %v", err)
	}
	return p
}

type stubTranscript struct {
	done chan struct{}
	text string
}

func (s *stubTranscript) Transcribe(ctx context.Context, audioPath string) (string, error) {
	if s.done != nil {
		close(s.done)
	}
	return s.text, nil
}

func TestAuthService_Flow(t *testing.T) {
	db := newTestDB(t)
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	cfg := config.Config{
		JWTSecret:       "secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	}
	service := NewAuthService(userRepo, tokenRepo, cfg)

	res, err := service.Register("a@example.com", "alice", "password", "Alice")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatalf("expected tokens")
	}

	loginRes, err := service.Login("a@example.com", "password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginRes.User.ID != res.User.ID {
		t.Fatalf("login user mismatch")
	}

	refreshRes, err := service.Refresh(loginRes.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshRes.AccessToken == "" {
		t.Fatalf("empty access token after refresh")
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	db := newTestDB(t)
	userRepo := repositories.NewUserRepository(db)
	svc := NewUserService(userRepo)
	u := createUser(t, userRepo, "b@example.com", "bob")

	updated, err := svc.UpdateProfile(u.ID, "Bobby", "About me")
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.DisplayName != "Bobby" || updated.Bio != "About me" {
		t.Fatalf("profile not updated: %+v", updated)
	}
}

func TestPodcastService_Permissions(t *testing.T) {
	db := newTestDB(t)
	podcastRepo := repositories.NewPodcastRepository(db)
	svc := NewPodcastService(podcastRepo)
	p := createPodcast(t, podcastRepo, 1, "MyCast")

	if err := svc.Update(2, false, p.ID, "New", "", ""); err == nil {
		t.Fatalf("expected forbidden for non-owner")
	}
	if err := svc.Update(p.OwnerID, false, p.ID, "New", "", ""); err != nil {
		t.Fatalf("owner update failed: %v", err)
	}
}

func TestEpisodeService_CreatePlayDelete(t *testing.T) {
	db := newTestDB(t)
	episodeRepo := repositories.NewEpisodeRepository(db)
	historyRepo := repositories.NewListeningHistoryRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	podcast := createPodcast(t, podcastRepo, 10, "Pod")
	tr := &stubTranscript{done: make(chan struct{}), text: "hello"}
	svc := NewEpisodeService(episodeRepo, tr, historyRepo)

	_, err := svc.Create(context.Background(), podcast.OwnerID, podcast.OwnerID+1, podcast.ID, "Ep1", "", "a.mp3", nil)
	if err == nil {
		t.Fatalf("expected forbidden create for non-owner")
	}

	ep, err := svc.Create(context.Background(), podcast.OwnerID, podcast.OwnerID, podcast.ID, "Ep1", "desc", "a.mp3", []string{"go"})
	if err != nil {
		t.Fatalf("create episode: %v", err)
	}
	<-tr.done
	time.Sleep(10 * time.Millisecond)
	saved, err := episodeRepo.FindByID(ep.ID)
	if err != nil {
		t.Fatalf("find episode: %v", err)
	}
	if saved.TranscriptStatus != "ready" || saved.Transcript == "" {
		t.Fatalf("transcript not updated: %+v", saved)
	}

	if err := svc.AddPlay(0, ep.ID); err != nil {
		t.Fatalf("add play: %v", err)
	}
	afterPlay, _ := episodeRepo.FindByID(ep.ID)
	if afterPlay.PlayCount == 0 {
		t.Fatalf("play count not incremented")
	}

	if err := svc.Delete(podcast.OwnerID+1, false, ep.ID); err == nil {
		t.Fatalf("expected forbidden delete")
	}
	if err := svc.Delete(podcast.OwnerID, false, ep.ID); err != nil {
		t.Fatalf("delete episode: %v", err)
	}
}

func TestLikeService_LikeUnlike(t *testing.T) {
	db := newTestDB(t)
	episodeRepo := repositories.NewEpisodeRepository(db)
	likeRepo := repositories.NewLikeRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	podcast := createPodcast(t, podcastRepo, 1, "Pod")
	ep := &models.Episode{PodcastID: podcast.ID, Title: "E1", AudioPath: "a.mp3"}
	if err := episodeRepo.Create(ep, nil); err != nil {
		t.Fatalf("create episode: %v", err)
	}
	svc := NewLikeService(likeRepo, episodeRepo)

	if err := svc.Like(2, ep.ID); err != nil {
		t.Fatalf("like: %v", err)
	}
	afterLike, _ := episodeRepo.FindByID(ep.ID)
	if afterLike.LikeCount != 1 {
		t.Fatalf("like count not incremented")
	}
	if err := svc.Unlike(2, ep.ID); err != nil {
		t.Fatalf("unlike: %v", err)
	}
	afterUnlike, _ := episodeRepo.FindByID(ep.ID)
	if afterUnlike.LikeCount != 0 {
		t.Fatalf("like count not decremented")
	}
}

func TestCommentService_CreateDelete(t *testing.T) {
	db := newTestDB(t)
	episodeRepo := repositories.NewEpisodeRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	podcast := createPodcast(t, podcastRepo, 3, "Pod")
	ep := &models.Episode{PodcastID: podcast.ID, Title: "E1", AudioPath: "a.mp3"}
	if err := episodeRepo.Create(ep, nil); err != nil {
		t.Fatalf("create episode: %v", err)
	}
	svc := NewCommentService(commentRepo, episodeRepo)

	if err := svc.Create(3, ep.ID, "nice"); err != nil {
		t.Fatalf("create comment: %v", err)
	}
	afterCreate, _ := episodeRepo.FindByID(ep.ID)
	if afterCreate.CommentCount != 1 {
		t.Fatalf("comment count not incremented")
	}
	comments, _, _ := commentRepo.ListByEpisode(ep.ID, 1, 10)
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment")
	}
	if err := svc.Delete(3, false, comments[0].ID); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	deleted, _ := commentRepo.FindByID(comments[0].ID)
	if !deleted.IsDeleted {
		t.Fatalf("comment not marked deleted")
	}
	afterDelete, _ := episodeRepo.FindByID(ep.ID)
	if afterDelete.CommentCount != 0 {
		t.Fatalf("comment count not decremented")
	}
}

func TestFollowService_Following(t *testing.T) {
	db := newTestDB(t)
	repo := repositories.NewFollowRepository(db)
	svc := NewFollowService(repo)

	if err := svc.Follow(1, 2); err != nil {
		t.Fatalf("follow: %v", err)
	}
	following, total, err := svc.Following(1, 1, 10)
	if err != nil {
		t.Fatalf("following: %v", err)
	}
	if total != 1 || len(following) != 1 {
		t.Fatalf("following not recorded")
	}
	if err := svc.Unfollow(1, 2); err != nil {
		t.Fatalf("unfollow: %v", err)
	}
}

func TestPlaylistService_Items(t *testing.T) {
	db := newTestDB(t)
	plRepo := repositories.NewPlaylistRepository(db)
	svc := NewPlaylistService(plRepo)
	pl, err := svc.Create(1, "Run", "desc")
	if err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	if err := svc.AddItem(pl.OwnerID, false, pl.ID, 5); err != nil {
		t.Fatalf("add item: %v", err)
	}
	items, total, err := svc.Items(pl.ID, 1, 10)
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("item not added")
	}
	if err := svc.RemoveItem(pl.OwnerID, false, pl.ID, items[0].ID); err != nil {
		t.Fatalf("remove item: %v", err)
	}
}

func TestFeedService_Feed(t *testing.T) {
	db := newTestDB(t)
	followRepo := repositories.NewFollowRepository(db)
	feedRepo := repositories.NewFeedRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	userRepo := repositories.NewUserRepository(db)
	episodeRepo := repositories.NewEpisodeRepository(db)

	u1 := createUser(t, userRepo, "u1@example.com", "u1")
	u2 := createUser(t, userRepo, "u2@example.com", "u2")
	_ = followRepo.Follow(u1.ID, u2.ID)
	pod := createPodcast(t, podcastRepo, u2.ID, "Pod")
	ep := &models.Episode{PodcastID: pod.ID, Title: "E1", AudioPath: "a.mp3"}
	if err := episodeRepo.Create(ep, nil); err != nil {
		t.Fatalf("create episode: %v", err)
	}

	svc := NewFeedService(followRepo, feedRepo, podcastRepo, userRepo)
	items, total, err := svc.Feed(u1.ID, 1, 10)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 feed item, got %d", len(items))
	}
	if items[0].Author.ID != u2.ID || items[0].Podcast.ID != pod.ID {
		t.Fatalf("feed item mismatch")
	}
}

func TestRecommendationService_Order(t *testing.T) {
	db := newTestDB(t)
	followRepo := repositories.NewFollowRepository(db)
	feedRepo := repositories.NewFeedRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	userRepo := repositories.NewUserRepository(db)
	episodeRepo := repositories.NewEpisodeRepository(db)

	user := createUser(t, userRepo, "me@example.com", "me")
	followed := createUser(t, userRepo, "f@example.com", "followed")
	other := createUser(t, userRepo, "o@example.com", "other")
	_ = followRepo.Follow(user.ID, followed.ID)

	followedPod := createPodcast(t, podcastRepo, followed.ID, "Fav")
	otherPod := createPodcast(t, podcastRepo, other.ID, "Other")

	newerFollow := &models.Episode{PodcastID: followedPod.ID, Title: "new", AudioPath: "n.mp3", PublishedAt: time.Now().Add(-10 * time.Minute)}
	olderFollow := &models.Episode{PodcastID: followedPod.ID, Title: "old", AudioPath: "o.mp3", PublishedAt: time.Now().Add(-2 * time.Hour)}
	newestOther := &models.Episode{PodcastID: otherPod.ID, Title: "other", AudioPath: "x.mp3", PublishedAt: time.Now()}
	for _, ep := range []*models.Episode{newerFollow, olderFollow, newestOther} {
		if err := episodeRepo.Create(ep, nil); err != nil {
			t.Fatalf("create episode: %v", err)
		}
	}

	svc := NewRecommendationService(followRepo, feedRepo, podcastRepo, userRepo)
	items, total, err := svc.Recommend(user.ID, 1, 10)
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if total != 3 || len(items) != 3 {
		t.Fatalf("expected 3 recommendations, got %d total %d", len(items), total)
	}
	if items[0].Episode.Title != "new" || items[1].Episode.Title != "old" || items[2].Episode.Title != "other" {
		t.Fatalf("unexpected ordering: %+v", []string{items[0].Episode.Title, items[1].Episode.Title, items[2].Episode.Title})
	}
	if items[0].Reason != "from_follow" || items[2].Reason != "fresh" {
		t.Fatalf("unexpected reasons: %s, %s", items[0].Reason, items[2].Reason)
	}
}

func TestAdminService_BanUnban(t *testing.T) {
	db := newTestDB(t)
	userRepo := repositories.NewUserRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	episodeRepo := repositories.NewEpisodeRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	historyRepo := repositories.NewListeningHistoryRepository(db)

	userSvc := NewUserService(userRepo)
	podcastSvc := NewPodcastService(podcastRepo)
	episodeSvc := NewEpisodeService(episodeRepo, &stubTranscript{text: "x"}, historyRepo)
	commentSvc := NewCommentService(commentRepo, episodeRepo)
	adminSvc := NewAdminService(userSvc, podcastSvc, episodeSvc, commentSvc)

	u := createUser(t, userRepo, "admin@test.com", "ad")
	if err := adminSvc.BanUser(u.ID); err != nil {
		t.Fatalf("ban: %v", err)
	}
	banned, _ := userRepo.FindByID(u.ID)
	if !banned.IsBanned {
		t.Fatalf("user not banned")
	}
	if err := adminSvc.UnbanUser(u.ID); err != nil {
		t.Fatalf("unban: %v", err)
	}
}
