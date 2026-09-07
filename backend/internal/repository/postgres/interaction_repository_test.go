package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetUserInteractionsBySongIDsCombinesAllStates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &interactionRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	ids := []uint64{11, 12}

	mock.ExpectQuery(`SELECT song_id FROM song_user`).
		WithArgs(uint64(7), uint64(11), uint64(12)).
		WillReturnRows(sqlmock.NewRows([]string{"song_id"}).AddRow(11))
	mock.ExpectQuery(`SELECT song_id, type FROM song_reactions`).
		WithArgs(uint64(7), uint64(11), uint64(12)).
		WillReturnRows(sqlmock.NewRows([]string{"song_id", "type"}).
			AddRow(11, 1).
			AddRow(12, -1))
	mock.ExpectQuery(`SELECT song_id, rating FROM song_ratings`).
		WithArgs(uint64(7), uint64(11), uint64(12)).
		WillReturnRows(sqlmock.NewRows([]string{"song_id", "rating"}).
			AddRow(12, 8.5))

	result, err := repo.GetUserInteractionsBySongIDs(context.Background(), 7, ids)
	if err != nil {
		t.Fatal(err)
	}
	if !result[11].IsFavorited || result[11].Reaction != 1 {
		t.Fatalf("unexpected interaction for song 11: %+v", result[11])
	}
	if result[12].Reaction != -1 || result[12].Rating == nil || *result[12].Rating != 8.5 {
		t.Fatalf("unexpected interaction for song 12: %+v", result[12])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetUserInteractionsBySongIDsPropagatesQueryErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &interactionRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	mock.ExpectQuery(`SELECT song_id FROM song_user`).
		WithArgs(uint64(7), uint64(11)).
		WillReturnError(errors.New("database unavailable"))

	_, err = repo.GetUserInteractionsBySongIDs(
		context.Background(),
		7,
		[]uint64{11},
	)
	if err == nil {
		t.Fatal("expected interaction query error")
	}
}
