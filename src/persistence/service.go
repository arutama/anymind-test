package persistence

import (
	"anymind"
	"context"
	"database/sql"
	"go.uber.org/zap"
	"time"
)

var _ anymind.PersistenceService = &Service{}

type Service struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewService(db *sql.DB, opts ...Option) *Service {
	s := &Service{
		db: db,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.logger == nil {
		s.logger = zap.NewNop()
	}

	return s
}
func (s Service) Deposit(ctx context.Context, input *anymind.DepositInput) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	adjtime := input.DateTime.Truncate(time.Second)

	_, err = tx.ExecContext(ctx, insertHistoriesQuery, adjtime, input.Amount)
	if err != nil {
		s.logger.Error("failed to execute insertHistoriesQuery", zap.Error(err))

		return err
	}

	_, err = tx.ExecContext(ctx, insertHourlyQuery, adjtime, input.Amount)
	if err != nil {
		s.logger.Error("failed to execute insertHourlyQuery", zap.Error(err))

		return err
	}

	_, err = tx.ExecContext(ctx, updatePostHourlyQuery, adjtime, input.Amount)
	if err != nil {
		s.logger.Error("failed to execute updatePostHourlyQuery", zap.Error(err))

		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s Service) Historical(ctx context.Context, req *anymind.HistoricalDataReq) ([]*anymind.HistoricalData, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	var res []*anymind.HistoricalData
	rows, err := tx.QueryContext(ctx, selectHourlyQuery, req.Start, req.End)
	if err != nil {
		s.logger.Error("failed to execute selectHourlyQuery", zap.Error(err))

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row anymind.HistoricalData
		err = rows.Scan(&row.DateTime, &row.Amount)
		if err != nil {
			s.logger.Error("failed to scan selectHourlyQuery", zap.Error(err))

			return nil, err
		}

		res = append(res, &row)
	}

	if rows.Err() != nil {
		s.logger.Error("error on next selectHourlyQuery", zap.Error(err))

		return nil, err
	}

	return res, nil
}
