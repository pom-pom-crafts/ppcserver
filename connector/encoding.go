package connector

// EncodingType represents client connection transport encoding format.
type EncodingType string

const (
	// EncodingTypeJSON represents that data will transport in JSON format.
	EncodingTypeJSON EncodingType = "json"
	// EncodingTypeProtobuf represents that data will transport in Protobuf format.
	EncodingTypeProtobuf EncodingType = "protobuf"
)
