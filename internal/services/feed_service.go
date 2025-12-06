package services

import (
	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type FeedItem struct {
	Episode models.Episode `json:"episode"`
	Author  models.User    `json:"author"`
	Podcast models.Podcast `json:"podcast"`
}

type FeedService struct {
	follows  *repositories.FollowRepository
	feedRepo *repositories.FeedRepository
	podcasts *repositories.PodcastRepository
	users    *repositories.UserRepository
}

func NewFeedService(follows *repositories.FollowRepository, feed *repositories.FeedRepository, podcasts *repositories.PodcastRepository, users *repositories.UserRepository) *FeedService {
	return &FeedService{
		follows:  follows,
		feedRepo: feed,
		podcasts: podcasts,
		users:    users,
	}
}

func (s *FeedService) Feed(userID uint, page, pageSize int) ([]FeedItem, int64, error) {
	following, err := s.follows.FollowingIDs(userID)
	if err != nil {
		return nil, 0, err
	}
	episodes, total, err := s.feedRepo.Feed(following, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]FeedItem, 0, len(episodes))
	userCache := map[uint]models.User{}
	podcastCache := map[uint]models.Podcast{}

	for _, ep := range episodes {
		podcast, ok := podcastCache[ep.PodcastID]
		if !ok {
			p, err := s.podcasts.FindByID(ep.PodcastID)
			if err != nil {
				continue
			}
			podcast = *p
			podcastCache[ep.PodcastID] = podcast
		}
		author, ok := userCache[podcast.OwnerID]
		if !ok {
			u, err := s.users.FindByID(podcast.OwnerID)
			if err != nil {
				continue
			}
			author = *u
			userCache[podcast.OwnerID] = author
		}
		items = append(items, FeedItem{
			Episode: ep,
			Author:  author,
			Podcast: podcast,
		})
	}
	return items, total, nil
}
