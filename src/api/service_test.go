package api

import (
	"anymind"
	"anymind/src/mock"
	"context"
	"fmt"
	"github.com/cockroachdb/apd"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func mustApd(number string) apd.Decimal {
	val, _, err := apd.NewFromString(number)
	if err != nil {
		panic(err)
	}

	return *val
}

func TestDepositInvalidValue(t *testing.T) {
	persistSvc := &mock.PersistenceServiceMock{}

	svc := NewService(persistSvc)
	ctx := context.Background()

	testCases := []string{
		"inf",
		"-inf",
		"0",
		"0.0000",
		"-0.0000",
		"nan",
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("invalid amount %s", tc), func(t *testing.T) {
			err := svc.Deposit(ctx, &anymind.DepositInput{
				DateTime: time.Now(),
				Amount:   mustApd(tc),
			})

			anyErr := anymind.ParameterError(nil)
			require.ErrorAs(t, err, &anyErr)
			require.Equal(t, anymind.ParameterErr, anyErr.Type)
		})
	}
}

func TestDepositValidValue(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		input  string
		parsed string
	}{
		{"1.2", "1.2"},
		{"0.00000001", "0.00000001"},
		{"+12", "12"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("invalid amount %s", tc.input), func(t *testing.T) {
			persistSvc := &mock.PersistenceServiceMock{
				DepositFunc: func(ctx context.Context, input *anymind.DepositInput) error {
					require.Equal(t, tc.parsed, fmt.Sprintf("%f", &input.Amount))

					return nil
				},
			}

			svc := NewService(persistSvc)

			err := svc.Deposit(ctx, &anymind.DepositInput{
				DateTime: time.Now(),
				Amount:   mustApd(tc.input),
			})

			require.NoError(t, err)
		})
	}
}

func TestHistoricalInvalidValue(t *testing.T) {
	persistSvc := &mock.PersistenceServiceMock{}

	svc := NewService(persistSvc)
	ctx := context.Background()

	_, err := svc.Historical(ctx, &anymind.HistoricalDataReq{
		Start: time.Now(),
		End:   time.Now().Add(-time.Second),
	})

	anyErr := anymind.ParameterError(nil)
	require.ErrorAs(t, err, &anyErr)
	require.Equal(t, anymind.ParameterErr, anyErr.Type)
}

func TestHistoricalValidValue(t *testing.T) {
	persistSvc := &mock.PersistenceServiceMock{
		HistoricalFunc: func(
			ctx context.Context,
			req *anymind.HistoricalDataReq,
		) ([]*anymind.HistoricalData, error) {
			return []*anymind.HistoricalData{
				&anymind.HistoricalData{
					DateTime: req.End.Truncate(time.Hour),
					Amount:   mustApd("123"),
				},
			}, nil
		},
	}

	svc := NewService(persistSvc)
	ctx := context.Background()

	res, err := svc.Historical(ctx, &anymind.HistoricalDataReq{
		Start: time.Now().Add(-24 * time.Hour).UTC(),
		End:   time.Now(),
	})

	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "123", fmt.Sprintf("%f", &res[0].Amount))
}
