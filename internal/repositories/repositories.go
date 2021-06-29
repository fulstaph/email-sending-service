package repositories

import (
	"email-sender/internal/repositories/emails"

	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	Emails emails.Repository
}

func New(client *mongo.Database) *Container {
	return &Container{Emails: emails.New(client)}
}
