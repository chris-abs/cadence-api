package recent

import "github.com/chrisabs/cadence/internal/models"

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetRecentEntities(familyID models.FamilyID) (*Response, error) {
    const defaultLimit = 10
    return s.repo.GetRecentEntities(familyID, defaultLimit)
}