package recent

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetRecentEntities(userID int) (*Response, error) {
    const defaultLimit = 10
    return s.repo.GetRecentEntities(userID, defaultLimit)
}