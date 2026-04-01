package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
)

type noopSearchProjector struct{}

type elasticsearchProjector struct {
	index    string
	client   *searchindex.Client
	listings domain.ListingRepository
}

type categorySearchDocument struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type listingSearchDocument struct {
	ID                 uuid.UUID               `json:"id"`
	UserID             uuid.UUID               `json:"user_id"`
	CategoryID         *uuid.UUID              `json:"category_id,omitempty"`
	Category           *categorySearchDocument `json:"category,omitempty"`
	Title              string                  `json:"title"`
	Slug               string                  `json:"slug"`
	DescriptionExcerpt string                  `json:"description_excerpt,omitempty"`
	TransactionType    string                  `json:"transaction_type"`
	Price              int64                   `json:"price"`
	Currency           string                  `json:"currency"`
	LocationProvince   *string                 `json:"location_province,omitempty"`
	LocationCity       *string                 `json:"location_city,omitempty"`
	LocationDistrict   *string                 `json:"location_district,omitempty"`
	LocationVillage    *string                 `json:"location_village,omitempty"`
	Latitude           *float64                `json:"latitude,omitempty"`
	Longitude          *float64                `json:"longitude,omitempty"`
	BedroomCount       *int                    `json:"bedroom_count,omitempty"`
	BathroomCount      *int                    `json:"bathroom_count,omitempty"`
	LandAreaSqm        *int                    `json:"land_area_sqm,omitempty"`
	BuildingAreaSqm    *int                    `json:"building_area_sqm,omitempty"`
	Status             string                  `json:"status"`
	IsFeatured         bool                    `json:"is_featured"`
	PrimaryImageURL    *string                 `json:"primary_image_url,omitempty"`
	ImageURLs          []string                `json:"image_urls,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
	DeletedAt          *time.Time              `json:"deleted_at,omitempty"`
}

func NewNoopSearchProjector() domain.SearchProjector {
	return noopSearchProjector{}
}

func NewElasticsearchSearchProjector(index string, client *searchindex.Client, listings domain.ListingRepository) (domain.SearchProjector, error) {
	if index == "" {
		return nil, fmt.Errorf("search projector: index is required")
	}
	if client == nil {
		return nil, fmt.Errorf("search projector: search client is required")
	}
	if listings == nil {
		return nil, fmt.Errorf("search projector: listing repository is required")
	}
	return &elasticsearchProjector{index: index, client: client, listings: listings}, nil
}

func (noopSearchProjector) HandleListingEvent(context.Context, domain.ListingEvent) error {
	return nil
}

func (noopSearchProjector) HandleCategoryEvent(context.Context, domain.CategoryEvent) error {
	return nil
}

func (p *elasticsearchProjector) HandleListingEvent(ctx context.Context, event domain.ListingEvent) error {
	if event.Metadata.EventType == domain.EventTypeListingDeleted {
		return p.client.DeleteDocument(ctx, p.index, event.Payload.ID.String())
	}
	return p.client.UpsertDocument(ctx, p.index, event.Payload.ID.String(), listingEventDocument(event))
}

func (p *elasticsearchProjector) HandleCategoryEvent(ctx context.Context, event domain.CategoryEvent) error {
	if event.Payload.ID == uuid.Nil {
		return nil
	}
	listings, err := p.listings.FindByCategoryID(ctx, event.Payload.ID)
	if err != nil {
		return err
	}
	for _, listing := range listings {
		if err := p.client.UpsertDocument(ctx, p.index, listing.ID.String(), listingDocumentFromEntity(listing)); err != nil {
			return err
		}
	}
	return nil
}

func listingEventDocument(event domain.ListingEvent) listingSearchDocument {
	return listingSearchDocument{
		ID:                 event.Payload.ID,
		UserID:             event.Payload.UserID,
		CategoryID:         event.Payload.CategoryID,
		Category:           mapCategoryReference(event.Payload.Category),
		Title:              event.Payload.Title,
		Slug:               event.Payload.Slug,
		DescriptionExcerpt: descriptionExcerpt(event.Payload.Description),
		TransactionType:    event.Payload.TransactionType,
		Price:              event.Payload.Price,
		Currency:           event.Payload.Currency,
		LocationProvince:   event.Payload.LocationProvince,
		LocationCity:       event.Payload.LocationCity,
		LocationDistrict:   event.Payload.LocationDistrict,
		LocationVillage:    event.Payload.LocationVillage,
		Latitude:           event.Payload.Latitude,
		Longitude:          event.Payload.Longitude,
		BedroomCount:       event.Payload.BedroomCount,
		BathroomCount:      event.Payload.BathroomCount,
		LandAreaSqm:        event.Payload.LandAreaSqm,
		BuildingAreaSqm:    event.Payload.BuildingAreaSqm,
		Status:             event.Payload.Status,
		IsFeatured:         event.Payload.IsFeatured,
		PrimaryImageURL:    primaryImageURLFromEvent(event.Payload.Images),
		ImageURLs:          imageURLsFromEvent(event.Payload.Images),
		CreatedAt:          event.Payload.CreatedAt,
		UpdatedAt:          event.Payload.UpdatedAt,
		DeletedAt:          event.Payload.DeletedAt,
	}
}

func listingDocumentFromEntity(listing *entity.Listing) listingSearchDocument {
	if listing == nil {
		return listingSearchDocument{}
	}
	return listingSearchDocument{
		ID:                 listing.ID,
		UserID:             listing.UserID,
		CategoryID:         listing.CategoryID,
		Category:           mapCategoryEntity(listing.Category),
		Title:              listing.Title,
		Slug:               listing.Slug,
		DescriptionExcerpt: descriptionExcerpt(listing.Description),
		TransactionType:    listing.TransactionType,
		Price:              listing.Price,
		Currency:           listing.Currency,
		LocationProvince:   listing.LocationProvince,
		LocationCity:       listing.LocationCity,
		LocationDistrict:   listing.LocationDistrict,
		LocationVillage:    listing.LocationVillage,
		Latitude:           listing.Latitude,
		Longitude:          listing.Longitude,
		BedroomCount:       listing.BedroomCount,
		BathroomCount:      listing.BathroomCount,
		LandAreaSqm:        listing.LandAreaSqm,
		BuildingAreaSqm:    listing.BuildingAreaSqm,
		Status:             listing.Status,
		IsFeatured:         listing.IsFeatured,
		PrimaryImageURL:    primaryImageURLFromEntities(listing.Images),
		ImageURLs:          imageURLsFromEntities(listing.Images),
		CreatedAt:          listing.CreatedAt,
		UpdatedAt:          listing.UpdatedAt,
		DeletedAt:          deletedAtValue(listing.DeletedAt),
	}
}

func RebuildListingIndex(ctx context.Context, repo domain.ListingRepository, client *searchindex.Client, index string, pageSize int) error {
	if repo == nil {
		return fmt.Errorf("search projector: listing repository is required")
	}
	if client == nil {
		return fmt.Errorf("search projector: search client is required")
	}
	if index == "" {
		return fmt.Errorf("search projector: index is required")
	}
	if pageSize <= 0 {
		pageSize = 200
	}
	if err := client.RecreateIndex(ctx, index, ListingIndexMapping()); err != nil {
		return err
	}
	for page := 1; ; page++ {
		listings, total, err := repo.List(ctx, domain.ListingFilter{Page: page, Limit: pageSize})
		if err != nil {
			return err
		}
		for _, listing := range listings {
			if err := client.UpsertDocument(ctx, index, listing.ID.String(), listingDocumentFromEntity(listing)); err != nil {
				return err
			}
		}
		if int64(page*pageSize) >= total || len(listings) == 0 {
			return nil
		}
	}
}

func ListingIndexMapping() map[string]any {
	return map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":                  map[string]any{"type": "keyword"},
				"user_id":             map[string]any{"type": "keyword"},
				"category_id":         map[string]any{"type": "keyword"},
				"category":            map[string]any{"properties": map[string]any{"id": map[string]any{"type": "keyword"}, "name": map[string]any{"type": "keyword"}, "slug": map[string]any{"type": "keyword"}}},
				"title":               map[string]any{"type": "text"},
				"slug":                map[string]any{"type": "keyword"},
				"description_excerpt": map[string]any{"type": "text"},
				"transaction_type":    map[string]any{"type": "keyword"},
				"price":               map[string]any{"type": "long"},
				"currency":            map[string]any{"type": "keyword"},
				"location_province":   map[string]any{"type": "keyword"},
				"location_city":       map[string]any{"type": "keyword"},
				"location_district":   map[string]any{"type": "keyword"},
				"location_village":    map[string]any{"type": "keyword"},
				"latitude":            map[string]any{"type": "double"},
				"longitude":           map[string]any{"type": "double"},
				"bedroom_count":       map[string]any{"type": "integer"},
				"bathroom_count":      map[string]any{"type": "integer"},
				"land_area_sqm":       map[string]any{"type": "integer"},
				"building_area_sqm":   map[string]any{"type": "integer"},
				"status":              map[string]any{"type": "keyword"},
				"is_featured":         map[string]any{"type": "boolean"},
				"primary_image_url":   map[string]any{"type": "keyword", "index": false},
				"image_urls":          map[string]any{"type": "keyword", "index": false},
				"created_at":          map[string]any{"type": "date"},
				"updated_at":          map[string]any{"type": "date"},
				"deleted_at":          map[string]any{"type": "date"},
			},
		},
	}
}

func mapCategoryReference(ref *domain.CategoryEventReference) *categorySearchDocument {
	if ref == nil {
		return nil
	}
	return &categorySearchDocument{ID: ref.ID, Name: ref.Name, Slug: ref.Slug}
}

func mapCategoryEntity(category *entity.Category) *categorySearchDocument {
	if category == nil {
		return nil
	}
	return &categorySearchDocument{ID: category.ID, Name: category.Name, Slug: category.Slug}
}

func descriptionExcerpt(description *string) string {
	if description == nil {
		return ""
	}
	text := strings.TrimSpace(*description)
	if len(text) <= 240 {
		return text
	}
	return text[:240]
}

func imageURLsFromEvent(images []domain.ListingImageEventImage) []string {
	urls := make([]string, 0, len(images))
	for _, image := range images {
		if image.DeletedAt != nil {
			continue
		}
		urls = append(urls, image.URL)
	}
	return urls
}

func primaryImageURLFromEvent(images []domain.ListingImageEventImage) *string {
	for _, image := range images {
		if image.DeletedAt == nil && image.IsPrimary {
			url := image.URL
			return &url
		}
	}
	return nil
}

func imageURLsFromEntities(images []entity.ListingImage) []string {
	urls := make([]string, 0, len(images))
	for _, image := range images {
		if image.DeletedAt.Valid {
			continue
		}
		urls = append(urls, image.URL)
	}
	return urls
}

func primaryImageURLFromEntities(images []entity.ListingImage) *string {
	for _, image := range images {
		if !image.DeletedAt.Valid && image.IsPrimary {
			url := image.URL
			return &url
		}
	}
	return nil
}
