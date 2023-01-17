package persistence

import (
	"anymind"
	"context"
	"database/sql"
	"fmt"
	"github.com/cockroachdb/apd"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"os"
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

func connTestDB(up []string) *sql.DB {
	db, err := sql.Open("pgx", os.Getenv("PG_DSN"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "DROP SCHEMA IF EXISTS public CASCADE")
	if err != nil {
		panic(err)
	}

	_, err = conn.ExecContext(ctx, "CREATE SCHEMA public")
	if err != nil {
		panic(err)
	}

	for i := range up {
		_, err = conn.ExecContext(ctx, up[i])
		if err != nil {
			panic(err)
		}
	}

	return db
}

func TestDepositReverseInsert(t *testing.T) {
	db := connTestDB(SchemaUp)
	defer db.Close()

	svc := NewService(db)
	now := mustTime("2020-01-01T16:05:00Z")
	ctx := context.Background()

	// list of insertion
	//  2020-01-01 15:56:00.000000 10 10
	//  2020-01-01 15:57:00.000000  9 19
	//  2020-01-01 15:58:00.000000  8 27
	//  2020-01-01 15:59:00.000000  7 34
	//  2020-01-01 16:00:00.000000  6 40 <-
	//  2020-01-01 16:01:00.000000  5 45
	//  2020-01-01 16:02:00.000000  4 49
	//  2020-01-01 16:03:00.000000  3 52
	//  2020-01-01 16:04:00.000000  2 54
	//  2020-01-01 16:05:00.000000  1 55 <-
	//

	for i := 0; i < 10; i++ {
		amount := apd.New(int64(i+1), 0)
		err := svc.Deposit(ctx, &anymind.DepositInput{
			DateTime: now.Add(time.Duration(-i) * time.Minute),
			Amount:   *amount,
		})

		require.NoError(t, err)
	}

	res, err := svc.Historical(ctx, &anymind.HistoricalDataReq{
		Start: mustTime("2020-01-01T15:00:00Z"),
		End:   mustTime("2020-01-01T17:00:00Z"),
	})

	expected := []struct {
		DateTime time.Time
		Amount   string
	}{
		{
			DateTime: mustTime("2020-01-01T16:00:00Z"),
			Amount:   "40",
		},
		{
			DateTime: mustTime("2020-01-01T17:00:00Z"),
			Amount:   "55",
		},
	}

	require.NoError(t, err)
	require.Len(t, res, len(expected))
	for i := range expected {
		amount, _ := res[i].Amount.Reduce(&res[i].Amount)
		require.Equal(t, expected[i].DateTime, res[i].DateTime)
		require.Equal(t, expected[i].Amount, fmt.Sprintf("%f", amount))
	}
}

func TestDepositForwardInsert(t *testing.T) {
	db := connTestDB(SchemaUp)
	defer db.Close()

	svc := NewService(db)
	now := mustTime("2020-01-01T15:56:00Z")
	ctx := context.Background()

	// list of insertion
	//  2020-01-01 15:56:00.000000  1  1
	//  2020-01-01 15:57:00.000000  2  3
	//  2020-01-01 15:58:00.000000  3  6
	//  2020-01-01 15:59:00.000000  4 10
	//  2020-01-01 16:00:00.000000  5 15 <-
	//  2020-01-01 16:01:00.000000  6 21
	//  2020-01-01 16:02:00.000000  7 28
	//  2020-01-01 16:03:00.000000  8 36
	//  2020-01-01 16:04:00.000000  9 45
	//  2020-01-01 16:05:00.000000 10 55 <-
	//

	for i := 0; i < 10; i++ {
		amount := apd.New(int64(i+1), 0)
		err := svc.Deposit(ctx, &anymind.DepositInput{
			DateTime: now.Add(time.Duration(i) * time.Minute),
			Amount:   *amount,
		})

		require.NoError(t, err)
	}

	res, err := svc.Historical(ctx, &anymind.HistoricalDataReq{
		Start: mustTime("2020-01-01T15:00:00Z"),
		End:   mustTime("2020-01-01T17:00:00Z"),
	})

	expected := []struct {
		DateTime time.Time
		Amount   string
	}{
		{
			DateTime: mustTime("2020-01-01T16:00:00Z"),
			Amount:   "15",
		},
		{
			DateTime: mustTime("2020-01-01T17:00:00Z"),
			Amount:   "55",
		},
	}

	require.NoError(t, err)
	require.Len(t, res, len(expected))
	for i := range expected {
		amount, _ := res[i].Amount.Reduce(&res[i].Amount)
		require.Equal(t, expected[i].DateTime, res[i].DateTime)
		require.Equal(t, expected[i].Amount, fmt.Sprintf("%f", amount))
	}
}

func TestDepositMiddleInsert(t *testing.T) {
	db := connTestDB(SchemaUp)
	defer db.Close()

	svc := NewService(db)
	ctx := context.Background()

	insert := []*anymind.DepositInput{
		{
			DateTime: mustTime("2020-01-01T15:55:00Z"),
			Amount:   mustApd("1"),
		},
		{
			DateTime: mustTime("2020-01-01T16:50:00Z"),
			Amount:   mustApd("2"),
		},
		{
			DateTime: mustTime("2020-01-01T15:56:00Z"),
			Amount:   mustApd("3"),
		},
		{
			DateTime: mustTime("2020-01-01T16:10:00Z"),
			Amount:   mustApd("4"),
		},
	}

	for i := range insert {
		err := svc.Deposit(ctx, insert[i])
		require.NoError(t, err)
	}

	res, err := svc.Historical(ctx, &anymind.HistoricalDataReq{
		Start: mustTime("2020-01-01T15:00:00Z"),
		End:   mustTime("2020-01-01T17:00:00Z"),
	})

	expected := []struct {
		DateTime time.Time
		Amount   string
	}{
		{
			DateTime: mustTime("2020-01-01T16:00:00Z"),
			Amount:   "4",
		},
		{
			DateTime: mustTime("2020-01-01T17:00:00Z"),
			Amount:   "10",
		},
	}

	require.NoError(t, err)
	require.Len(t, res, len(expected))
	for i := range expected {
		amount, _ := res[i].Amount.Reduce(&res[i].Amount)
		require.Equal(t, expected[i].DateTime, res[i].DateTime)
		require.Equal(t, expected[i].Amount, fmt.Sprintf("%f", amount))
	}
}

func TestEmpty(t *testing.T) {
	db := connTestDB(SchemaUp)
	defer db.Close()

	svc := NewService(db)
	ctx := context.Background()

	res, err := svc.Historical(ctx, &anymind.HistoricalDataReq{
		Start: mustTime("2020-01-01T15:00:00Z"),
		End:   mustTime("2020-01-01T17:00:00Z"),
	})

	require.NoError(t, err)
	require.Len(t, res, 0)
}
