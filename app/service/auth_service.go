package service

import (
	"database/sql"
	// "errors"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"
	"achievement_backend/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     repository.UserRepository
	rolePermRepo repository.RolePermissionRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	rolePermRepo repository.RolePermissionRepository,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username and password required"})
	}

	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{"error": "wrong username or password"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "database error"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "user inactive"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "wrong username or password"})
	}

	perms, _ := s.rolePermRepo.GetPermissionsByRole(user.RoleID)

	var permList []string
	for _, p := range perms {
		permList = append(permList, p.Name)
	}

	token, err := utils.GenerateToken(*user, permList)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token generation failed"})
	}

	resp := models.LoginResponse{
		Token:       token,
		UserID:      user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		RoleID:      user.RoleID,
		Permissions: permList,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "login success",
		"data":    resp,
	})
}

func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	username := c.Locals("username")
	roleID := c.Locals("role_id")
	perms := c.Locals("permissions")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile berhasil diambil",
		"data": fiber.Map{
			"user_id":     userID,
			"username":    username,
			"role_id":     roleID,
			"permissions": perms,
		},
	})
}