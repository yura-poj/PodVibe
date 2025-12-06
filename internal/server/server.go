package server

import (
	"log"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/handlers"
)

// New builds gin router with all routes registered.
func New(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	podcastHandler *handlers.PodcastHandler,
	episodeHandler *handlers.EpisodeHandler,
	likeHandler *handlers.LikeHandler,
	commentHandler *handlers.CommentHandler,
	followHandler *handlers.FollowHandler,
	feedHandler *handlers.FeedHandler,
	playlistHandler *handlers.PlaylistHandler,
	adminHandler *handlers.AdminHandler,
	jwtSecret string,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Public endpoints
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/refresh", authHandler.Refresh)

	router.GET("/users/:id", userHandler.Get)
	router.GET("/users/search", userHandler.Search)

	router.GET("/podcasts/:id", podcastHandler.Get)
	router.GET("/users/:id/podcasts", podcastHandler.ListByUser)

	router.GET("/episodes/:id", episodeHandler.Get)
	router.GET("/podcasts/:podcast_id/episodes", episodeHandler.ListForPodcast)
	router.GET("/episodes/popular", episodeHandler.Popular)
	router.POST("/episodes/:id/plays", episodeHandler.Play)

	router.GET("/episodes/:id/comments", commentHandler.List)

	router.GET("/users/:id/followers", followHandler.Followers)
	router.GET("/users/:id/following", followHandler.Following)

	router.GET("/playlists/:id", playlistHandler.Get)
	router.GET("/users/:id/playlists", playlistHandler.ListByUser)

	// Protected routes
	authRequired := router.Group("/")
	authRequired.Use(auth.Middleware(jwtSecret))
	{
		authRequired.GET("/users/me", userHandler.Me)
		authRequired.PATCH("/users/me", userHandler.UpdateMe)

		authRequired.POST("/podcasts", podcastHandler.Create)
		authRequired.PATCH("/podcasts/:id", podcastHandler.Update)
		authRequired.DELETE("/podcasts/:id", podcastHandler.Delete)

		authRequired.POST("/podcasts/:podcast_id/episodes", episodeHandler.Create)

		authRequired.POST("/episodes/:id/like", likeHandler.Like)
		authRequired.DELETE("/episodes/:id/like", likeHandler.Unlike)

		authRequired.POST("/episodes/:id/comments", commentHandler.Create)
		authRequired.DELETE("/comments/:id", commentHandler.Delete)

		authRequired.POST("/users/:id/follow", followHandler.Follow)
		authRequired.DELETE("/users/:id/follow", followHandler.Unfollow)

		authRequired.GET("/feed", feedHandler.Feed)

		authRequired.POST("/playlists", playlistHandler.Create)
		authRequired.POST("/playlists/:id/items", playlistHandler.AddItem)
		authRequired.DELETE("/playlists/:id/items/:item_id", playlistHandler.RemoveItem)
	}

	admin := router.Group("/admin")
	admin.Use(auth.Middleware(jwtSecret), auth.RequireRole("admin"))
	{
		admin.GET("/users", adminHandler.ListUsers)
		admin.POST("/users/:id/ban", adminHandler.BanUser)
		admin.POST("/users/:id/unban", adminHandler.UnbanUser)
		admin.DELETE("/episodes/:id", adminHandler.DeleteEpisode)
		admin.DELETE("/podcasts/:id", adminHandler.DeletePodcast)
	}

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("cannot set trusted proxies: %v", err)
	}

	return router
}
