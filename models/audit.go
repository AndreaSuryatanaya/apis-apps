package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Action    string             `bson:"action" json:"action"` // CREATE, UPDATE, DELETE
	Entity    string             `bson:"entity" json:"entity"` // users, tasks, positions, user_positions
	EntityID  string             `bson:"entity_id" json:"entity_id"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Meta      AuditMeta          `bson:"meta" json:"meta"`
}

type AuditMeta struct {
	Before interface{} `bson:"before,omitempty" json:"before,omitempty"`
	After  interface{} `bson:"after,omitempty" json:"after,omitempty"`
}
