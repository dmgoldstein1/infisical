package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/doordash-oss/oapi-codegen-dd/v3/pkg/runtime"

	"github.com/infisical/api/internal/libs/errutil"
	"github.com/infisical/api/internal/libs/requestid"
)

// ErrorHandler implements the OapiErrorHandler interface for all handlers.
// Create one instance with a logger and pass it to NewHTTPAdapter for each handler.
type ErrorHandler struct {
	Logger *slog.Logger
}

// NewErrorHandler creates an ErrorHandler with the given logger.
func NewErrorHandler(logger *slog.Logger) *ErrorHandler {
	return &ErrorHandler{Logger: logger}
}

// HandleError translates errors to HTTP responses with proper logging.
// The statusCode parameter is used as a fallback when the error type doesn't
// specify its own status (e.g., for oapi-codegen generated handler errors).
func (h *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	ctx := r.Context()
	reqID := requestid.FromContext(ctx)

	w.Header().Set("Content-Type", "application/json")

	var httpErr *errutil.Error
	if errors.As(err, &httpErr) {
		h.writeHTTPError(ctx, w, reqID, httpErr)
		return
	}

	var validationErrors runtime.ValidationErrors
	if errors.As(err, &validationErrors) {
		h.writeValidationError(ctx, w, reqID, validationErrors)
		return
	}

	// Use the passed-in status code for unknown errors (from oapi-codegen adapter)
	if statusCode >= 400 && statusCode < 500 {
		h.writeClientError(ctx, w, reqID, statusCode, err)
		return
	}

	h.writeInternalError(ctx, w, reqID, err)
}

// writeHTTPError writes an errutil.Error as JSON response.
func (h *ErrorHandler) writeHTTPError(ctx context.Context, w http.ResponseWriter, reqID string, httpErr *errutil.Error) {
	status := httpErr.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}

	resp := Error{
		ReqID:      reqID,
		StatusCode: status,
		ErrorData:  httpErr.Name,
		Message:    httpErr.Message,
	}

	if httpErr.Details != nil {
		if detailsMap, ok := httpErr.Details.(map[string]any); ok {
			resp.Details = detailsMap
		}
	}

	if status >= 500 {
		h.Logger.ErrorContext(ctx, "server error",
			slog.String("reqId", reqID),
			slog.String("name", httpErr.Name),
			slog.Int("status", status),
			slog.String("message", httpErr.Message),
			slog.Any("cause", httpErr.Err),
		)
		resp.Message = "Something went wrong"
	} else {
		h.Logger.WarnContext(ctx, "client error",
			slog.String("reqId", reqID),
			slog.String("name", httpErr.Name),
			slog.Int("status", status),
			slog.String("message", httpErr.Message),
			slog.Any("cause", httpErr.Err),
		)
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeValidationError writes validation errors as a ValidationError response.
func (h *ErrorHandler) writeValidationError(ctx context.Context, w http.ResponseWriter, reqID string, validationErrors runtime.ValidationErrors) {
	// Log only field names + count: validation failures echo back the rejected
	// input values, which may contain secrets (e.g. a password that failed a
	// policy check). Never log the values themselves.
	fieldNames := make([]string, 0, len(validationErrors))
	for _, ve := range validationErrors {
		fieldNames = append(fieldNames, ve.Field)
	}
	h.Logger.WarnContext(ctx, "validation error",
		slog.String("reqId", reqID),
		slog.Int("count", len(validationErrors)),
		slog.Any("fields", fieldNames),
	)

	fieldErrors := make([]FieldError, 0, len(validationErrors))
	for _, ve := range validationErrors {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   ve.Field,
			Message: ve.Message,
		})
	}

	resp := ValidationError{
		ReqID:       reqID,
		StatusCode:  http.StatusBadRequest,
		ErrorData:   "ValidationFailure",
		Message:     "Validation failed",
		FieldErrors: fieldErrors,
	}

	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeClientError writes a 4xx error from the oapi-codegen adapter.
func (h *ErrorHandler) writeClientError(ctx context.Context, w http.ResponseWriter, reqID string, status int, err error) {
	h.Logger.WarnContext(ctx, "client error",
		slog.String("reqId", reqID),
		slog.Int("status", status),
		slog.String("message", err.Error()),
	)

	// Strip common prefixes from validation error messages
	msg := err.Error()
	for _, prefix := range []string{"Query.", "Path.", "Body."} {
		msg = strings.ReplaceAll(msg, prefix, "")
	}

	resp := Error{
		ReqID:      reqID,
		StatusCode: status,
		ErrorData:  "BadRequestError",
		Message:    msg,
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// sensitiveLogPatterns mask credential-shaped values before they reach logs.
// Error chains can embed request headers (Authorization, cookies) or user
// input, so the message is scrubbed while the error type is kept for debugging.
var sensitiveLogPatterns = []*struct {
	re   *regexp.Regexp
	repl string
}{
	{regexp.MustCompile(`(?i)\b(bearer\s+)[A-Za-z0-9\-._~+/=]+`), "${1}[REDACTED]"},
	{regexp.MustCompile(`(?i)\b(basic\s+)[A-Za-z0-9+/=]+`), "${1}[REDACTED]"},
	{regexp.MustCompile(`(?i)("?(?:password|passwd|secret|api[_-]?key|access[_-]?token|client[_-]?secret|authorization)"?\s*[:=]\s*"?)[^"'\s,}]+`), "${1}[REDACTED]"},
}

func redactSensitiveText(s string) string {
	for _, p := range sensitiveLogPatterns {
		s = p.re.ReplaceAllString(s, p.repl)
	}
	return s
}

// writeInternalError writes a generic 500 error.
func (h *ErrorHandler) writeInternalError(ctx context.Context, w http.ResponseWriter, reqID string, err error) {
	h.Logger.ErrorContext(ctx, "unhandled error",
		slog.String("reqId", reqID),
		slog.String("errorType", fmt.Sprintf("%T", err)),
		slog.String("error", redactSensitiveText(err.Error())),
	)

	resp := Error{
		ReqID:      reqID,
		StatusCode: http.StatusInternalServerError,
		ErrorData:  "InternalServerError",
		Message:    "Something went wrong",
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(resp)
}
