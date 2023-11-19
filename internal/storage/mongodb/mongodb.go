package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	db *mongo.Database
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.mongodb.New"
	db, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(storagePath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		err = db.Disconnect(context.TODO())
		if err != nil {
			panic(err)
		}
	}()
	return &Storage{
		db: db.Database("testRest"),
	}, nil
}
