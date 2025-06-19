package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestContainer struct {
	Container testcontainers.Container
	URI       string
}

// setupTestContainer creates a MongoDB test container
func setupTestContainer(ctx context.Context) (*TestContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(time.Second * 50),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, fmt.Errorf("failed to get container external port: %v", err)
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}

	uri := fmt.Sprintf("mongodb://%s:%s", hostIP, mappedPort.Port())

	return &TestContainer{
		Container: container,
		URI:       uri,
	}, nil
}

// createTestStorage creates a Storage instance for testing
func createTestStorage(t *testing.T) (*MongoStorage, func()) {
	ctx := context.Background()

	container, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test container: %v", err)
	}

	// Create client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(container.URI))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Create storage instance
	storage := &MongoStorage{
		client:     client,
		db:         client.Database("test_db"),
		collection: client.Database("test_db").Collection("recipes"),
	}

	// Initialize storage
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := storage.collection.Drop(ctx); err != nil {
			t.Errorf("Failed to drop test collection: %v", err)
		}
		if err := storage.client.Disconnect(ctx); err != nil {
			t.Errorf("Failed to disconnect client: %v", err)
		}
		if err := container.Container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}

	return storage, cleanup
}
