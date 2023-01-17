package api

import "go.uber.org/zap"

type Option func(*Service)

func WithLogger(logger *zap.Logger) Option {
	return func(svc *Service) {
		svc.logger = logger
	}
}
