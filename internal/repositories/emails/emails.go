package emails

import (
	"context"

	"email-sender/internal/entities" //nolint:goimports
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "notifications"

type Repository interface {
	List(ctx context.Context, limit, skip int64) ([]entities.Notification, int64, error)
	Get(ctx context.Context, id primitive.ObjectID) (*entities.Notification, error)
	Save(ctx context.Context, email *entities.Notification) (primitive.ObjectID, error)
}

func New(client *mongo.Database) Repository {
	return &repository{
		client: client,
	}
}

type repository struct {
	client *mongo.Database
}

func (e *repository) getCollection() *mongo.Collection {
	return e.client.Collection(collectionName)
}

func (e *repository) List(ctx context.Context, limit, skip int64) ([]entities.Notification, int64, error) {
	collection := e.getCollection()

	var totalCount int64
	totalCount, err := collection.CountDocuments(ctx, bson.M{}, nil)
	if err != nil {
		return nil, 0, err
	}

	var result []entities.Notification

	var finalSkip int64

	if skip > 0 {
		finalSkip = (skip - 1) * limit
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(finalSkip).
		SetSort(bson.D{{"_id", -1}}) //nolint:govet

	cur, err := collection.Find(ctx, bson.D{{}}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var notification entities.Notification
	for cur.Next(ctx) {
		if err = cur.Decode(&notification); err != nil {
			return nil, 0, err
		}
		result = append(result, notification)
	}

	if err = cur.Err(); err != nil {
		return nil, 0, err
	}

	return result, totalCount, nil
}

func (e *repository) Get(ctx context.Context, id primitive.ObjectID) (*entities.Notification, error) {
	collection := e.getCollection()
	filter := bson.M{"_id": id}

	var result entities.Notification
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (e *repository) Save(ctx context.Context, email *entities.Notification) (primitive.ObjectID, error) {
	collection := e.getCollection()

	result, err := collection.InsertOne(ctx, email)
	if err != nil {
		return [12]byte{}, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}
