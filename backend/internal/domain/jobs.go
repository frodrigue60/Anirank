package domain

import "context"

type JobsRepository interface {
	// CalculateDailyRankings aggregates views, ratings, and creates snapshots of ranks in ranking_histories.
	CalculateDailyRankings(ctx context.Context) error

	// SnapshotRankingPositions calculates current positions and persists them as prev_rank in songs table.
	SnapshotRankingPositions(ctx context.Context) error
}
