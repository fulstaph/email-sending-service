package repository

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"projects/email-sending-service/config"
	"projects/email-sending-service/internal/models"
)

type EmailRepository interface {
	Connect() error
	Get(limit, skip int) ([]models.Notification, int64, error)
	GetByID(id primitive.ObjectID) (models.Notification, error)
	Save(email models.Notification) (primitive.ObjectID, error)
	Close() error
}

func New(cfg config.Database) EmailRepository {
	return &emailRepo{cfg: cfg}
}

type emailRepo struct {
	client *mongo.Client
	cfg    config.Database
}

func (e *emailRepo) getColl() *mongo.Collection {
	if e.client == nil {
		err := e.Connect()
		if err != nil {
			log.Error(err)
		}
	} else {
		if err := e.client.Ping(context.TODO(), nil); err != nil {
			connected := false
			reconnectCount := 0
			// try to reconnect every 5 sec
			for !connected {
				err := e.Connect()
				if err == nil {
					connected = true
				}
				time.Sleep(time.Second * 5)
				reconnectCount++
			}

			return e.client.Database(e.cfg.Name).Collection(e.cfg.Collection)
		}
	}
	return e.client.Database(e.cfg.Name).Collection(e.cfg.Collection)
}

func (e *emailRepo) Get(limit, skip int) ([]models.Notification, int64, error) {
	collection := e.getColl()

	var totalCount int64
	totalCount, err := collection.CountDocuments(context.TODO(), bson.M{}, nil)
	if err != nil {
		return nil, 0, err
	}

	var result []models.Notification

	if limit == 0 {
		return result, totalCount, nil
	}

	var finalSkip int64 = 0

	if skip > 0 {
		finalSkip = int64((skip - 1) * limit)
	}

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(finalSkip)
	findOptions.SetSort(bson.D{{"_id", -1}}) //nolint:govet

	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(context.TODO())

	var notif models.Notification
	for cur.Next(context.TODO()) {
		if err := cur.Decode(&notif); err != nil {
			return nil, 0, err
		}
		result = append(result, notif)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return result, totalCount, nil
}

func (e *emailRepo) GetByID(id primitive.ObjectID) (models.Notification, error) {
	collection := e.getColl()
	var result models.Notification

	filter := bson.M{"_id": id}

	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (e *emailRepo) Save(email models.Notification) (primitive.ObjectID, error) {
	collection := e.getColl()
	result, err := collection.InsertOne(context.Background(), email)
	if err != nil {
		return [12]byte{}, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (e *emailRepo) Connect() error {
	url := strings.Builder{}
	url.WriteString(e.cfg.Host)
	url.WriteString(e.cfg.Port)
	opts := options.Client().ApplyURI(url.String())

	var err error
	e.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		return err
	}

	if err = e.client.Ping(context.TODO(), nil); err != nil {
		return err
	}

	return nil
}

func (e *emailRepo) Close() error {
	if err := e.client.Disconnect(context.TODO()); err != nil {
		return err
	}
	return nil
}
