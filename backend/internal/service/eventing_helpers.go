package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
)

func buildListingEvent(eventType string, listing *entity.Listing) domain.ListingEvent {
	images := make([]domain.ListingImageEventImage, 0, len(listing.Images))
	for _, image := range listing.Images {
		img := image
		images = append(images, domain.ListingImageEventImage{
			ID:               img.ID,
			URL:              img.URL,
			AssetID:          img.AssetID,
			PublicID:         img.PublicID,
			Format:           img.Format,
			Bytes:            img.Bytes,
			Width:            img.Width,
			Height:           img.Height,
			OriginalFilename: img.OriginalFilename,
			IsPrimary:        img.IsPrimary,
			SortOrder:        img.SortOrder,
			CreatedAt:        img.CreatedAt,
			DeletedAt:        deletedAtValue(img.DeletedAt),
		})
	}

	var categoryRef *domain.CategoryEventReference
	if listing.Category != nil {
		categoryRef = &domain.CategoryEventReference{ID: listing.Category.ID, Name: listing.Category.Name, Slug: listing.Category.Slug, IconURL: listing.Category.IconURL}
	}

	return domain.ListingEvent{
		Metadata: domain.EventMetadata{EventID: newEventID(), EventType: eventType, AggregateType: domain.AggregateTypeListing, AggregateID: listing.ID, Version: 1, OccurredAt: time.Now().UTC()},
		Payload: domain.ListingEventPayload{
			ID:                listing.ID,
			UserID:            listing.UserID,
			CategoryID:        listing.CategoryID,
			Category:          categoryRef,
			Title:             listing.Title,
			Slug:              listing.Slug,
			Description:       listing.Description,
			TransactionType:   listing.TransactionType,
			Price:             listing.Price,
			Currency:          listing.Currency,
			IsNegotiable:      listing.IsNegotiable,
			SpecialOffers:     []byte(listing.SpecialOffers),
			LocationProvince:  listing.LocationProvince,
			LocationCity:      listing.LocationCity,
			LocationDistrict:  listing.LocationDistrict,
			AddressDetail:     listing.AddressDetail,
			Latitude:          listing.Latitude,
			Longitude:         listing.Longitude,
			BedroomCount:      listing.BedroomCount,
			BathroomCount:     listing.BathroomCount,
			FloorCount:        listing.FloorCount,
			CarportCapacity:   listing.CarportCapacity,
			LandAreaSqm:       listing.LandAreaSqm,
			BuildingAreaSqm:   listing.BuildingAreaSqm,
			CertificateType:   listing.CertificateType,
			Condition:         listing.Condition,
			Furnishing:        listing.Furnishing,
			ElectricalPowerVA: listing.ElectricalPowerVA,
			FacingDirection:   listing.FacingDirection,
			YearBuilt:         listing.YearBuilt,
			Facilities:        []byte(listing.Facilities),
			Status:            listing.Status,
			IsFeatured:        listing.IsFeatured,
			Specifications:    []byte(listing.Specifications),
			ViewCount:         listing.ViewCount,
			Images:            images,
			CreatedAt:         listing.CreatedAt,
			UpdatedAt:         listing.UpdatedAt,
			DeletedAt:         deletedAtValue(listing.DeletedAt),
		},
	}
}

func buildCategoryEvent(eventType string, category *entity.Category) domain.CategoryEvent {
	children := make([]domain.CategoryEventReference, 0, len(category.Children))
	for _, child := range category.Children {
		children = append(children, domain.CategoryEventReference{ID: child.ID, Name: child.Name, Slug: child.Slug, IconURL: child.IconURL})
	}
	var parent *domain.CategoryEventReference
	if category.Parent != nil {
		parent = &domain.CategoryEventReference{ID: category.Parent.ID, Name: category.Parent.Name, Slug: category.Parent.Slug, IconURL: category.Parent.IconURL}
	}
	return domain.CategoryEvent{
		Metadata: domain.EventMetadata{EventID: newEventID(), EventType: eventType, AggregateType: domain.AggregateTypeCategory, AggregateID: category.ID, Version: 1, OccurredAt: time.Now().UTC()},
		Payload:  domain.CategoryEventPayload{ID: category.ID, Name: category.Name, Slug: category.Slug, ParentID: category.ParentID, IconURL: category.IconURL, Parent: parent, Children: children, CreatedAt: category.CreatedAt, DeletedAt: nil},
	}
}

func deletedAtValue(value gorm.DeletedAt) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func newEventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
