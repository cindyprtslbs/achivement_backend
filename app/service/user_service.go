package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"
)

type UserService interface {
	GetAll() ([]models.User, error)
	GetByID(id string) (*models.User, error)
	Create(req models.CreateUserRequest) (*models.User, error)
	Update(id string, req models.UpdateUserRequest) (*models.User, error)
	Delete(id string) error
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *userService) GetAll() ([]models.User, error) {
	return s.userRepo.GetAll()
}

func (s *userService) GetByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) Create(req models.CreateUserRequest) (*models.User, error) {
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	existingUsername, _ := s.userRepo.GetByUsername(req.Username)
	if existingUsername != nil {
		return nil, errors.New("username already taken")
	}

	_, err := s.roleRepo.GetByID(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	req.PasswordHash = string(hashed)

	return s.userRepo.Create(req)
}

func (s *userService) Update(id string, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Email != user.Email {
		existingEmail, _ := s.userRepo.GetByEmail(req.Email)
		if existingEmail != nil {
			return nil, errors.New("email already registered")
		}
	}

	if req.Username != user.Username {
		existingUsername, _ := s.userRepo.GetByUsername(req.Username)
		if existingUsername != nil {
			return nil, errors.New("username already taken")
		}
	}

	_, err = s.roleRepo.GetByID(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id")
	}

	return s.userRepo.Update(id, req)
}

func (s *userService) Delete(id string) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(id)
}
