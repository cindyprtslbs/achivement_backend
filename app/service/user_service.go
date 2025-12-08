package service

import (
	"golang.org/x/crypto/bcrypt"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// =======================================================
// GET ALL USERS
// =======================================================
func (s *UserService) GetAll(c *fiber.Ctx) error {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch users"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// =======================================================
// GET USER BY ID
// =======================================================
func (s *UserService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// =======================================================
// CREATE USER
// =======================================================
func (s *UserService) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validasi email
	if u, _ := s.userRepo.GetByEmail(req.Email); u != nil {
		return c.Status(400).JSON(fiber.Map{"error": "email already registered"})
	}

	// Validasi username
	if u, _ := s.userRepo.GetByUsername(req.Username); u != nil {
		return c.Status(400).JSON(fiber.Map{"error": "username already taken"})
	}

	// Hash password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	req.PasswordHash = string(hashed)

	user, err := s.userRepo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create user", "detail": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "user created",
		"data":    user,
	})
}

// =======================================================
// UPDATE USER
// =======================================================
func (s *UserService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validasi email
	if req.Email != existing.Email {
		if u, _ := s.userRepo.GetByEmail(req.Email); u != nil {
			return c.Status(400).JSON(fiber.Map{"error": "email already registered"})
		}
	}

	// Validasi username
	if req.Username != existing.Username {
		if u, _ := s.userRepo.GetByUsername(req.Username); u != nil {
			return c.Status(400).JSON(fiber.Map{"error": "username already taken"})
		}
	}

	// Validasi role
	if _, err := s.roleRepo.GetByID(req.RoleID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid role_id"})
	}

	user, err := s.userRepo.Update(id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update user"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "user updated",
		"data":    user,
	})
}

func (s *UserService) UpdatePassword(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "password is required"})
	}

	// cek user ada
	if _, err := s.userRepo.GetByID(id); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	// hash
	hashed, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)

	// update password
	if err := s.userRepo.UpdatePassword(id, string(hashed)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update password"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "password updated",
	})
}

// =======================================================
// DELETE USER
// =======================================================
func (s *UserService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := s.userRepo.GetByID(id); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if err := s.userRepo.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete user"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "user deleted",
	})
}
