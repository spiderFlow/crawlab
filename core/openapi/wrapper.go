package openapi

import (
	"fmt"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/loopfz/gadgeto/tonic"

	"github.com/gin-gonic/gin"
	"github.com/wI2L/fizz"
)

// FizzWrapper wraps an existing Gin Engine to add OpenAPI functionality
type FizzWrapper struct {
	fizz   *fizz.Fizz
	gin    *gin.Engine
	logger interfaces.Logger
}

// NewFizzWrapper creates a new wrapper around an existing Gin Engine
// This approach ensures we don't break existing functionality
func NewFizzWrapper(engine *gin.Engine) *FizzWrapper {
	// Create a new Fizz instance using the existing Gin engine
	f := fizz.NewFromEngine(engine)
	return &FizzWrapper{
		fizz:   f,
		gin:    engine,
		logger: utils.NewLogger("FizzWrapper"),
	}
}

// GetFizz returns the underlying Fizz instance
func (w *FizzWrapper) GetFizz() *fizz.Fizz {
	return w.fizz
}

// GetGin returns the underlying Gin engine
func (w *FizzWrapper) GetGin() *gin.Engine {
	return w.gin
}

// Response represents an OpenAPI response
type Response struct {
	Description string
	Model       interface{}
}

// RegisterRoute registers a route with OpenAPI documentation
func (w *FizzWrapper) RegisterRoute(method, path string, handler interface{}, id, summary, description string, responses map[int]Response) {
	// Build operation options for OpenAPI documentation
	opts := w.buildOperationOptions(id, summary, description, responses)

	// Register the route with OpenAPI documentation
	switch method {
	case "GET":
		w.fizz.GET(path, opts, tonic.Handler(handler, 200))
	case "POST":
		w.fizz.POST(path, opts, tonic.Handler(handler, 200))
	case "PUT":
		w.fizz.PUT(path, opts, tonic.Handler(handler, 200))
	case "DELETE":
		w.fizz.DELETE(path, opts, tonic.Handler(handler, 200))
	case "PATCH":
		w.fizz.PATCH(path, opts, tonic.Handler(handler, 200))
	case "HEAD":
		w.fizz.HEAD(path, opts, tonic.Handler(handler, 200))
	case "OPTIONS":
		w.fizz.OPTIONS(path, opts, tonic.Handler(handler, 200))
	}
}

// BuildModelResponse builds a standard response model with a specific data type
func (w *FizzWrapper) BuildModelResponse() map[int]Response {
	return map[int]Response{
		400: {
			Description: "Bad Request",
		},
		401: {
			Description: "Unauthorized",
		},
		500: {
			Description: "Internal Server Error",
		},
	}
}

// buildOperationOptions builds the options for a Fizz operation
func (w *FizzWrapper) buildOperationOptions(id, summary, description string, responses map[int]Response) []fizz.OperationOption {
	var opts []fizz.OperationOption

	// Add ID
	if id != "" {
		opts = append(opts, fizz.ID(id))
	}

	// Add summary
	if summary != "" {
		opts = append(opts, fizz.Summary(summary))
	}

	// Add description
	if description != "" {
		opts = append(opts, fizz.Description(description))
	}

	// Add responses
	if responses != nil {
		for status, response := range responses {
			if response.Model != nil {
				opts = append(opts, fizz.Response(fmt.Sprintf("%d", status), response.Description, response.Model, nil, nil))
			} else {
				opts = append(opts, fizz.Response(fmt.Sprintf("%d", status), response.Description, nil, nil, nil))
			}
		}
	}

	return opts
}
