package httpapi

import (
	"anymind"
	"context"
	"encoding/json"
	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/endpoint"
	"go.uber.org/zap"
	"time"
)

type depositRequest struct {
	DateTime time.Time   `json:"datetime"`
	Amount   json.Number `json:"amount"`
}

type depositResponse depositRequest

func depositEndpoint(logger *zap.Logger, s anymind.APIService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (result interface{}, err error) {
		defer func() {
			if err != nil {
				logger.Error("error deposit request", zap.Error(err))
			} else {
				logger.Info("success deposit request")
			}
		}()

		req := request.(*depositRequest)
		amount, _, err := apd.NewFromString(req.Amount.String())
		if err != nil {
			return nil, anymind.InternalError(err)
		}

		input := anymind.DepositInput{
			DateTime: req.DateTime.UTC(),
			Amount:   *amount,
		}
		err = s.Deposit(ctx, &input)
		if err != nil {
			return nil, err
		}

		return &APIResponse{
			JSONPayload: (*depositResponse)(req),
		}, nil
	}
}
