package asset

// AssetStatus represents the lifecycle state of an uploaded asset.
type AssetStatus string

const (
	// StatusUploading indicates that the asset is currently being uploaded.
	StatusUploading AssetStatus = "uploading"

	// StatusUploaded indicates that the original asset has been stored successfully.
	StatusUploaded AssetStatus = "uploaded"

	// StatusProcessing indicates that the asset is being processed into one or more variants.
	StatusProcessing AssetStatus = "processing"

	// StatusProcessed indicates that all requested processing has completed successfully.
	StatusProcessed AssetStatus = "processed"

	// StatusFailed indicates that the upload or processing operation failed.
	StatusFailed AssetStatus = "failed"
)
