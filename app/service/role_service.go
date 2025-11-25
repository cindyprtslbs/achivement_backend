package service

import (
	"errors"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"
)

type RoleService interface {
	GetAll() ([]models.Role, error)
	GetByID(id string) (*models.Role, error)
	Create(req models.CreateRoleRequest) (*models.Role, error)
	Update(id string, req models.UpdateRoleRequest) (*models.Role, error)
	Delete(id string) error
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{
		roleRepo: roleRepo,
	}
}

func (s *roleService) GetAll() ([]models.Role, error) {
	return s.roleRepo.GetAll()
}

func (s *roleService) GetByID(id string) (*models.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *roleService) Create(req models.CreateRoleRequest) (*models.Role, error) {
	roles, _ := s.roleRepo.GetAll()
	for _, r := range roles {
		if r.Name == req.Name {
			return nil, errors.New("role name already exists")
		}
	}

	return s.roleRepo.Create(req)
}

func (s *roleService) Update(id string, req models.UpdateRoleRequest) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	if req.Name != role.Name {
		roles, _ := s.roleRepo.GetAll()
		for _, r := range roles {
			if r.Name == req.Name {
				return nil, errors.New("role name already exists")
			}
		}
	}

	return s.roleRepo.Update(id, req)
}

func (s *roleService) Delete(id string) error {
	_, err := s.roleRepo.GetByID(id)
	if err != nil {
		return errors.New("role not found")
	}

	return s.roleRepo.Delete(id)
}
