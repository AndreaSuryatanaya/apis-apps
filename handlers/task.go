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

type TaskHandler struct {
	db           *gorm.DB
	auditService *services.AuditService
}

func NewTaskHandler(cfg *config.Config, auditService *services.AuditService) *TaskHandler {
	return &TaskHandler{
		db:           cfg.Database,
		auditService: auditService,
	}
}

// GET /tasks - Get all tasks
func (h *TaskHandler) GetTasks(c *fiber.Ctx) error {
	var tasks []models.Task

	if err := h.db.Preload("User").Find(&tasks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tasks",
		})
	}

	return c.JSON(fiber.Map{
		"data": tasks,
	})
}

// POST /tasks - Create a new task
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var task models.Task

	if err := c.BodyParser(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create task
	if err := h.db.Create(&task).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create task",
		})
	}

	// Load user relationship
	h.db.Preload("User").First(&task, task.ID)

	// Log audit
	userID := c.Locals("user_id")
	if userID != nil {
		taskJSON, _ := json.Marshal(task)
		var taskData map[string]interface{}
		json.Unmarshal(taskJSON, &taskData)

		h.auditService.LogCreate(userID.(string), "tasks", task.ID.String(), taskData)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": task,
	})
}

// PUT /tasks/:id - Update a task
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	taskID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(taskID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Get existing task
	var existingTask models.Task
	if err := h.db.Preload("User").First(&existingTask, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch task",
		})
	}

	// Store before state for audit
	beforeJSON, _ := json.Marshal(existingTask)
	var beforeData map[string]interface{}
	json.Unmarshal(beforeJSON, &beforeData)

	// Parse update data
	var updateData models.Task
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update task
	if err := h.db.Model(&existingTask).Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update task",
		})
	}

	// Get updated task
	h.db.Preload("User").First(&existingTask, "id = ?", id)

	// Store after state for audit
	afterJSON, _ := json.Marshal(existingTask)
	var afterData map[string]interface{}
	json.Unmarshal(afterJSON, &afterData)

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogUpdate(authUserID.(string), "tasks", id.String(), beforeData, afterData)
	}

	return c.JSON(fiber.Map{
		"data": existingTask,
	})
}

// DELETE /tasks/:id - Delete a task
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	taskID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(taskID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Get existing task for audit
	var task models.Task
	if err := h.db.Preload("User").First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch task",
		})
	}

	// Store data for audit
	taskJSON, _ := json.Marshal(task)
	var taskData map[string]interface{}
	json.Unmarshal(taskJSON, &taskData)

	// Delete task
	if err := h.db.Delete(&task, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete task",
		})
	}

	// Log audit
	authUserID := c.Locals("user_id")
	if authUserID != nil {
		h.auditService.LogDelete(authUserID.(string), "tasks", id.String(), taskData)
	}

	return c.JSON(fiber.Map{
		"message": "Task deleted successfully",
	})
}
