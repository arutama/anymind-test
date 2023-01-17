package api

import (
	"anymind"
	"context"
	"errors"
	"github.com/cockroachdb/apd"
	"go.uber.org/zap"
)

var _ anymind.APIService = &Service{}

var (
	ErrInvalid  = errors.New("invalid argument")
	ErrNotFound = errors.New("not found")
	ErrRequest  = errors.New("bad request")
)

type Service struct {
	persistence anymind.PersistenceService
	logger      *zap.Logger
}

func NewService(persistence anymind.PersistenceService, opts ...Option) *Service {
	s := &Service{
		persistence: persistence,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.logger == nil {
		s.logger = zap.NewNop()
	}

	return s
}

func (s *Service) Deposit(ctx context.Context, input *anymind.DepositInput) error {
	if input.Amount.Form != apd.Finite ||
		input.Amount.Negative ||
		input.Amount.IsZero() {
		return anymind.ParameterError(errors.New("invalid amount"))
	}

	err := s.persistence.Deposit(ctx, input)
	if err != nil {
		return anymind.InternalError(err)
	}

	return nil
}

func (s *Service) Historical(ctx context.Context, req *anymind.HistoricalDataReq) ([]*anymind.HistoricalData, error) {
	if req.Start.After(req.End) {
		return nil, anymind.ParameterError(errors.New("invalid start and end date"))
	}

	res, err := s.persistence.Historical(ctx, req)
	if err != nil {
		return res, anymind.InternalError(err)
	}

	return res, nil
}
