package classification

import (
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/media/classification/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateClassification(req *CreateClassificationRequest, familyID models.FamilyID, createdBy models.ProfileID) (*entities.Classification, error) {
	if err := s.validateCreateClassificationRequest(req); err != nil {
		return nil, err
	}
	
	now := time.Now().UTC()
	classification := &entities.Classification{
		ID:          models.NewClassificationID(),
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		ImageURL:    "", 
		FamilyID:    familyID,
		ProfileID:   createdBy,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
	}
	
	if err := s.repo.Create(classification); err != nil {
		return nil, fmt.Errorf("error creating classification: %v", err)
	}
	
	return classification, nil
}

func (s *Service) GetClassificationByID(classificationID models.ClassificationID, profileID models.ProfileID) (*entities.Classification, error) {
	classification, err := s.repo.GetByID(classificationID, profileID)
	if err != nil {
		return nil, err
	}
	
	return classification, nil
}

func (s *Service) UpdateClassification(classificationID models.ClassificationID, profileID models.ProfileID, req *UpdateClassificationRequest) (*entities.Classification, error) {
	if err := s.validateUpdateClassificationRequest(req); err != nil {
		return nil, err
	}
	
	existingClassification, err := s.repo.GetByID(classificationID, profileID)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		existingClassification.Name = *req.Name
	}
	if req.Description != nil {
		existingClassification.Description = *req.Description
	}
	if req.Color != nil {
		existingClassification.Color = *req.Color
	}
	
	existingClassification.UpdatedAt = time.Now().UTC()
	
	if err := s.repo.Update(existingClassification, profileID); err != nil {
		return nil, fmt.Errorf("error updating classification: %v", err)
	}
	
	return existingClassification, nil
}

func (s *Service) UpdateClassificationImage(classificationID models.ClassificationID, profileID models.ProfileID, imageURL string) (*entities.Classification, error) {
	existingClassification, err := s.repo.GetByID(classificationID, profileID)
	if err != nil {
		return nil, err
	}
	
	existingClassification.ImageURL = imageURL
	existingClassification.UpdatedAt = time.Now().UTC()
	
	if err := s.repo.Update(existingClassification, profileID); err != nil {
		return nil, fmt.Errorf("error updating classification image: %v", err)
	}
	
	return existingClassification, nil
}

func (s *Service) GetAllClassifications(profileID models.ProfileID, params ClassificationSearchRequest) (*ClassificationSearchResponse, error) {
	limit := 50 
	if params.Limit != nil {
		limit = *params.Limit
	}
	
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}
	
	classifications, total, err := s.repo.GetAllByProfile(profileID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting classifications: %v", err)
	}
	
	var classificationEntities []entities.Classification
	for _, c := range classifications {
		classificationEntities = append(classificationEntities, *c)
	}
	
	hasMore := (offset + limit) < total
	
	return &ClassificationSearchResponse{
		Data:   classificationEntities,
		Total:  total,
		Limit:  limit,
		Offset: offset,
		HasMore: hasMore,
	}, nil
}

func (s *Service) DeleteClassification(classificationID models.ClassificationID, profileID models.ProfileID, deletedBy models.ProfileID) error {
	// Verify classification exists and belongs to the profile
	_, err := s.repo.GetByID(classificationID, profileID)
	if err != nil {
		return err
	}
	
	count, err := s.repo.GetMaterialCountByClassification(classificationID, profileID)
	if err != nil {
		return fmt.Errorf("error checking classification usage: %v", err)
	}
	
	if count > 0 {
		return fmt.Errorf("cannot delete classification: it is being used by %d material items", count)
	}
	
	return s.repo.Delete(classificationID, profileID, deletedBy)
}

func (s *Service) validateCreateClassificationRequest(req *CreateClassificationRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(req.Name) > 100 {
		return fmt.Errorf("name must be 100 characters or less")
	}
	
	if len(req.Description) > 500 {
		return fmt.Errorf("description must be 500 characters or less")
	}
	
	if req.Color == "" {
		return fmt.Errorf("color is required")
	}
	
	return nil
}

func (s *Service) validateUpdateClassificationRequest(req *UpdateClassificationRequest) error {
	if req.Name != nil && len(*req.Name) > 100 {
		return fmt.Errorf("name must be 100 characters or less")
	}
	
	if req.Description != nil && len(*req.Description) > 500 {
		return fmt.Errorf("description must be 500 characters or less")
	}
	
	return nil
}
