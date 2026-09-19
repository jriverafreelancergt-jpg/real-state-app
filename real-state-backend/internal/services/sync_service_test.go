package services

import (
	"context"
	"errors"
	"real-state-backend/internal/core/domain"
	"sync/atomic"
	"testing"
)

// Mock de WasiProvider
type mockWasiProvider struct {
	FetchPropertiesFunc func(ctx context.Context, page int) ([]domain.Property, int, error)
}

func (m *mockWasiProvider) FetchProperties(ctx context.Context, page int) ([]domain.Property, int, error) {
	return m.FetchPropertiesFunc(ctx, page)
}

// Mock de PropertyRepository
type mockPropertyRepository struct {
	GetByIDFunc                  func(ctx context.Context, id int64) (*domain.Property, error)
	GetAllFunc                   func(ctx context.Context, limit, offset int) ([]domain.Property, error)
	CreateFunc                   func(ctx context.Context, property *domain.Property) error
	UpsertBatchFunc              func(ctx context.Context, properties []domain.Property, syncBatchID string) error
	SweepFunc                    func(ctx context.Context, syncBatchID string, origin string) (int64, error)
	GetLastSyncMetadataFunc      func(ctx context.Context) (*domain.SyncMetadata, error)
	CreateSyncMetadataFunc       func(ctx context.Context, batchID string, executedBy *string) error
	CompleteSyncMetadataFunc     func(ctx context.Context, batchID string, propertiesSynced, propertiesDeactivated int) error
	FailSyncMetadataFunc         func(ctx context.Context, batchID string, errorMessage string) error
	GetActivePropertiesCountFunc func(ctx context.Context) (int, error)
}

func (m *mockPropertyRepository) GetByID(ctx context.Context, id int64) (*domain.Property, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockPropertyRepository) GetAll(ctx context.Context, limit, offset int) ([]domain.Property, error) {
	return m.GetAllFunc(ctx, limit, offset)
}
func (m *mockPropertyRepository) Create(ctx context.Context, property *domain.Property) error {
	return m.CreateFunc(ctx, property)
}
func (m *mockPropertyRepository) UpsertBatch(ctx context.Context, properties []domain.Property, syncBatchID string) error {
	return m.UpsertBatchFunc(ctx, properties, syncBatchID)
}
func (m *mockPropertyRepository) Sweep(ctx context.Context, syncBatchID string, origin string) (int64, error) {
	return m.SweepFunc(ctx, syncBatchID, origin)
}
func (m *mockPropertyRepository) GetLastSyncMetadata(ctx context.Context) (*domain.SyncMetadata, error) {
	if m.GetLastSyncMetadataFunc != nil {
		return m.GetLastSyncMetadataFunc(ctx)
	}
	return nil, nil
}
func (m *mockPropertyRepository) CreateSyncMetadata(ctx context.Context, batchID string, executedBy *string) error {
	if m.CreateSyncMetadataFunc != nil {
		return m.CreateSyncMetadataFunc(ctx, batchID, executedBy)
	}
	return nil
}
func (m *mockPropertyRepository) CompleteSyncMetadata(ctx context.Context, batchID string, propertiesSynced, propertiesDeactivated int) error {
	if m.CompleteSyncMetadataFunc != nil {
		return m.CompleteSyncMetadataFunc(ctx, batchID, propertiesSynced, propertiesDeactivated)
	}
	return nil
}
func (m *mockPropertyRepository) FailSyncMetadata(ctx context.Context, batchID string, errorMessage string) error {
	if m.FailSyncMetadataFunc != nil {
		return m.FailSyncMetadataFunc(ctx, batchID, errorMessage)
	}
	return nil
}
func (m *mockPropertyRepository) GetActivePropertiesCount(ctx context.Context) (int, error) {
	if m.GetActivePropertiesCountFunc != nil {
		return m.GetActivePropertiesCountFunc(ctx)
	}
	return 0, nil
}

func TestSyncAllProperties_Success(t *testing.T) {
	var fetchCount int32
	var upsertCount int32
	var sweepCount int32

	provider := &mockWasiProvider{
		FetchPropertiesFunc: func(ctx context.Context, page int) ([]domain.Property, int, error) {
			atomic.AddInt32(&fetchCount, 1)
			props := []domain.Property{
				{ID: int64(page * 10), Title: "Prop"},
			}
			return props, 3, nil // 3 páginas totales
		},
	}

	repo := &mockPropertyRepository{
		UpsertBatchFunc: func(ctx context.Context, properties []domain.Property, syncBatchID string) error {
			atomic.AddInt32(&upsertCount, 1)
			if syncBatchID == "" {
				t.Error("Expected non-empty syncBatchID")
			}
			return nil
		},
		SweepFunc: func(ctx context.Context, syncBatchID string, origin string) (int64, error) {
			atomic.AddInt32(&sweepCount, 1)
			if origin != "WASI" {
				t.Errorf("Expected origin 'WASI', got %s", origin)
			}
			return 5, nil // 5 eliminaciones lógicas
		},
	}

	syncService := NewPropertySyncService(provider, repo, 2) // 2 workers
	err := syncService.SyncAllProperties(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if atomic.LoadInt32(&fetchCount) != 3 {
		t.Errorf("Expected 3 calls to FetchProperties, got %d", fetchCount)
	}

	if atomic.LoadInt32(&upsertCount) != 3 {
		t.Errorf("Expected 3 calls to UpsertBatch, got %d", upsertCount)
	}

	if atomic.LoadInt32(&sweepCount) != 1 {
		t.Errorf("Expected 1 call to Sweep, got %d", sweepCount)
	}
}

func TestSyncAllProperties_FetcherFailureAbortsSweep(t *testing.T) {
	var fetchCount int32
	var upsertCount int32
	var sweepCount int32

	provider := &mockWasiProvider{
		FetchPropertiesFunc: func(ctx context.Context, page int) ([]domain.Property, int, error) {
			atomic.AddInt32(&fetchCount, 1)
			if page == 2 {
				return nil, 0, errors.New("Wasi server exploded")
			}
			return []domain.Property{{ID: int64(page), Title: "Prop"}}, 3, nil
		},
	}

	repo := &mockPropertyRepository{
		UpsertBatchFunc: func(ctx context.Context, properties []domain.Property, syncBatchID string) error {
			atomic.AddInt32(&upsertCount, 1)
			return nil
		},
		SweepFunc: func(ctx context.Context, syncBatchID string, origin string) (int64, error) {
			atomic.AddInt32(&sweepCount, 1)
			return 0, nil
		},
	}

	syncService := NewPropertySyncService(provider, repo, 2)
	err := syncService.SyncAllProperties(context.Background())

	if err == nil {
		t.Fatal("Expected error due to fetcher failure, got nil")
	}

	if atomic.LoadInt32(&sweepCount) != 0 {
		t.Errorf("Expected Sweep to NOT be called on failure, but it was called %d times", sweepCount)
	}
}
