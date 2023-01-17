package anymind

import (
	"context"
	"github.com/cockroachdb/apd"
	"time"
)

type DepositInput struct {
	DateTime time.Time
	Amount   apd.Decimal
}

type HistoricalData struct {
	DateTime time.Time
	Amount   apd.Decimal
}

type HistoricalDataReq struct {
	Start time.Time
	End   time.Time
}

// PersistenceService is data persistence service interface.
//
//go:generate moq -out src/mock/mock_persistence_service.go -pkg mock . PersistenceService
type PersistenceService interface {
	Deposit(ctx context.Context, input *DepositInput) error
	Historical(ctx context.Context, req *HistoricalDataReq) ([]*HistoricalData, error)
}

// APIService is deposit service API interface.
//
//go:generate moq -out src/mock/mock_api_service.go -pkg mock . APIService
type APIService interface {
	Deposit(ctx context.Context, input *DepositInput) error
	Historical(ctx context.Context, req *HistoricalDataReq) ([]*HistoricalData, error)
}

// HTTPService provide API to listen and serve http services.
type HTTPService interface {
	// Start service, this will block and only return when service stopped / errored.
	Start(ctx context.Context) error
}
