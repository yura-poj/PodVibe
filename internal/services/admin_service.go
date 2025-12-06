package services

type AdminService struct {
	users    *UserService
	podcasts *PodcastService
	episodes *EpisodeService
	comments *CommentService
}

func NewAdminService(users *UserService, podcasts *PodcastService, episodes *EpisodeService, comments *CommentService) *AdminService {
	return &AdminService{
		users:    users,
		podcasts: podcasts,
		episodes: episodes,
		comments: comments,
	}
}

func (s *AdminService) BanUser(id uint) error {
	return s.users.BanUser(id)
}

func (s *AdminService) UnbanUser(id uint) error {
	return s.users.UnbanUser(id)
}

func (s *AdminService) DeletePodcast(id uint) error {
	return s.podcasts.Delete(0, true, id)
}

func (s *AdminService) DeleteEpisode(id uint) error {
	return s.episodes.Delete(0, true, id)
}

func (s *AdminService) DeleteComment(id uint, episodeOwner uint) error {
	return s.comments.Delete(0, true, id)
}
