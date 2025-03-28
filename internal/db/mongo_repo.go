package db

import (
	"blocsy/internal/solana"
	"blocsy/internal/types"
	"context"
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

func (repo *MongoRepository) LookupByToken(ctx context.Context, address string, network string) (types.Token, []types.Pair, error) {
	token, err := repo.lookupToken(ctx, address, network)
	if err != nil {
		return types.Token{}, nil, fmt.Errorf("cannot find token: %w", err)
	}

	if token.Address == "" {
		return types.Token{}, []types.Pair{}, nil
	}

	filter := bson.M{
		"token":   token.Address,
		"network": token.Network,
	}

	cur, err := repo.pairsCollection.Find(ctx, filter, nil)
	if err != nil {
		return token, nil, nil
	}

	var pairs []types.Pair
	if err := cur.All(ctx, &pairs); err != nil {
		return types.Token{}, nil, fmt.Errorf("cannot decode pairs: %w", err)
	}

	return token, pairs, nil
}

func (repo *MongoRepository) lookupToken(ctx context.Context, address string, network string) (types.Token, error) {
	filer := bson.M{"address": address, "network": network}

	res := repo.tokensCollection.FindOne(ctx, filer)
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return types.Token{}, nil
		}

		return types.Token{}, fmt.Errorf("cannot find token: %w", err)
	}

	var token types.Token
	if err := res.Decode(&token); err != nil {
		return types.Token{}, fmt.Errorf("cannot decode token: %w", err)
	}

	return token, nil
}

func (repo *MongoRepository) LookupByPair(ctx context.Context, address string, network string, tf solana.TokenFinder) (types.Pair, types.Token, types.QuoteToken, error) {
	filter := bson.M{
		"address": address,
		"network": network,
	}

	cur, err := repo.pairsCollection.Find(ctx, filter, nil)
	if err != nil {
		return types.Pair{}, types.Token{}, types.QuoteToken{}, fmt.Errorf("cannot find pair: %w", err)
	}

	var pairs []types.Pair
	if err := cur.All(ctx, &pairs); err != nil {
		return types.Pair{}, types.Token{}, types.QuoteToken{}, fmt.Errorf("cannot decode pairs: %w", err)
	}

	if len(pairs) == 0 {
		return types.Pair{}, types.Token{}, types.QuoteToken{}, nil
	}

	token, _, err := tf.FindToken(ctx, pairs[0].Token, false)
	if err != nil {
		return types.Pair{}, types.Token{}, types.QuoteToken{}, fmt.Errorf("cannot find token: %w", err)
	}

	if token.Address == "" {
		//When no pair found return empty pair and token and caller can trigger a function to find pair and store it.
		return types.Pair{}, types.Token{}, types.QuoteToken{}, nil
	}

	quoteToken, _, err := tf.FindToken(ctx, pairs[0].QuoteToken.Address, false)
	if err != nil {
		return types.Pair{}, types.Token{}, types.QuoteToken{}, fmt.Errorf("cannot find quote token: %w", err)
	}

	if quoteToken.Address == "" {
		return types.Pair{}, types.Token{}, types.QuoteToken{}, nil
	}

	return pairs[0], *token, types.QuoteToken{
		Identifier: pairs[0].QuoteToken.Identifier,
		Name:       quoteToken.Name,
		Symbol:     quoteToken.Symbol,
		Address:    quoteToken.Address,
		Decimals:   quoteToken.Decimals,
	}, nil
}

func (repo *MongoRepository) UpdateTokenSupply(ctx context.Context, address string, supply string) error {
	filter := bson.M{"address": address}
	update := bson.M{"$set": bson.M{"supply": supply}}

	opts := options.Update().SetUpsert(true)

	_, err := repo.tokensCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("cannot update token supply: %w", err)
	}

	return nil
}

func (repo *MongoRepository) PullPairs(ctx context.Context) (<-chan types.Pair, <-chan error) {
	pairCh := make(chan types.Pair)
	errCh := make(chan error, 1)

	opts := options.Find().SetBatchSize(1000)

	go func() {
		defer close(pairCh)
		defer close(errCh)

		cursor, err := repo.pairsCollection.Find(ctx, bson.M{}, opts)
		if err != nil {
			errCh <- fmt.Errorf("failed to execute find: %w", err)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var pair types.Pair
			if err = cursor.Decode(&pair); err != nil {
				errCh <- fmt.Errorf("failed to decode token: %w", err)
				return
			}
			pairCh <- pair
		}
		if err = cursor.Err(); err != nil {
			errCh <- fmt.Errorf("cursor encountered an error: %w", err)
			return
		}
	}()

	return pairCh, errCh
}
