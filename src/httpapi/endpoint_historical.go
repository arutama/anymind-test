package httpapi

import (
	"anymind"
	"context"
	"fmt"
	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/endpoint"
	"go.uber.org/zap"
	"time"
)

type historicalRequest struct {
	Start time.Time `json:"startDatetime"`
	End   time.Time `json:"endDateTime"`
}

type historicalEntry struct {
	DateTime time.Time `json:"datetime"`
	Amount   string    `json:"amount"`
}

func historicalEndpoint(logger *zap.Logger, s anymind.APIService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (result interface{}, err error) {
		defer func() {
			if err != nil {
				logger.Error("error deposit request", zap.Error(err))
			} else {
				logger.Info("success deposit request")
			}
		}()

		req := request.(*historicalRequest)
		histreq := &anymind.HistoricalDataReq{
			Start: req.Start.UTC(),
			End:   req.End.UTC(),
		}

		res, err := s.Historical(ctx, histreq)
		if err != nil {
			return nil, err
		}

		return &APIResponse{
			JSONPayload: toHistoricalEntryArray(histreq, res),
		}, nil
	}
}

func toHistoricalEntryArray(req *anymind.HistoricalDataReq, entries []*anymind.HistoricalData) []*historicalEntry {
	var res []*historicalEntry

	pos := 0
	cur := req.Start.Truncate(time.Second).Add(-time.Second).Truncate(time.Hour).Add(time.Hour)
	val := apd.New(0, 0)
	for {
		if cur.After(req.End) {
			break
		}

		if pos < len(entries) && entries[pos].DateTime.Equal(cur) {
			val = &entries[pos].Amount
			pos++
		}

		res = append(res, &historicalEntry{
			DateTime: cur,
			Amount:   fmt.Sprintf("%f", val),
		})

		cur = cur.Add(time.Hour)
	}

	return res
}
