package httpapi

import (
	"anymind"
	"context"
	"encoding/json"
	transport "github.com/go-kit/kit/transport/http"
	"go.uber.org/zap"
	"net/http"
)

type APIResponse struct {
	JSONPayload any
	StatusCode  *int
}

type APIErrorResponse struct {
	Message string `json:"message"`
}

func decoder[T any](logger *zap.Logger) func(context.Context, *http.Request) (interface{}, error) {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		var req T
		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			logger.Error("parsing error", zap.Error(err))

			return nil, anymind.ParameterError(err)
		}

		return &req, err
	}
}

func encodeAPIResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := response.(*APIResponse)
	if resp.StatusCode != nil {
		w.WriteHeader(*resp.StatusCode)
	}

	return json.NewEncoder(w).Encode(resp.JSONPayload)
}

var httpInternalServerCode = http.StatusInternalServerError
var httpBadRequestCode = http.StatusBadRequest

func errorHandler(logger *zap.Logger) transport.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		resp := &APIResponse{}

		anyerr, ok := (err).(*anymind.Error)
		if ok {
			switch anyerr.Type {
			case anymind.InternalErr:
				resp.StatusCode = &httpInternalServerCode
			case anymind.ParameterErr:
				resp.StatusCode = &httpBadRequestCode
			default:
				resp.StatusCode = &httpInternalServerCode
			}

			resp.JSONPayload = APIErrorResponse{
				Message: anyerr.Error(),
			}

			encodeAPIResponse(ctx, w, resp)
			return
		}

		resp.JSONPayload = APIErrorResponse{
			Message: err.Error(),
		}

		encodeAPIResponse(ctx, w, resp)
	}
}
