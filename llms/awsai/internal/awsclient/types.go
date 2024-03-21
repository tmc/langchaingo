package awsclient

import (
	"net/http"

	smithydocument "github.com/aws/smithy-go/document"
	"github.com/aws/smithy-go/middleware"
)

type noSmithyDocumentSerde = smithydocument.NoSerde

// EndpointResolverOptions is the service endpoint resolver options
type EndpointResolverOptions = interface{}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// EndpointResolver interface for resolving service endpoints.
type EndpointResolver interface{}
type EndpointResolverV2 interface{}
type HTTPSignerV4 interface{}
type AuthSchemeResolver interface{}
type InvokeModelInput struct {

	// Input data in the format specified in the content-type request header. To see
	// the format and content of this field for different models, refer to Inference
	// parameters (https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html)
	// .
	//
	// This member is required.
	Body []byte

	// Identifier of the model.
	//
	// This member is required.
	ModelId *string

	// The desired MIME type of the inference body in the response. The default value
	// is application/json .
	Accept *string

	// The MIME type of the input data in the request. The default value is
	// application/json .
	ContentType *string

	noSmithyDocumentSerde
}

type InvokeModelOutput struct {

	// Inference response from the model in the format specified in the content-type
	// header field. To see the format and content of this field for different models,
	// refer to Inference parameters (https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html)
	// .
	//
	// This member is required.
	Body []byte

	// The MIME type of the inference result.
	//
	// This member is required.
	ContentType *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}
