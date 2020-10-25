package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Notification struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	PostNotification `bson:",inline"`
	SentStatus       bool      `json:"sent_status" bson:"sent_status"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
}

type PostNotification struct {
	Sender  string   `json:"sender,omitempty" bson:"sender"`
	To      []string `json:"to" bson:"to"`
	Subject string   `json:"subject,omitempty" bson:"subject"`
	Message string   `json:"message" bson:"message"`
}
