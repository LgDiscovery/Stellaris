package service

import (
	"kubeai/internal/model"
	"kubeai/internal/modelmanage/repository"
)

type ModelService struct {
	repo *repository.ModelRepository
}

func NewModelService(repo *repository.ModelRepository) *ModelService {
	return &ModelService{repo: repo}
}

func (s *ModelService) Create(name, description string) (*model.Model, error) {
	m := &model.Model{
		Name:        name,
		Description: description,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *ModelService) GetByID(id uint) (*model.Model, error) {
	m, err := s.repo.GetByID(id)
	return m, err
}

func (s *ModelService) List(page, size int) ([]model.Model, int64, error) {
	return s.repo.List(page, size)
}

func (s *ModelService) Update(id uint, description string) (*model.Model, error) {
	m, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	m.Description = description
	if err := s.repo.Update(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *ModelService) Delete(id uint) error {
	return s.repo.Delete(id)
}
