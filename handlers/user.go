package handlers

import (
	"encoding/json"
	"todo-apps/config"
	"todo-apps/models"
	"todo-apps/services"
	"todo-apps/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	db           *gorm.DB
	auditService *services.AuditService
}

func NewUserHandler(cfg *config.Config, auditService *services.AuditService) *UserHandler {
	return &UserHandler{
		db:           cfg.Database,
		auditService: auditService,
	}
}

// GET /users - Get all users
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	var users []models.User

	if err := h.db.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"data": users,
	})
}

// POST /users - Create a new user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}
	user.Password = hashedPassword

	// Create user
	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Log audit
	userID := c.Locals("user_id")
	if userID != nil {
		userJSON, _ := json.Marshal(user)
		var userData map[string]interface{}
		json.Unmarshal(userJSON, &userData)
		delete(userData, "password") // Remove password from audit log

		h.auditService.LogCreate(userID.(string), "users", user.ID.String(), userData)
	}

	// Remove password from response
	user.Password = ""

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": user,
	})
}

// PUT /users/:id - Update a user
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Get existing user
	var existingUser models.User
	if err := h.db.First(&existingUser, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Store before state for audit
	beforeJSON, _ := json.Marshal(existingUser)
	var beforeData map[string]interface{}
	json.Unmarshal(beforeJSON, &beforeData)
	delete(beforeData, "password")

	// Parse update data
	var updateData models.User
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Hash password if provided
	if updateData.Password != "" {
		hashedPassword, err := utils.HashPassword(updateData.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash password",
			})
		}
		updateData.Password = hashedPassword
	}

	// Update user
	if err := h.db.Model(&existingUser).Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	// Get updated user
	h.db.First(&existingUser, "id = ?", id)

	// Store after state for audit
	afterJSON, _ := json.Marshal(existingUser)
	var afterData map[string]interface{}
	json.Unmarshal(afterJSON, &afterData)
	delete(afterData, "password")

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogUpdate(authUserID.(string), "users", id.String(), beforeData, afterData)
	}

	// Remove password from response
	existingUser.Password = ""

	return c.JSON(fiber.Map{
		"data": existingUser,
	})
}

// DELETE /users/:id - Delete a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Get existing user for audit
	var user models.User
	if err := h.db.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Store data for audit
	userJSON, _ := json.Marshal(user)
	var userData map[string]interface{}
	json.Unmarshal(userJSON, &userData)
	delete(userData, "password")

	// Delete user
	if err := h.db.Delete(&user, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogDelete(authUserID.(string), "users", id.String(), userData)
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
