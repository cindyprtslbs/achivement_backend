package service

import (
	models "achievement_backend/app/model"
	"achievement_backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

type RoleService struct {
	repo repository.RoleRepository
}

func NewRoleService(r repository.RoleRepository) *RoleService {
	return &RoleService{repo: r}
}

// =======================================================
// GET ALL ROLES
// =======================================================
func (s *RoleService) GetAll(c *fiber.Ctx) error {
	roles, err := s.repo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch roles"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    roles,
	})
}

// =======================================================
// GET ROLE BY ID
// =======================================================
func (s *RoleService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	role, err := s.repo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "role not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    role,
	})
}

// =======================================================
// CREATE ROLE
// =======================================================
func (s *RoleService) Create(c *fiber.Ctx) error {
	var req models.CreateRoleRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	role, err := s.repo.Create(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create role"})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "role created",
		"data":    role,
	})
}

// =======================================================
// UPDATE ROLE
// =======================================================
func (s *RoleService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	role, err := s.repo.Update(id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update role"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "role updated",
		"data":    role,
	})
}

// =======================================================
// DELETE ROLE
// =======================================================
func (s *RoleService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := s.repo.Delete(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete role"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "role deleted",
	})
}
