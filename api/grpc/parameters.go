package grpc

// CustomErrorMetadataKey is the name of the metadata key storing additional error information.
// Metadata value is expected to be a byte-encoded JSON object.
const CustomErrorMetadataKey = "Custom-Error"

// HTTPCodeMetadataKey is the key to use when setting custom HTTP status codes in gRPC metadata.
const HTTPCodeMetadataKey = "X-Http-Code"
