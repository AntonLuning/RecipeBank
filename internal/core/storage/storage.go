package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client *mongo.Client
	db     *mongo.Database
}

type StorageConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func NewStorage(config StorageConfig) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Construct the connection string
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(config.Database)

	// Create index on the 'id' field
	_, err = db.Collection("recipes").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &Storage{
		client: client,
		db:     db,
	}, nil
}

func (s *Storage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}

func (s *Storage) SaveRecipe(recipe models.Recipe) error {
	if recipe.Title == "" {
		return fmt.Errorf("title can not be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Collection("recipes").InsertOne(ctx, recipe)
	if err != nil {
		return fmt.Errorf("failed to save recipe: %w", err)
	}

	return nil
}

func (s *Storage) FetchRecipe(id string) (*models.Recipe, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var recipe models.Recipe
	err := s.db.Collection("recipes").FindOne(ctx, bson.M{"id": id}).Decode(&recipe)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("error fetching recipe: %w", err)
	}

	return &recipe, nil
}

func (s *Storage) FetchRecipes(filter models.RecipeFilter) ([]*models.Recipe, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Convert RecipeFilter to bson.M
	bsonFilter := bson.M{}
	if filter.Title != "" {
		bsonFilter["title"] = filter.Title
	}

	cursor, err := s.db.Collection("recipes").Find(ctx, bsonFilter)
	if err != nil {
		return nil, fmt.Errorf("error fetching recipes: %w", err)
	}
	defer cursor.Close(ctx)

	var recipes []*models.Recipe
	if err := cursor.All(ctx, &recipes); err != nil {
		return nil, fmt.Errorf("error decoding recipes: %w", err)
	}

	return recipes, nil
}
