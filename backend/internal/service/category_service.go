package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/response"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/utils/slug"
	"go.uber.org/zap"
)

type CategoryService interface {
	List(ctx context.Context) ([]*response.CategoryResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*response.CategoryResponse, error)
	GetBySlug(ctx context.Context, slug string) (*response.CategoryResponse, error)
	Create(ctx context.Context, req request.CreateCategoryRequest) (*response.CategoryResponse, error)
	Update(ctx context.Context, id uuid.UUID, req request.UpdateCategoryRequest) (*response.CategoryResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
	repo    domain.CategoryRepository
	publish domain.EventPublisher
	jobs    domain.SearchIndexJobRepository
	txm     domain.SearchIndexTransactionManager
}

func NewCategoryService(repo domain.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func NewCategoryServiceWithPublisher(repo domain.CategoryRepository, publisher domain.EventPublisher) CategoryService {
	return &categoryService{repo: repo, publish: publisher}
}

func NewCategoryServiceWithJobs(repo domain.CategoryRepository, jobs domain.SearchIndexJobRepository) CategoryService {
	return &categoryService{repo: repo, jobs: jobs}
}

func NewCategoryServiceWithJobsAndTransactions(repo domain.CategoryRepository, jobs domain.SearchIndexJobRepository, txm domain.SearchIndexTransactionManager) CategoryService {
	return &categoryService{repo: repo, jobs: jobs, txm: txm}
}

func NewCategoryServiceWithPublisherAndJobs(repo domain.CategoryRepository, publisher domain.EventPublisher, jobs domain.SearchIndexJobRepository) CategoryService {
	return &categoryService{repo: repo, publish: publisher, jobs: jobs}
}

func NewCategoryServiceWithPublisherJobsAndTransactions(repo domain.CategoryRepository, publisher domain.EventPublisher, jobs domain.SearchIndexJobRepository, txm domain.SearchIndexTransactionManager) CategoryService {
	return &categoryService{repo: repo, publish: publisher, jobs: jobs, txm: txm}
}

func (s *categoryService) List(ctx context.Context) ([]*response.CategoryResponse, error) {
	cats, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*response.CategoryResponse, len(cats))
	for i, cat := range cats {
		c := cat // copy
		res[i] = mapCategoryToResponse(&c)
	}
	return res, nil
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*response.CategoryResponse, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapCategoryToResponse(cat), nil
}

func (s *categoryService) GetBySlug(ctx context.Context, slugStr string) (*response.CategoryResponse, error) {
	cat, err := s.repo.FindBySlug(ctx, slugStr)
	if err != nil {
		return nil, err
	}
	return mapCategoryToResponse(cat), nil
}

func (s *categoryService) Create(ctx context.Context, req request.CreateCategoryRequest) (*response.CategoryResponse, error) {
	categorySlug := slug.GenerateUnique(req.Name, func(candidate string) bool {
		exists, _ := s.repo.ExistsBySlug(ctx, candidate)
		return exists
	})

	cat := &entity.Category{
		Name:     req.Name,
		Slug:     categorySlug,
		ParentID: req.ParentID,
		IconURL:  req.IconURL,
	}

	var (
		created *entity.Category
		err     error
	)
	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			var txErr error
			created, txErr = store.Categories().Create(ctx, cat)
			if txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromCategory(domain.EventTypeCategoryCreated, created)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		created, err = s.repo.Create(ctx, cat)
		if err == nil {
			s.enqueueCategoryIndexJob(ctx, domain.EventTypeCategoryCreated, created)
		}
	}
	if err != nil {
		return nil, err
	}
	s.publishCategoryEvent(ctx, domain.EventTypeCategoryCreated, created)

	return mapCategoryToResponse(created), nil
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req request.UpdateCategoryRequest) (*response.CategoryResponse, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fields := []string{}
	if req.Name != nil && *req.Name != cat.Name {
		cat.Name = *req.Name
		fields = append(fields, "name")
	}
	if req.IconURL != nil {
		cat.IconURL = req.IconURL
		fields = append(fields, "icon_url")
	}

	if len(fields) == 0 {
		return mapCategoryToResponse(cat), nil
	}

	var updated *entity.Category
	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			var txErr error
			updated, txErr = store.Categories().Update(ctx, cat, fields)
			if txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromCategory(domain.EventTypeCategoryUpdated, updated)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		updated, err = s.repo.Update(ctx, cat, fields)
		if err == nil {
			s.enqueueCategoryIndexJob(ctx, domain.EventTypeCategoryUpdated, updated)
		}
	}
	if err != nil {
		return nil, err
	}
	s.publishCategoryEvent(ctx, domain.EventTypeCategoryUpdated, updated)

	return mapCategoryToResponse(updated), nil
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	childCount, err := s.repo.CountChildrenByParent(ctx, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return domain.ErrConflict
	}

	listingCount, err := s.repo.CountListingsByCategory(ctx, id)
	if err != nil {
		return err
	}
	if listingCount > 0 {
		return domain.ErrConflict
	}

	deleted := *cat
	deleted.Children = nil
	if s.txm != nil && s.jobs != nil {
		err = s.txm.WithinTransaction(ctx, func(store domain.SearchIndexTransactionStore) error {
			if txErr := store.Categories().Delete(ctx, id); txErr != nil {
				return txErr
			}
			job, txErr := buildSearchIndexJobFromCategory(domain.EventTypeCategoryDeleted, &deleted)
			if txErr != nil {
				return txErr
			}
			_, txErr = store.Jobs().Enqueue(ctx, job)
			return txErr
		})
	} else {
		err = s.repo.Delete(ctx, id)
		if err == nil {
			s.enqueueCategoryIndexJob(ctx, domain.EventTypeCategoryDeleted, &deleted)
		}
	}
	if err != nil {
		return err
	}
	s.publishCategoryEvent(ctx, domain.EventTypeCategoryDeleted, &deleted)
	return nil
}

func (s *categoryService) publishCategoryEvent(ctx context.Context, eventType string, category *entity.Category) {
	if s.publish == nil || category == nil {
		return
	}
	if err := s.publish.PublishCategoryEvent(ctx, buildCategoryEvent(eventType, category)); err != nil {
		logger.Log.Warn("Failed to publish category event", zap.String("event_type", eventType), zap.String("category_id", category.ID.String()), zap.Error(err))
	}
}

func (s *categoryService) enqueueCategoryIndexJob(ctx context.Context, eventType string, category *entity.Category) {
	if s.jobs == nil || category == nil {
		return
	}
	job, err := buildSearchIndexJobFromCategory(eventType, category)
	if err != nil {
		logger.Log.Warn("Failed to build search index job for category", zap.String("event_type", eventType), zap.String("category_id", category.ID.String()), zap.Error(err))
		return
	}
	if _, err := s.jobs.Enqueue(ctx, job); err != nil {
		logger.Log.Warn("Failed to enqueue search index job for category", zap.String("event_type", eventType), zap.String("category_id", category.ID.String()), zap.Error(err))
	}
}

func mapCategoryToResponse(cat *entity.Category) *response.CategoryResponse {
	r := &response.CategoryResponse{
		ID:        cat.ID,
		Name:      cat.Name,
		Slug:      cat.Slug,
		ParentID:  cat.ParentID,
		IconURL:   cat.IconURL,
		CreatedAt: cat.CreatedAt,
	}
	if cat.Parent != nil {
		r.Parent = &response.CategoryShortResponse{
			ID:      cat.Parent.ID,
			Name:    cat.Parent.Name,
			Slug:    cat.Parent.Slug,
			IconURL: cat.Parent.IconURL,
		}
	}
	for _, child := range cat.Children {
		c := child // copy
		r.Children = append(r.Children, response.CategoryShortResponse{
			ID:      c.ID,
			Name:    c.Name,
			Slug:    c.Slug,
			IconURL: c.IconURL,
		})
	}
	return r
}
