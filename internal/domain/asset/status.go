package asset

// AssetStatus represents the lifecycle state of an uploaded asset.
type AssetStatus string

const (
	// Uploading indicates the asset is currently being uploaded.
	Uploading AssetStatus = "uploading"

	// Uploaded indicates the original asset has been stored successfully.
	Uploaded AssetStatus = "uploaded"

	// Processing indicates the asset is being processed into one or more variants.
	Processing AssetStatus = "processing"

	// Ready indicates all requested processing has completed successfully.
	Ready AssetStatus = "ready"

	// Failed indicates the upload or processing operation failed.
	Failed AssetStatus = "failed"
)
