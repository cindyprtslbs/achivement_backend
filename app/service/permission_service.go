package service

import (
	"github.com/gofiber/fiber/v2"
	models "achievement_backend/app/model"
	"achievement_backend/app/repository"
)

type PermissionService struct {
	repo repository.PermissionRepository
}

func NewPermissionService(r repository.PermissionRepository) *PermissionService {
	return &PermissionService{repo: r}
}

//
// ===================== GET ALL PERMISSIONS =====================
//
func (s *PermissionService) GetAll(c *fiber.Ctx) error {
	data, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch permissions",
		})
	}
	return c.JSON(data)
}

//
// ===================== GET PERMISSION BY ID =====================
//
func (s *PermissionService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	perm, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "permission not found",
		})
	}

	return c.JSON(perm)
}

//
// ===================== CREATE PERMISSION =====================
//
func (s *PermissionService) Create(c *fiber.Ctx) error {
	var req models.CreatePermissionRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" || req.Resource == "" || req.Action == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name, resource, and action are required",
		})
	}

	perm, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to create permission",
		})
	}

	return c.Status(201).JSON(perm)
}

//
// ===================== UPDATE PERMISSION =====================
//
func (s *PermissionService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdatePermissionRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" || req.Resource == "" || req.Action == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name, resource, and action are required",
		})
	}

	perm, err := s.repo.Update(id, req)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "permission not found or update failed",
		})
	}

	return c.JSON(perm)
}

//
// ===================== DELETE PERMISSION =====================
//
func (s *PermissionService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := s.repo.Delete(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "permission not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}
