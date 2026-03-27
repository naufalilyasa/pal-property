package postgres

import (
	"context"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"gorm.io/gorm"
)

type searchIndexTransactionManager struct {
	db *gorm.DB
}

type searchIndexTransactionStore struct {
	listings   domain.ListingRepository
	categories domain.CategoryRepository
	jobs       domain.SearchIndexJobRepository
}

func NewSearchIndexTransactionManager(db *gorm.DB) domain.SearchIndexTransactionManager {
	return &searchIndexTransactionManager{db: db}
}

func (m *searchIndexTransactionManager) WithinTransaction(ctx context.Context, fn func(store domain.SearchIndexTransactionStore) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := &searchIndexTransactionStore{
			listings:   &listingRepository{db: tx},
			categories: &categoryRepository{db: tx},
			jobs:       &searchIndexJobRepository{db: tx},
		}
		return fn(store)
	})
}

func (s *searchIndexTransactionStore) Listings() domain.ListingRepository {
	return s.listings
}

func (s *searchIndexTransactionStore) Categories() domain.CategoryRepository {
	return s.categories
}

func (s *searchIndexTransactionStore) Jobs() domain.SearchIndexJobRepository {
	return s.jobs
}
