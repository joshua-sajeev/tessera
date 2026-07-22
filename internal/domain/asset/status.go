package asset

// AssetStatus represents the lifecycle state of an uploaded asset.
type AssetStatus string

const (
	// Uploading indicates the asset is currently being uploaded.
	StatusUploading AssetStatus = "uploading"

	// Uploaded indicates the original asset has been stored successfully.
	StatusUploaded AssetStatus = "uploaded"

	// Processing indicates the asset is being processed into one or more variants.
	StatusProcessing AssetStatus = "processing"

	// Processed indicates all requested processing has completed successfully.
	StatusProcessed AssetStatus = "processed"

	// Failed indicates the upload or processing operation failed.
	StatusFailed AssetStatus = "failed"
)
