package cloudinary

import (
	"context"
	"errors"
	"fmt"

	cloudinarysdk "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	adminapi "github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/naufalilyasa/pal-property-backend/pkg/mediaasset"
)

type Config struct {
	CloudName string
	APIKey    string
	APISecret string
}

type uploadAPI interface {
	Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error)
	Destroy(ctx context.Context, params uploader.DestroyParams) (*uploader.DestroyResult, error)
}

type adminAPI interface {
	Asset(ctx context.Context, params adminapi.AssetParams) (*adminapi.AssetResult, error)
}

type Adapter struct {
	uploader uploadAPI
	admin    adminAPI
}

func New(cfg Config) (*Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client, err := cloudinarysdk.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: create client: %w", err)
	}

	return &Adapter{uploader: &client.Upload, admin: &client.Admin}, nil
}

func NewWithUploader(uploaderAPI uploadAPI) (*Adapter, error) {
	if uploaderAPI == nil {
		return nil, errors.New("cloudinary: uploader is required")
	}

	return &Adapter{uploader: uploaderAPI}, nil
}

func (c Config) Validate() error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "cloud_name", value: c.CloudName},
		{name: "api_key", value: c.APIKey},
		{name: "api_secret", value: c.APISecret},
	}

	for _, field := range fields {
		if field.value == "" {
			return fmt.Errorf("cloudinary: %s is required", field.name)
		}
	}
	return nil
}

func (a *Adapter) UploadListingImage(ctx context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error) {
	if input.File == nil {
		return nil, errors.New("cloudinary: file is required")
	}

	params := uploader.UploadParams{
		Folder:       input.Folder,
		PublicID:     input.PublicID,
		ResourceType: normalizeResourceType(input.ResourceType, mediaasset.DefaultResourceType),
		Type:         api.DeliveryType(normalizeDeliveryType(input.DeliveryType)),
	}

	result, err := a.uploader.Upload(ctx, input.File, params)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: upload listing image: %w", err)
	}

	return &mediaasset.UploadResult{
		AssetID:          result.AssetID,
		PublicID:         result.PublicID,
		Version:          int64(result.Version),
		SecureURL:        result.SecureURL,
		ResourceType:     result.ResourceType,
		DeliveryType:     result.Type,
		Format:           result.Format,
		Bytes:            int64(result.Bytes),
		Width:            result.Width,
		Height:           result.Height,
		OriginalFilename: result.OriginalFilename,
		Metadata:         mediaasset.Metadata(result.Metadata),
	}, nil
}

func (a *Adapter) DestroyListingImage(ctx context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error) {
	if input.PublicID == "" {
		return nil, errors.New("cloudinary: public_id is required")
	}

	params := uploader.DestroyParams{
		PublicID:     input.PublicID,
		ResourceType: normalizeResourceType(input.ResourceType, mediaasset.DefaultResourceType),
		Type:         normalizeDeliveryType(input.DeliveryType),
	}
	if input.Invalidate {
		params.Invalidate = api.Bool(true)
	}

	result, err := a.uploader.Destroy(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: destroy listing image: %w", err)
	}

	return &mediaasset.DestroyResult{Result: result.Result}, nil
}

func (a *Adapter) UploadListingVideo(ctx context.Context, input mediaasset.UploadInput) (*mediaasset.UploadResult, error) {
	if input.File == nil {
		return nil, errors.New("cloudinary: file is required")
	}

	params := uploader.UploadParams{
		Folder:       input.Folder,
		PublicID:     input.PublicID,
		ResourceType: normalizeResourceType(input.ResourceType, mediaasset.DefaultVideoResourceType),
		Type:         api.DeliveryType(normalizeDeliveryType(input.DeliveryType)),
	}

	result, err := a.uploader.Upload(ctx, input.File, params)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: upload listing video: %w", err)
	}

	durationSeconds := durationSecondsFromMetadata(result.Metadata)
	if a.admin != nil {
		asset, assetErr := a.admin.Asset(ctx, adminapi.AssetParams{
			AssetType:     api.AssetType(result.ResourceType),
			DeliveryType:  api.DeliveryType(result.Type),
			PublicID:      result.PublicID,
			MediaMetadata: api.Bool(true),
		})
		if assetErr == nil {
			if duration := durationSecondsFromMetadata(asset.VideoMetadata); duration != nil {
				durationSeconds = duration
			}
		}
	}

	return &mediaasset.UploadResult{
		AssetID:          result.AssetID,
		PublicID:         result.PublicID,
		Version:          int64(result.Version),
		SecureURL:        result.SecureURL,
		ResourceType:     result.ResourceType,
		DeliveryType:     result.Type,
		Format:           result.Format,
		Bytes:            int64(result.Bytes),
		Width:            result.Width,
		Height:           result.Height,
		DurationSeconds:  durationSeconds,
		OriginalFilename: result.OriginalFilename,
		Metadata:         mediaasset.Metadata(result.Metadata),
	}, nil
}

func (a *Adapter) DestroyListingVideo(ctx context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error) {
	if input.PublicID == "" {
		return nil, errors.New("cloudinary: public_id is required")
	}

	params := uploader.DestroyParams{
		PublicID:     input.PublicID,
		ResourceType: normalizeResourceType(input.ResourceType, mediaasset.DefaultVideoResourceType),
		Type:         normalizeDeliveryType(input.DeliveryType),
	}
	if input.Invalidate {
		params.Invalidate = api.Bool(true)
	}

	result, err := a.uploader.Destroy(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: destroy listing video: %w", err)
	}

	return &mediaasset.DestroyResult{Result: result.Result}, nil
}

func normalizeResourceType(resourceType, fallback string) string {
	if resourceType == "" {
		return fallback
	}

	return resourceType
}

func normalizeDeliveryType(deliveryType string) string {
	if deliveryType == "" {
		return mediaasset.DefaultDeliveryType
	}

	return deliveryType
}

func intPointerFromFloat(value float64) *int {
	if value <= 0 {
		return nil
	}
	converted := int(value)
	if float64(converted) < value {
		converted++
	}
	return &converted
}

func durationSecondsFromMetadata(metadata map[string]interface{}) *int {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["duration"]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case float64:
		return intPointerFromFloat(value)
	case int:
		converted := value
		return &converted
	case int64:
		converted := int(value)
		return &converted
	default:
		return nil
	}
}
