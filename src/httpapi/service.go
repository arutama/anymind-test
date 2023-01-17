package httpapi

import (
	"anymind"
	"context"
	"fmt"
	transport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

var _ anymind.HTTPService = &Service{}

const depositPath = "/deposit"
const historicalPath = "/historical"

type Service struct {
	api    anymind.APIService
	port   int
	logger *zap.Logger
}

// NewRouter create new router with predefined path and method.
func (s *Service) NewRouter() *mux.Router {
	opt := []transport.ServerOption{
		transport.ServerErrorEncoder(errorHandler(s.logger)),
	}

	root := mux.NewRouter()

	root.Methods(http.MethodPost).Path(depositPath).Handler(transport.NewServer(
		depositEndpoint(s.logger, s.api),
		decoder[depositRequest](s.logger),
		encodeAPIResponse,
		opt...,
	))

	root.Methods(http.MethodPost).Path(historicalPath).Handler(transport.NewServer(
		historicalEndpoint(s.logger, s.api),
		decoder[historicalRequest](s.logger),
		encodeAPIResponse,
		opt...,
	))

	return root
}

func (s *Service) Start(ctx context.Context) error {
	root := s.NewRouter()

	webserver := http.Server{
		Handler: root,
		Addr:    fmt.Sprintf(":%d", s.port),
	}

	go webserver.ListenAndServe()

	<-ctx.Done()

	return webserver.Shutdown(ctx)
}

func NewService(api anymind.APIService, opts ...Option) *Service {
	s := &Service{
		api:  api,
		port: 80,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.logger == nil {
		s.logger = zap.NewNop()
	}

	return s
}
