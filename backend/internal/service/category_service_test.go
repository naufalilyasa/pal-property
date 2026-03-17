package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/mocks"
	"github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func strPtr(s string) *string { return &s }

func TestCategoryService_List_Success(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	cats := []entity.Category{
		{
			ID:   uuid.New(),
			Name: "Residential",
			Slug: "residential",
			Children: []entity.Category{
				{ID: uuid.New(), Name: "House", Slug: "house"},
				{ID: uuid.New(), Name: "Apartment", Slug: "apartment"},
			},
		},
		{
			ID:   uuid.New(),
			Name: "Commercial",
			Slug: "commercial",
			Children: []entity.Category{
				{ID: uuid.New(), Name: "Office", Slug: "office"},
				{ID: uuid.New(), Name: "Shop", Slug: "shop"},
			},
		},
	}

	repo.On("FindAll", mock.Anything).Return(cats, nil)

	res, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Residential", res[0].Name)
	assert.Len(t, res[0].Children, 2)
	assert.Equal(t, "Commercial", res[1].Name)
	assert.Len(t, res[1].Children, 2)
}

func TestCategoryService_List_Empty(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	repo.On("FindAll", mock.Anything).Return([]entity.Category{}, nil)

	res, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 0)
}

func TestCategoryService_GetByID_Success(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	parentID := uuid.New()
	cat := &entity.Category{
		ID:       id,
		Name:     "House",
		Slug:     "house",
		ParentID: &parentID,
		Parent: &entity.Category{
			ID:   parentID,
			Name: "Residential",
			Slug: "residential",
		},
		Children: []entity.Category{
			{ID: uuid.New(), Name: "Small House", Slug: "small-house"},
		},
	}

	repo.On("FindByID", mock.Anything, id).Return(cat, nil)

	res, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
	assert.NotNil(t, res.Parent)
	assert.Equal(t, "Residential", res.Parent.Name)
	assert.Len(t, res.Children, 1)
	assert.Equal(t, "Small House", res.Children[0].Name)
}

func TestCategoryService_GetByID_NotFound(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

	res, err := svc.GetByID(context.Background(), id)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCategoryService_GetBySlug_Success(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	cat := &entity.Category{
		ID:   uuid.New(),
		Name: "Residential",
		Slug: "residential",
	}

	repo.On("FindBySlug", mock.Anything, "residential").Return(cat, nil)

	res, err := svc.GetBySlug(context.Background(), "residential")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "residential", res.Slug)
}

func TestCategoryService_GetBySlug_NotFound(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	repo.On("FindBySlug", mock.Anything, "non-existent").Return(nil, domain.ErrNotFound)

	res, err := svc.GetBySlug(context.Background(), "non-existent")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCategoryService_Create_Success(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	req := request.CreateCategoryRequest{
		Name: "Residential",
	}

	repo.On("ExistsBySlug", mock.Anything, "residential").Return(false, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
		return c.Name == req.Name && c.Slug == "residential"
	})).Return(&entity.Category{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: "residential",
	}, nil)

	res, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "residential", res.Slug)
}

func TestCategoryService_Create_SlugCollision(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	req := request.CreateCategoryRequest{
		Name: "Residential",
	}

	// First attempt exists, second one (with suffix) does not
	repo.On("ExistsBySlug", mock.Anything, "residential").Return(true, nil).Once()
	repo.On("ExistsBySlug", mock.Anything, mock.MatchedBy(func(s string) bool {
		return len(s) > len("residential")
	})).Return(false, nil).Once()

	repo.On("Create", mock.Anything, mock.Anything).Return(&entity.Category{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: "residential-random",
	}, nil)

	res, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, "residential", res.Slug)
}

func TestCategoryService_Create_WithParentID(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	parentID := uuid.New()
	req := request.CreateCategoryRequest{
		Name:     "House",
		ParentID: &parentID,
	}

	repo.On("ExistsBySlug", mock.Anything, "house").Return(false, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
		return c.Name == req.Name && c.ParentID != nil && *c.ParentID == parentID
	})).Return(&entity.Category{
		ID:       uuid.New(),
		Name:     req.Name,
		Slug:     "house",
		ParentID: &parentID,
	}, nil)

	res, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, &parentID, res.ParentID)
}

func TestCategoryService_Create_RepoError(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	req := request.CreateCategoryRequest{
		Name: "Residential",
	}

	repo.On("ExistsBySlug", mock.Anything, mock.Anything).Return(false, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrConflict)

	res, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestCategoryService_Update_Success_NameOnly(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	existing := &entity.Category{
		ID:   id,
		Name: "Old Name",
		Slug: "old-name",
	}

	req := request.UpdateCategoryRequest{
		Name: strPtr("New Name"),
	}

	repo.On("FindByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
		return c.Name == "New Name"
	}), []string{"name"}).Return(&entity.Category{
		ID:   id,
		Name: "New Name",
		Slug: "old-name",
	}, nil)

	res, err := svc.Update(context.Background(), id, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New Name", res.Name)
}

func TestCategoryService_Update_Success_IconURL(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	existing := &entity.Category{
		ID:   id,
		Name: "Residential",
		Slug: "residential",
	}

	req := request.UpdateCategoryRequest{
		IconURL: strPtr("https://example.com/icon.png"),
	}

	repo.On("FindByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
		return c.IconURL != nil && *c.IconURL == "https://example.com/icon.png"
	}), []string{"icon_url"}).Return(&entity.Category{
		ID:      id,
		Name:    "Residential",
		Slug:    "residential",
		IconURL: strPtr("https://example.com/icon.png"),
	}, nil)

	res, err := svc.Update(context.Background(), id, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://example.com/icon.png", *res.IconURL)
}

func TestCategoryService_Update_Success_Both(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	existing := &entity.Category{
		ID:   id,
		Name: "Old Name",
		Slug: "old-name",
	}

	req := request.UpdateCategoryRequest{
		Name:    strPtr("New Name"),
		IconURL: strPtr("https://example.com/icon.png"),
	}

	repo.On("FindByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
		return c.Name == "New Name" && *c.IconURL == "https://example.com/icon.png"
	}), []string{"name", "icon_url"}).Return(&entity.Category{
		ID:      id,
		Name:    "New Name",
		Slug:    "old-name",
		IconURL: strPtr("https://example.com/icon.png"),
	}, nil)

	res, err := svc.Update(context.Background(), id, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New Name", res.Name)
	assert.Equal(t, "https://example.com/icon.png", *res.IconURL)
}

func TestCategoryService_Update_NoChanges(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	existing := &entity.Category{
		ID:   id,
		Name: "Residential",
		Slug: "residential",
	}

	// Case 1: All nil
	req1 := request.UpdateCategoryRequest{}
	repo.On("FindByID", mock.Anything, id).Return(existing, nil).Once()

	res1, err1 := svc.Update(context.Background(), id, req1)
	assert.NoError(t, err1)
	assert.Equal(t, "Residential", res1.Name)

	// Case 2: Same name
	req2 := request.UpdateCategoryRequest{
		Name: strPtr("Residential"),
	}
	repo.On("FindByID", mock.Anything, id).Return(existing, nil).Once()

	res2, err2 := svc.Update(context.Background(), id, req2)
	assert.NoError(t, err2)
	assert.Equal(t, "Residential", res2.Name)

	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestCategoryService_Update_NotFound(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

	req := request.UpdateCategoryRequest{
		Name: strPtr("New Name"),
	}

	res, err := svc.Update(context.Background(), id, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCategoryService_Delete_Success(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(&entity.Category{ID: id}, nil)
	repo.On("CountChildrenByParent", mock.Anything, id).Return(int64(0), nil)
	repo.On("CountListingsByCategory", mock.Anything, id).Return(int64(0), nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
}

func TestCategoryService_Delete_HasChildren(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(&entity.Category{ID: id}, nil)
	repo.On("CountChildrenByParent", mock.Anything, id).Return(int64(1), nil)

	err := svc.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConflict)
	repo.AssertNotCalled(t, "CountListingsByCategory", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestCategoryService_Delete_HasListings(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(&entity.Category{ID: id}, nil)
	repo.On("CountChildrenByParent", mock.Anything, id).Return(int64(0), nil)
	repo.On("CountListingsByCategory", mock.Anything, id).Return(int64(1), nil)

	err := svc.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConflict)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestCategoryService_Delete_NotFound(t *testing.T) {
	repo := mocks.NewCategoryRepository(t)
	svc := service.NewCategoryService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

	err := svc.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
