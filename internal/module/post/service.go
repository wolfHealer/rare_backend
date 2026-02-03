package post

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreatePost(post *Post) error {
	post.Status = "pending"
	return s.repo.Insert(post)
}
