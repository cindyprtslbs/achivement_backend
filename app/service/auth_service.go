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
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

type refreshEntry struct {
	ExpiresAt time.Time
	UserID    string
}

var (
	// refreshTokens maps refresh token -> metadata (expires, user)
	refreshTokens   = make(map[string]refreshEntry)
	refreshMutex    sync.RWMutex
	refreshTokenTTL = time.Hour * 24 * 7 // 7 hari
)

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	rolePermRepo repository.RolePermissionRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// Login godoc
// @Summary Login pengguna
// @Description Autentikasi pengguna menggunakan username dan password, mengembalikan access token dan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Username dan Password"
// @Success 200 {object} models.LoginResponse "Login berhasil"
// @Failure 400 {object} map[string]interface{} "Request tidak valid"
// @Failure 401 {object} map[string]interface{} "Username atau password salah"
// @Failure 403 {object} map[string]interface{} "User tidak aktif"
// @Failure 500 {object} map[string]interface{} "Kesalahan server"
// @Router /api/v1/auth/login [post]
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
	role, err := s.roleRepo.GetByID(*user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "role not found"})
	}

	// ===============================================================
	// GET PERMISSIONS
	// ===============================================================
	perms, _ := s.rolePermRepo.GetPermissionsByRole(*user.RoleID)

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

	refreshMutex.Lock()
	refreshTokens[refreshToken] = refreshEntry{ExpiresAt: time.Now().Add(refreshTokenTTL), UserID: user.ID}
	refreshMutex.Unlock()

	// ===============================================================
	// SUCCESS RESPONSE (SESUIAI SRS)
	// ===============================================================
	resp := models.LoginResponse{
		Status: "success",
		Data: models.LoginData{
			Token:        accessToken,
			RefreshToken: refreshToken,
			User: models.LoginUser{
				ID:          user.ID,
				Username:    user.Username,
				FullName:    user.FullName,
				Role:        role.Name,
				Permissions: permList,
			},
		},
	}

	return c.JSON(resp)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Menghasilkan access token baru menggunakan refresh token yang masih valid
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object true "Refresh token" example({"refresh_token":"string"})
// @Success 200 {object} map[string]interface{} "Token berhasil diperbarui"
// @Failure 400 {object} map[string]interface{} "Refresh token wajib diisi"
// @Failure 401 {object} map[string]interface{} "Refresh token tidak valid atau kadaluarsa"
// @Failure 500 {object} map[string]interface{} "Kesalahan server"
// @Router /api/v1/auth/refresh [post]
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		Refresh string `json:"refresh_token"`
	}

	if err := c.BodyParser(&body); err != nil || body.Refresh == "" {
		return c.Status(400).JSON(fiber.Map{"error": "refresh_token required"})
	}

	// cek refresh token valid atau expired
	refreshMutex.RLock()
	entry, exists := refreshTokens[body.Refresh]
	refreshMutex.RUnlock()

	if !exists || time.Now().After(entry.ExpiresAt) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid or expired refresh token"})
	}

	// get user id from stored refresh token entry
	user, err := s.userRepo.GetByID(entry.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "user not found"})
	}

	// permissions
	perms, _ := s.rolePermRepo.GetPermissionsByRole(*user.RoleID)

	var permList []string
	for _, p := range perms {
		permList = append(permList, p.Name)
	}

	// ambil role
	role, _ := s.roleRepo.GetByID(*user.RoleID)

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

// Logout godoc
// @Summary Logout pengguna
// @Description Logout pengguna dan mem-blacklist access token serta menghapus refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Logout berhasil"
// @Failure 400 {object} map[string]interface{} "Token tidak ditemukan"
// @Security Bearer
// @Router /api/v1/auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	raw := c.Locals("raw_token")
	if raw == nil {
		return c.Status(400).JSON(fiber.Map{"error": "token not found in context"})
	}

	token := raw.(string)
	exp := time.Now().Add(time.Hour * 1)

	// add token to global utils blacklist so middleware recognizes it
	utils.BlacklistMutex.Lock()
	utils.TokenBlacklist[token] = exp
	utils.BlacklistMutex.Unlock()

	// Revoke refresh tokens belonging to this user (if available in context)
	uid := c.Locals("user_id")
	if uid != nil {
		userID := uid.(string)
		refreshMutex.Lock()
		for t, e := range refreshTokens {
			if e.UserID == userID {
				delete(refreshTokens, t)
			}
		}
		refreshMutex.Unlock()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "logout success",
	})
}

// GetProfile godoc
// @Summary Mendapatkan profil pengguna
// @Description Mengambil data profil pengguna berdasarkan token (Mahasiswa, Dosen Wali, atau Admin)
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Profil pengguna"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User tidak ditemukan"
// @Security Bearer
// @Router /api/v1/auth/profile [get]
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	role := c.Locals("role_name").(string)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	data := fiber.Map{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.RoleName,
		"is_active": user.IsActive,
	}

	switch role {
	case "Mahasiswa":
		student, _ := s.studentRepo.GetByUserID(userID)
		data["profile"] = student

	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		data["profile"] = lecturer

	case "Admin":
		data["profile"] = nil
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}
