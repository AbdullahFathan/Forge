package department

import (
	"strings"

	"github.com/google/uuid"

	"workspace/pkg/apperr"
)

type Store interface {
	Create(*Department) error
	Save(*Department) error
	GetByID(uuid.UUID) (*Department, error)
	NameTaken(name string, excludeID *uuid.UUID) (bool, error)
	List(ListFilter) ([]Department, int64, error)
	SoftDelete(uuid.UUID) error
}

type Service struct {
	repo Store
}

func NewService(repo Store) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(name string) (*Department, error) {
	name = strings.TrimSpace(name)
	taken, err := s.repo.NameTaken(name, nil)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, apperr.ErrConflict.WithMessage("department name already exists")
	}
	d := &Department{Name: name}
	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Get(id uuid.UUID) (*Department, error) {
	return s.repo.GetByID(id)
}

func (s *Service) List(f ListFilter) ([]Department, int64, error) {
	return s.repo.List(f)
}

func (s *Service) Patch(id uuid.UUID, name string) (*Department, error) {
	d, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	taken, err := s.repo.NameTaken(name, &id)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, apperr.ErrConflict.WithMessage("department name already exists")
	}
	d.Name = name
	if err := s.repo.Save(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Delete(id uuid.UUID) error {
	return s.repo.SoftDelete(id)
}
