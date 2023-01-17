package httpapi

import (
	"anymind"
	"anymind/src/mock"
	"context"
	"fmt"
	"github.com/cockroachdb/apd"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
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

func mustTime(ts string) time.Time {
	val, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		panic(err)
	}

	return val
}

func TestDepositValidValue(t *testing.T) {
	var testCase = []struct {
		name         string
		json         string
		parsedtime   string
		parsedamount string
	}{
		{
			name: "filled",
			json: `
			{
				"datetime": "2020-01-01T00:00:00Z",
				"amount": 123.14223
			}`,
			parsedtime:   "2020-01-01T00:00:00Z",
			parsedamount: "123.14223",
		},
		{
			name: "filled negative",
			json: `
			{
				"datetime": "2020-01-01T08:01:01+07:00",
				"amount": "123.122223"
			}`,
			parsedtime:   "2020-01-01T01:01:01Z",
			parsedamount: "123.122223",
		},
	}

	for _, tc := range testCase {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&mock.APIServiceMock{
				DepositFunc: func(_ context.Context, input *anymind.DepositInput) error {
					require.Equal(t, tc.parsedtime, input.DateTime.Format(time.RFC3339))
					require.Equal(t, tc.parsedamount, fmt.Sprintf("%f", &input.Amount))

					return nil
				},
			})
			router := svc.NewRouter()

			rec := httptest.NewRecorder()

			req, err := http.NewRequest("POST", depositPath, strings.NewReader(tc.json))
			require.NoError(t, err)

			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestDepositInvalidValue(t *testing.T) {
	testCase := []struct {
		name     string
		json     string
		httpcode int
	}{
		{
			name:     "unclosed json",
			json:     `{`,
			httpcode: http.StatusBadRequest,
		},
		{
			name: "empty date",
			json: `
			{
				"datetime": "",
				"amount": "123"
			}`,
			httpcode: http.StatusBadRequest,
		},
		{
			name: "empty amount",
			json: `
			{
				"datetime": "2020-01-01T00:00:00Z",
				"amount": ""
			}`,
			httpcode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCase {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&mock.APIServiceMock{})
			router := svc.NewRouter()

			rec := httptest.NewRecorder()

			req, err := http.NewRequest("POST", depositPath, strings.NewReader(tc.json))
			require.NoError(t, err)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.httpcode, rec.Code)
		})
	}
}

func TestHistoricalValidValue(t *testing.T) {
	var testCase = []struct {
		name        string
		reqjs       string
		parsedstart string
		parsedend   string
		respjs      string
	}{
		{
			name: "filled",
			reqjs: `
			{
				"startDatetime": "2020-01-01T00:00:00Z",
				"endDateTime": "2020-01-01T07:01:00+07:00"
			}`,
			parsedstart: "2020-01-01T00:00:00Z",
			parsedend:   "2020-01-01T00:01:00Z",
			respjs: `
			[{
				"datetime": "2020-01-01T00:00:00Z",
				"amount": "123.456"
			}]`,
		},
	}

	for _, tc := range testCase {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&mock.APIServiceMock{
				HistoricalFunc: func(
					_ context.Context,
					req *anymind.HistoricalDataReq,
				) ([]*anymind.HistoricalData, error) {
					return []*anymind.HistoricalData{
						{
							DateTime: mustTime("2020-01-01T00:00:00Z"),
							Amount:   mustApd("123.456"),
						},
					}, nil
				},
			})
			router := svc.NewRouter()

			rec := httptest.NewRecorder()

			req, err := http.NewRequest("POST", historicalPath, strings.NewReader(tc.reqjs))
			require.NoError(t, err)

			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.JSONEq(t, tc.respjs, rec.Body.String())
		})
	}
}

func TestHistoricalInvalidValue(t *testing.T) {
	testCase := []struct {
		name     string
		json     string
		httpcode int
	}{
		{
			name:     "unclosed json",
			json:     `{`,
			httpcode: http.StatusBadRequest,
		},
		{
			name: "invalid start date",
			json: `
			{
				"startDateTime": "",
				"endDateTime": "2020-01-01T00:00:00Z"
			}`,
			httpcode: http.StatusBadRequest,
		},
		{
			name: "invalid end date",
			json: `
			{
				"startDateTime": "2020-01-01T00:00:00Z",
				"endDateTime": ""
			}`,
			httpcode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCase {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&mock.APIServiceMock{})
			router := svc.NewRouter()

			rec := httptest.NewRecorder()

			req, err := http.NewRequest("POST", historicalPath, strings.NewReader(tc.json))
			require.NoError(t, err)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.httpcode, rec.Code)
		})
	}
}

func TestHistoricalFillValue(t *testing.T) {
	var testCase = []struct {
		data     []*anymind.HistoricalData
		expectjs string
	}{
		{
			data: []*anymind.HistoricalData{
				{
					DateTime: mustTime("2020-01-01T23:00:00Z"),
					Amount:   mustApd("10"),
				},
				{
					DateTime: mustTime("2020-01-02T01:00:00Z"),
					Amount:   mustApd("11"),
				},
			},
			expectjs: `
			[
				{
					"datetime": "2020-01-01T23:00:00Z",
					"amount": "10"
				},
				{
					"datetime": "2020-01-02T00:00:00Z",
					"amount": "10"
				},
				{
					"datetime": "2020-01-02T01:00:00Z",
					"amount": "11"
				},
				{
					"datetime": "2020-01-02T02:00:00Z",
					"amount": "11"
				}
			]`,
		},
		{
			data: []*anymind.HistoricalData{
				{
					DateTime: mustTime("2020-01-01T23:00:00Z"),
					Amount:   mustApd("10"),
				},
			},
			expectjs: `
			[
				{
					"datetime": "2020-01-01T23:00:00Z",
					"amount": "10"
				},
				{
					"datetime": "2020-01-02T00:00:00Z",
					"amount": "10"
				},
				{
					"datetime": "2020-01-02T01:00:00Z",
					"amount": "10"
				},
				{
					"datetime": "2020-01-02T02:00:00Z",
					"amount": "10"
				}
			]`,
		},
		{
			data: []*anymind.HistoricalData{
				{
					DateTime: mustTime("2020-01-02T01:00:00Z"),
					Amount:   mustApd("11"),
				},
			},
			expectjs: `
			[
				{
					"datetime": "2020-01-01T23:00:00Z",
					"amount": "0"
				},
				{
					"datetime": "2020-01-02T00:00:00Z",
					"amount": "0"
				},
				{
					"datetime": "2020-01-02T01:00:00Z",
					"amount": "11"
				},
				{
					"datetime": "2020-01-02T02:00:00Z",
					"amount": "11"
				}
			]`,
		},
		{
			data: []*anymind.HistoricalData{},
			expectjs: `
			[
				{
					"datetime": "2020-01-01T23:00:00Z",
					"amount": "0"
				},
				{
					"datetime": "2020-01-02T00:00:00Z",
					"amount": "0"
				},
				{
					"datetime": "2020-01-02T01:00:00Z",
					"amount": "0"
				},
				{
					"datetime": "2020-01-02T02:00:00Z",
					"amount": "0"
				}
			]`,
		},
	}

	for i, tc := range testCase {
		tc := tc

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			svc := NewService(&mock.APIServiceMock{
				HistoricalFunc: func(
					_ context.Context,
					req *anymind.HistoricalDataReq,
				) ([]*anymind.HistoricalData, error) {
					return tc.data, nil
				},
			})
			router := svc.NewRouter()

			rec := httptest.NewRecorder()

			req, err := http.NewRequest(
				"POST",
				historicalPath,
				strings.NewReader(`
					{
						"startDatetime": "2020-01-01T23:00:00Z",
						"endDatetime": "2020-01-02T02:00:00Z"
					}
				`,
				))
			require.NoError(t, err)

			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.JSONEq(t, tc.expectjs, rec.Body.String())
		})
	}
}
