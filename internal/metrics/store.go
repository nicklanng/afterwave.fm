package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// MAU partition: PK = MAU#YYYY-MM, SK = user_id. One item per user per month (idempotent put).
const mauPKPrefix = "MAU#"

type mauRow struct {
	PK string `dynamo:"pk"`
	SK string `dynamo:"sk"`
}

// MAUStore records user activity by month and provides MAU count.
// It implements auth.ActiveMonthRecorder (RecordActiveMonth) for use by the auth service.
type MAUStore struct {
	db        *infra.Dynamo
	tableName string
}

func NewMAUStore(db *infra.Dynamo, tableName string) *MAUStore {
	return &MAUStore{db: db, tableName: tableName}
}

func (s *MAUStore) tbl() dynamo.Table {
	return s.db.Table(s.tableName)
}

func mauPK(yearMonth string) string {
	return mauPKPrefix + yearMonth
}

// YearMonth returns the current year-month in YYYY-MM format (UTC).
func YearMonth(t time.Time) string {
	return t.UTC().Format("2006-01")
}

// RecordActiveMonth records that the user was active in the given month. Idempotent.
func (s *MAUStore) RecordActiveMonth(ctx context.Context, userID string) error {
	inserted, err := s.RecordActiveMonthIfNew(ctx, userID)
	_ = inserted
	return err
}

// RecordActiveMonthIfNew puts the MAU row only if it does not exist (conditional put).
// Returns inserted=true if the row was new, false if it already existed (or userID empty). Caller can use this to increment a counter once per user per month.
func (s *MAUStore) RecordActiveMonthIfNew(ctx context.Context, userID string) (inserted bool, err error) {
	if userID == "" {
		return false, nil
	}
	ym := YearMonth(time.Now())
	row := mauRow{PK: mauPK(ym), SK: userID}
	err = s.tbl().Put(row).If("attribute_not_exists(sk)").Run(ctx)
	if err != nil {
		if dynamo.IsCondCheckFailed(err) {
			return false, nil // already recorded this month
		}
		return false, err
	}
	return true, nil
}

// CountMAU returns the number of distinct users active in the given year-month (e.g. "2025-02").
func (s *MAUStore) CountMAU(ctx context.Context, yearMonth string) (int, error) {
	if yearMonth == "" {
		return 0, fmt.Errorf("year_month required")
	}
	pk := mauPK(yearMonth)
	var count int
	var row mauRow
	iter := s.tbl().Get("pk", pk).Iter()
	for iter.Next(ctx, &row) {
		count++
	}
	return count, iter.Err()
}
