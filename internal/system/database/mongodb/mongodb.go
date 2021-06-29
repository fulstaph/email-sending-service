package mongodb

import (
	"context"

	"email-sender/config" //nolint:goimports

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client interface {
	GetConnection() *mongo.Database
	Close() error
}

type client struct {
	ctx  context.Context
	conn *mongo.Client
	cfg  *config.Database
}

func (c *client) GetConnection() *mongo.Database {
	// reconnect ?
	return c.conn.Database(c.cfg.Name)
}

func (c *client) Close() error {
	return c.conn.Disconnect(c.ctx)
}

func NewClient(cfg *config.Database) (Client, error) {
	opts := options.Client().ApplyURI(cfg.DSN)

	ctx := context.Background()

	clientConn, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := clientConn.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &client{
		ctx:  ctx,
		conn: clientConn,
		cfg:  cfg,
	}, nil
}
