package handlers

import (
	"encoding/json"
	"todo-apps/config"
	"todo-apps/models"
	"todo-apps/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserPositionHandler struct {
	db           *gorm.DB
	auditService *services.AuditService
}

func NewUserPositionHandler(cfg *config.Config, auditService *services.AuditService) *UserPositionHandler {
	return &UserPositionHandler{
		db:           cfg.Database,
		auditService: auditService,
	}
}

// GET /user-positions - Get all user positions
func (h *UserPositionHandler) GetUserPositions(c *fiber.Ctx) error {
	var userPositions []models.UserPosition

	if err := h.db.Preload("User").Preload("Position").Find(&userPositions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user positions",
		})
	}

	return c.JSON(fiber.Map{
		"data": userPositions,
	})
}

// POST /user-positions - Create a new user position assignment
func (h *UserPositionHandler) CreateUserPosition(c *fiber.Ctx) error {
	var userPosition models.UserPosition

	if err := c.BodyParser(&userPosition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate that user and position exist
	var user models.User
	if err := h.db.First(&user, "id = ?", userPosition.UserID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var position models.Position
	if err := h.db.First(&position, "id = ?", userPosition.PositionID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Position not found",
		})
	}

	// Check if assignment already exists
	var existingAssignment models.UserPosition
	if err := h.db.Where("user_id = ? AND position_id = ?", userPosition.UserID, userPosition.PositionID).First(&existingAssignment).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User is already assigned to this position",
		})
	}

	// Create user position
	if err := h.db.Create(&userPosition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user position",
		})
	}

	// Load relationships
	h.db.Preload("User").Preload("Position").First(&userPosition, userPosition.ID)

	// Log audit
	userIDAuth := c.Locals("user_id")
	if userIDAuth != nil {
		userPositionJSON, _ := json.Marshal(userPosition)
		var userPositionData map[string]interface{}
		json.Unmarshal(userPositionJSON, &userPositionData)

		h.auditService.LogCreate(userIDAuth.(string), "user_positions", userPosition.ID.String(), userPositionData)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": userPosition,
	})
}

// DELETE /user-positions/:id - Delete a user position assignment
func (h *UserPositionHandler) DeleteUserPosition(c *fiber.Ctx) error {
	userPositionID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(userPositionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user position ID",
		})
	}

	// Get existing user position for audit
	var userPosition models.UserPosition
	if err := h.db.Preload("User").Preload("Position").First(&userPosition, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User position not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user position",
		})
	}

	// Store data for audit
	userPositionJSON, _ := json.Marshal(userPosition)
	var userPositionData map[string]interface{}
	json.Unmarshal(userPositionJSON, &userPositionData)

	// Delete user position
	if err := h.db.Delete(&userPosition, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user position",
		})
	}

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogDelete(authUserID.(string), "user_positions", id.String(), userPositionData)
	}

	return c.JSON(fiber.Map{
		"message": "User position deleted successfully",
	})
}
