package storage

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/AntonLuning/RecipeBank/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func NewMongoStorage(ctx context.Context, config StorageConfig) (RecipeStorage, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

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

	bsonFilter := bson.M{}
	if filter.Title != "" {
		bsonFilter["title"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(filter.Title), Options: "i"}}
	}
	if len(filter.IngredientNames) > 0 {
		var ingredientQueries []bson.M
		for _, name := range filter.IngredientNames {
			if name != "" { // Additional safety check
				ingredientQuery := bson.M{"ingredients.name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(name), Options: "i"}}}
				ingredientQueries = append(ingredientQueries, ingredientQuery)
			}
		}
		if len(ingredientQueries) > 0 {
			bsonFilter["$and"] = ingredientQueries
		}
	}
	if filter.CookTime > 0 {
		bsonFilter["cook_time"] = bson.M{"$lte": filter.CookTime}
	}
	if len(filter.Tags) > 0 {
		var tagQueries []bson.M
		for _, tag := range filter.Tags {
			if tag != "" { // Additional safety check
				tagQuery := bson.M{"tags": bson.M{"$regex": primitive.Regex{Pattern: "^" + regexp.QuoteMeta(tag) + "$", Options: "i"}}}
				tagQueries = append(tagQueries, tagQuery)
			}
		}
		if len(tagQueries) > 0 {
			if existingAnd, exists := bsonFilter["$and"]; exists {
				bsonFilter["$and"] = append(existingAnd.([]bson.M), tagQueries...)
			} else {
				bsonFilter["$and"] = tagQueries
			}
		}
	}

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
				"image":       recipe.Image,
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
		return fmt.Errorf("%w: recipe with ID %s", ErrNotFound, id)
	}

	return nil
}

func (s *MongoStorage) GetIngredients(ctx context.Context, sort string) ([]models.ResourceSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Build aggregation pipeline
	pipeline := []bson.M{
		// Unwind the ingredients array
		{"$unwind": "$ingredients"},
		// Group by ingredient name and count occurrences
		{
			"$group": bson.M{
				"_id":   "$ingredients.name",
				"count": bson.M{"$sum": 1},
			},
		},
		// Project to match our ResourceSummary model
		{
			"$project": bson.M{
				"_id":        0,
				"name":       "$_id",
				"count":      1,
				"name_upper": bson.M{"$toUpper": "$_id"}, // Convert to uppercase for sorting
			},
		},
	}

	// Add sort stage based on sort parameter
	pipeline = append(pipeline, createSortStage(sort))

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to aggregate ingredients: %v", ErrDatabaseError, err)
	}
	defer cursor.Close(ctx)

	var ingredients []models.ResourceSummary
	if err = cursor.All(ctx, &ingredients); err != nil {
		return nil, fmt.Errorf("%w: failed to decode ingredients: %v", ErrDatabaseError, err)
	}

	return ingredients, nil
}

func (s *MongoStorage) GetTags(ctx context.Context, sort string) ([]models.ResourceSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Build aggregation pipeline
	pipeline := []bson.M{
		// Unwind the tags array
		{"$unwind": "$tags"},
		// Group by tag name and count occurrences
		{
			"$group": bson.M{
				"_id":   "$tags",
				"count": bson.M{"$sum": 1},
			},
		},
		// Project to match our ResourceSummary model
		{
			"$project": bson.M{
				"_id":        0,
				"name":       "$_id",
				"count":      1,
				"name_upper": bson.M{"$toUpper": "$_id"}, // Convert to uppercase for sorting
			},
		},
	}

	// Add sort stage based on sort parameter
	pipeline = append(pipeline, createSortStage(sort))

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to aggregate tags: %v", ErrDatabaseError, err)
	}
	defer cursor.Close(ctx)

	var tags []models.ResourceSummary
	if err = cursor.All(ctx, &tags); err != nil {
		return nil, fmt.Errorf("%w: failed to decode tags: %v", ErrDatabaseError, err)
	}

	return tags, nil
}

func createSortStage(sort string) bson.M {
	var sortStage bson.M
	switch sort {
	case "name_asc":
		sortStage = bson.M{"$sort": bson.M{"name_upper": 1}}
	case "name_desc":
		sortStage = bson.M{"$sort": bson.M{"name_upper": -1}}
	case "count_asc":
		sortStage = bson.M{"$sort": bson.M{"count": 1, "name_upper": 1}} // Secondary sort by name for deterministic results
	case "count_desc":
		sortStage = bson.M{"$sort": bson.M{"count": -1, "name_upper": 1}} // Secondary sort by name for deterministic results
	default:
		sortStage = bson.M{"$sort": bson.M{"name_upper": 1}}
	}

	return sortStage
}
