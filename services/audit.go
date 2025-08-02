package services

import (
	"context"
	"time"

	"todo-apps/config"
	"todo-apps/models"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuditService struct {
	collection *mongo.Collection
}

func NewAuditService(mongodb *config.MongoDB) *AuditService {
	return &AuditService{
		collection: mongodb.Database.Collection("audit_logs"),
	}
}

func (s *AuditService) LogAction(userID, action, entity, entityID string, before, after interface{}) error {
	auditLog := models.AuditLog{
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Timestamp: time.Now(),
		Meta: models.AuditMeta{
			Before: before,
			After:  after,
		},
	}

	_, err := s.collection.InsertOne(context.TODO(), auditLog)
	return err
}

func (s *AuditService) LogCreate(userID, entity, entityID string, data interface{}) error {
	return s.LogAction(userID, "CREATE", entity, entityID, nil, data)
}

func (s *AuditService) LogUpdate(userID, entity, entityID string, before, after interface{}) error {
	return s.LogAction(userID, "UPDATE", entity, entityID, before, after)
}

func (s *AuditService) LogDelete(userID, entity, entityID string, data interface{}) error {
	return s.LogAction(userID, "DELETE", entity, entityID, data, nil)
}
