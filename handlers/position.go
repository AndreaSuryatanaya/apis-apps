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

type PositionHandler struct {
	db           *gorm.DB
	auditService *services.AuditService
}

func NewPositionHandler(cfg *config.Config, auditService *services.AuditService) *PositionHandler {
	return &PositionHandler{
		db:           cfg.Database,
		auditService: auditService,
	}
}

// GET /positions - Get all positions
func (h *PositionHandler) GetPositions(c *fiber.Ctx) error {
	var positions []models.Position

	if err := h.db.Preload("UserPositions.User").Find(&positions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch positions",
		})
	}

	return c.JSON(fiber.Map{
		"data": positions,
	})
}

// POST /positions - Create a new position
func (h *PositionHandler) CreatePosition(c *fiber.Ctx) error {
	var position models.Position

	if err := c.BodyParser(&position); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create position
	if err := h.db.Create(&position).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create position",
		})
	}

	// Log audit
	userID := c.Locals("user_id")
	if userID != nil {
		positionJSON, _ := json.Marshal(position)
		var positionData map[string]interface{}
		json.Unmarshal(positionJSON, &positionData)

		h.auditService.LogCreate(userID.(string), "positions", position.ID.String(), positionData)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": position,
	})
}

// PUT /positions/:id - Update a position
func (h *PositionHandler) UpdatePosition(c *fiber.Ctx) error {
	positionID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(positionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid position ID",
		})
	}

	// Get existing position
	var existingPosition models.Position
	if err := h.db.First(&existingPosition, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Position not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch position",
		})
	}

	// Store before state for audit
	beforeJSON, _ := json.Marshal(existingPosition)
	var beforeData map[string]interface{}
	json.Unmarshal(beforeJSON, &beforeData)

	// Parse update data
	var updateData models.Position
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update position
	if err := h.db.Model(&existingPosition).Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update position",
		})
	}

	// Get updated position
	h.db.First(&existingPosition, "id = ?", id)

	// Store after state for audit
	afterJSON, _ := json.Marshal(existingPosition)
	var afterData map[string]interface{}
	json.Unmarshal(afterJSON, &afterData)

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogUpdate(authUserID.(string), "positions", id.String(), beforeData, afterData)
	}

	return c.JSON(fiber.Map{
		"data": existingPosition,
	})
}

// DELETE /positions/:id - Delete a position
func (h *PositionHandler) DeletePosition(c *fiber.Ctx) error {
	positionID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(positionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid position ID",
		})
	}

	// Get existing position for audit
	var position models.Position
	if err := h.db.First(&position, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Position not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch position",
		})
	}

	// Store data for audit
	positionJSON, _ := json.Marshal(position)
	var positionData map[string]interface{}
	json.Unmarshal(positionJSON, &positionData)

	// Delete position
	if err := h.db.Delete(&position, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete position",
		})
	}

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogDelete(authUserID.(string), "positions", id.String(), positionData)
	}

	return c.JSON(fiber.Map{
		"message": "Position deleted successfully",
	})
}
