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
}

func NewCategoryService(repo domain.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func NewCategoryServiceWithPublisher(repo domain.CategoryRepository, publisher domain.EventPublisher) CategoryService {
	return &categoryService{repo: repo, publish: publisher}
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

	created, err := s.repo.Create(ctx, cat)
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

	updated, err := s.repo.Update(ctx, cat, fields)
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

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	deleted := *cat
	deleted.Children = nil
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
