package mediaasset

import "mime/multipart"

const (
	DefaultResourceType = "image"
	DefaultDeliveryType = "upload"
)

type UploadInput struct {
	File         *multipart.FileHeader
	Folder       string
	PublicID     string
	ResourceType string
	DeliveryType string
}

type UploadResult struct {
	AssetID          string
	PublicID         string
	Version          int64
	SecureURL        string
	ResourceType     string
	DeliveryType     string
	Format           string
	Bytes            int64
	Width            int
	Height           int
	OriginalFilename string
}

type DestroyInput struct {
	PublicID     string
	ResourceType string
	DeliveryType string
	Invalidate   bool
}

type DestroyResult struct {
	Result string
}
