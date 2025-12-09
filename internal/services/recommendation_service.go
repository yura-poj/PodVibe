package services

import (
	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type RecommendationItem struct {
	Episode models.Episode `json:"episode"`
	Author  models.User    `json:"author"`
	Podcast models.Podcast `json:"podcast"`
	Reason  string         `json:"reason"`
}

type RecommendationService struct {
	follows  *repositories.FollowRepository
	feedRepo *repositories.FeedRepository
	podcasts *repositories.PodcastRepository
	users    *repositories.UserRepository
}

func NewRecommendationService(follows *repositories.FollowRepository, feed *repositories.FeedRepository, podcasts *repositories.PodcastRepository, users *repositories.UserRepository) *RecommendationService {
	return &RecommendationService{
		follows:  follows,
		feedRepo: feed,
		podcasts: podcasts,
		users:    users,
	}
}

func (s *RecommendationService) Recommend(userID uint, page, pageSize int) ([]RecommendationItem, int64, error) {
	following, err := s.follows.FollowingIDs(userID)
	if err != nil {
		return nil, 0, err
	}

	episodes, total, err := s.feedRepo.Recommendations(following, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	followItems := make([]RecommendationItem, 0, len(episodes))
	otherItems := make([]RecommendationItem, 0, len(episodes))
	userCache := map[uint]models.User{}
	podcastCache := map[uint]models.Podcast{}
	followingSet := map[uint]struct{}{}

	for _, id := range following {
		followingSet[id] = struct{}{}
	}

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

		reason := "fresh"
		if _, ok := followingSet[podcast.OwnerID]; ok {
			reason = "from_follow"
		}

		item := RecommendationItem{
			Episode: ep,
			Author:  author,
			Podcast: podcast,
			Reason:  reason,
		}

		if reason == "from_follow" {
			followItems = append(followItems, item)
		} else {
			otherItems = append(otherItems, item)
		}
	}

	return append(followItems, otherItems...), total, nil
}
