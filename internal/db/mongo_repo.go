package db

import (
	"context"
	"defi-intel/internal/types"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type MongoRepository struct {
	client            *mongo.Client
	db                *mongo.Database
	pairsCollection   *mongo.Collection
	tokensCollection  *mongo.Collection
	metricsCollection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client) *MongoRepository {
	dbName := os.Getenv("MONGO_DB_NAME")

	db := client.Database(dbName)
	return &MongoRepository{
		client:            client,
		db:                db,
		pairsCollection:   db.Collection("pairs"),
		tokensCollection:  db.Collection("tokens"),
		metricsCollection: db.Collection("metrics"),
	}
}

func (repo *MongoRepository) StorePair(ctx context.Context, pair types.Pair) error {
	filter := bson.M{"address": pair.Address}
	update := bson.M{"$set": pair}

	opts := options.Update().SetUpsert(true)

	_, err := repo.pairsCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("cannot insert pair: %w", err)
	}

	return nil
}

func (repo *MongoRepository) StoreToken(ctx context.Context, token types.Token) error {
	filter := bson.M{"address": token.Address}
	update := bson.M{"$set": token}

	opts := options.Update().SetUpsert(true)

	_, err := repo.tokensCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("cannot insert token: %w", err)
	}

	return nil
}
