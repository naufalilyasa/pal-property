package cloudinary

import (
	"context"
	"errors"
	"fmt"

	cloudinarysdk "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
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

type Adapter struct {
	uploader uploadAPI
}

func New(cfg Config) (*Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client, err := cloudinarysdk.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: create client: %w", err)
	}

	return &Adapter{uploader: &client.Upload}, nil
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
		ResourceType: normalizeResourceType(input.ResourceType),
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
	}, nil
}

func (a *Adapter) DestroyListingImage(ctx context.Context, input mediaasset.DestroyInput) (*mediaasset.DestroyResult, error) {
	if input.PublicID == "" {
		return nil, errors.New("cloudinary: public_id is required")
	}

	params := uploader.DestroyParams{
		PublicID:     input.PublicID,
		ResourceType: normalizeResourceType(input.ResourceType),
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

func normalizeResourceType(resourceType string) string {
	if resourceType == "" {
		return mediaasset.DefaultResourceType
	}

	return resourceType
}

func normalizeDeliveryType(deliveryType string) string {
	if deliveryType == "" {
		return mediaasset.DefaultDeliveryType
	}

	return deliveryType
}
