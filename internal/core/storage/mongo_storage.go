package storage

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage implements RecipeRepository using MongoDB
type MongoStorage struct {
	client      *mongo.Client
	db          *mongo.Database
	collection  *mongo.Collection
	initialized bool
}

type StorageConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// NewMongoStorage creates a new MongoDB storage implementation
func NewMongoStorage(ctx context.Context, config StorageConfig) (RecipeRepository, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
	collection := db.Collection("recipes")

	return &MongoStorage{
		client:     client,
		db:         db,
		collection: collection,
	}, nil
}

func (s *MongoStorage) Close(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}

func (s *MongoStorage) CreateRecipe(ctx context.Context, recipe *models.Recipe) (*models.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.collection.InsertOne(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to save recipe: %v", ErrDatabaseError, err)
	}

	recipe.ID = result.InsertedID.(primitive.ObjectID)

	return recipe, nil
}

func (s *MongoStorage) GetRecipeByID(ctx context.Context, id string) (*models.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidID, err)
	}

	var recipe models.Recipe
	err = s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&recipe)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: recipe with ID %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return &recipe, nil
}

func (s *MongoStorage) GetRecipes(ctx context.Context, filter models.RecipeFilter, page int, limit int) (*models.RecipePage, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Add input validation
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Maximum limit to prevent excessive data fetching
	}

	// Convert RecipeFilter to bson.M
	bsonFilter := bson.M{}
	if filter.Title != "" {
		bsonFilter["title"] = bson.M{"$regex": primitive.Regex{Pattern: filter.Title, Options: "i"}}
	}
	if len(filter.IngredientNames) > 0 {
		var ingredientQueries []bson.M
		for _, name := range filter.IngredientNames {
			ingredientQuery := bson.M{"ingredients.name": bson.M{"$regex": primitive.Regex{Pattern: name, Options: "i"}}}
			ingredientQueries = append(ingredientQueries, ingredientQuery)
		}
		bsonFilter["$and"] = ingredientQueries
	}
	if filter.CookTime > 0 {
		bsonFilter["cook_time"] = bson.M{"$lte": filter.CookTime}
	}
	if len(filter.Tags) > 0 {
		bsonFilter["tags"] = bson.M{"$all": filter.Tags}
	}

	// Count total documents
	total, err := s.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to count documents: %v", ErrDatabaseError, err)
	}

	options := options.Find()
	options.SetSkip(int64((page - 1) * limit))
	options.SetLimit(int64(limit))
	options.SetSort(bson.M{"created_at": -1})

	cursor, err := s.collection.Find(ctx, bsonFilter, options)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch recipes: %v", ErrDatabaseError, err)
	}
	defer cursor.Close(ctx)

	var recipes []models.Recipe
	if err = cursor.All(ctx, &recipes); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	return &models.RecipePage{
		Recipes:    recipes,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *MongoStorage) UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe ID format: %w", err)
	}
	recipe.ID = objID

	result, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": recipe.ID},
		bson.M{
			"$set": bson.M{
				"title":       recipe.Title,
				"description": recipe.Description,
				"ingredients": recipe.Ingredients,
				"steps":       recipe.Steps,
				"cook_time":   recipe.CookTime,
				"servings":    recipe.Servings,
				"tags":        recipe.Tags,
				"updated_at":  recipe.UpdatedAt,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update recipe: %v", ErrDatabaseError, err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("%w: recipe with ID %s", ErrNotFound, recipe.ID.Hex())
	}

	return recipe, nil
}

func (s *MongoStorage) DeleteRecipe(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid recipe ID format: %w", err)
	}

	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("recipe not found with ID: %s", id)
	}

	return nil
}

func (s *MongoStorage) Initialize(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if s.initialized {
		return nil
	}

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "description", Value: "text"},
			},
			Options: options.Index().SetName("text_search"),
		},
		{
			Keys:    bson.D{{Key: "tags", Value: 1}},
			Options: options.Index().SetName("tags"),
		},
		{
			Keys:    bson.D{{Key: "ingredients.name", Value: 1}},
			Options: options.Index().SetName("ingredients_name"),
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("%w: failed to create indexes: %v", ErrDatabaseError, err)
	}

	s.initialized = true
	return nil
}
