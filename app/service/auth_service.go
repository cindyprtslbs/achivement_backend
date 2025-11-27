package service

import (
	"database/sql"
	"sync"
	"time"

	models "achievement_backend/app/model"
	"achievement_backend/app/repository"
	"achievement_backend/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	rolePermRepo repository.RolePermissionRepository
}

var (
	tokenBlacklist  = make(map[string]time.Time)
	blacklistMutex  sync.RWMutex
	refreshTokenTTL = time.Hour * 24 * 7 // 7 hari
)

// ===============================================================
// CONSTRUCTOR
// ===============================================================

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	rolePermRepo repository.RolePermissionRepository,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

// ===============================================================
// LOGIN (FR-001)
// ===============================================================

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	// Ambil user
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{"error": "wrong username or password"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "database error"})
	}

	// User nonaktif → tidak boleh login
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "user inactive"})
	}

	// Cek password
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return c.Status(401).JSON(fiber.Map{"error": "wrong username or password"})
	}

	// ===============================================================
	// GET ROLE NAME
	// ===============================================================
	role, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "role not found"})
	}

	// ===============================================================
	// GET PERMISSIONS
	// ===============================================================
	perms, _ := s.rolePermRepo.GetPermissionsByRole(user.RoleID)

	var permList []string
	for _, p := range perms {
		permList = append(permList, p.Name)
	}

	// ===============================================================
	// GENERATE ACCESS TOKEN
	// ===============================================================
	accessToken, err := utils.GenerateToken(*user, role.Name, permList)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	// ===============================================================
	// GENERATE REFRESH TOKEN
	// ===============================================================
	refreshToken := utils.GenerateRefreshToken()

	blacklistMutex.Lock()
	tokenBlacklist[refreshToken] = time.Now().Add(refreshTokenTTL)
	blacklistMutex.Unlock()

	// ===============================================================
	// SUCCESS RESPONSE (SESUIAI SRS)
	// ===============================================================
	resp := models.LoginResponse{
		Token:       accessToken,
		Refresh:     refreshToken,
		UserID:      user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		RoleName:    role.Name,
		Permissions: permList,
	}

	return c.JSON(resp)
}

// ===============================================================
// REFRESH TOKEN
// ===============================================================

func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		Refresh string `json:"refresh_token"`
	}

	if err := c.BodyParser(&body); err != nil || body.Refresh == "" {
		return c.Status(400).JSON(fiber.Map{"error": "refresh_token required"})
	}

	// cek refresh token valid atau expired
	blacklistMutex.RLock()
	exp, exists := tokenBlacklist[body.Refresh]
	blacklistMutex.RUnlock()

	if !exists || time.Now().After(exp) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid or expired refresh token"})
	}

	// access token lama harus membawa user_id
	userID := c.Locals("user_id").(string)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "user not found"})
	}

	// permissions
	perms, _ := s.rolePermRepo.GetPermissionsByRole(user.RoleID)

	var permList []string
	for _, p := range perms {
		permList = append(permList, p.Name)
	}

	// ambil role
	role, _ := s.roleRepo.GetByID(user.RoleID)

	// token baru
	newToken, err := utils.GenerateToken(*user, role.Name, permList)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"token":   newToken,
	})
}

// ===============================================================
// LOGOUT
// ===============================================================

func (s *AuthService) Logout(c *fiber.Ctx) error {
	raw := c.Locals("raw_token")
	if raw == nil {
		return c.Status(400).JSON(fiber.Map{"error": "token not found in context"})
	}

	token := raw.(string)
	exp := time.Now().Add(time.Hour * 1)

	blacklistMutex.Lock()
	tokenBlacklist[token] = exp
	blacklistMutex.Unlock()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "logout success",
	})
}

// ===============================================================
// GET PROFILE (FR-002 Authorization)
// ===============================================================

func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":     c.Locals("user_id"),
			"username":    c.Locals("username"),
			"role_name":   c.Locals("role_name"),
			"permissions": c.Locals("permissions"),
		},
	})
}
