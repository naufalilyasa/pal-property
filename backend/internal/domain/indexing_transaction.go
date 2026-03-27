package domain

import "context"

type SearchIndexTransactionStore interface {
	Listings() ListingRepository
	Categories() CategoryRepository
	Jobs() SearchIndexJobRepository
}

type SearchIndexTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(store SearchIndexTransactionStore) error) error
}
