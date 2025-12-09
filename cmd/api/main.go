package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"podvibe/internal/config"
	"podvibe/internal/handlers"
	"podvibe/internal/models"
	"podvibe/internal/repositories"
	"podvibe/internal/server"
	"podvibe/internal/services"
	"podvibe/internal/storage"
	"podvibe/internal/transcript"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
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
		&models.ListeningHistory{},
	); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	store, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		logger.Fatal("failed to init storage", zap.Error(err))
	}

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	podcastRepo := repositories.NewPodcastRepository(db)
	episodeRepo := repositories.NewEpisodeRepository(db)
	historyRepo := repositories.NewListeningHistoryRepository(db)
	followRepo := repositories.NewFollowRepository(db)
	likeRepo := repositories.NewLikeRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	playlistRepo := repositories.NewPlaylistRepository(db)
	feedRepo := repositories.NewFeedRepository(db)

	// Services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, tokenRepo, cfg)
	podcastService := services.NewPodcastService(podcastRepo)
	transcriptService := transcript.NewDummyService()
	episodeService := services.NewEpisodeService(episodeRepo, transcriptService, historyRepo)
	likeService := services.NewLikeService(likeRepo, episodeRepo)
	commentService := services.NewCommentService(commentRepo, episodeRepo)
	followService := services.NewFollowService(followRepo)
	playlistService := services.NewPlaylistService(playlistRepo)
	feedService := services.NewFeedService(followRepo, feedRepo, podcastRepo, userRepo)
	recommendationService := services.NewRecommendationService(followRepo, feedRepo, podcastRepo, userRepo)
	adminService := services.NewAdminService(userService, podcastService, episodeService, commentService)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	podcastHandler := handlers.NewPodcastHandler(podcastService, episodeService, store)
	episodeHandler := handlers.NewEpisodeHandler(episodeService, podcastService, store, cfg.JWTSecret)
	likeHandler := handlers.NewLikeHandler(likeService)
	commentHandler := handlers.NewCommentHandler(commentService, episodeService)
	followHandler := handlers.NewFollowHandler(followService)
	feedHandler := handlers.NewFeedHandler(feedService)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationService)
	playlistHandler := handlers.NewPlaylistHandler(playlistService)
	adminHandler := handlers.NewAdminHandler(adminService, userService)

	gin.SetMode(gin.ReleaseMode)
	router := server.New(
		authHandler,
		userHandler,
		podcastHandler,
		episodeHandler,
		likeHandler,
		commentHandler,
		followHandler,
		feedHandler,
		recommendationHandler,
		playlistHandler,
		adminHandler,
		cfg.JWTSecret,
	)
	router.Static("/static", cfg.StoragePath)

	addr := ":" + cfg.AppPort
	logger.Info("Starting server", zap.String("addr", addr))
	if err := router.Run(addr); err != nil {
		logger.Fatal("server stopped", zap.Error(err))
		os.Exit(1)
	}
}
