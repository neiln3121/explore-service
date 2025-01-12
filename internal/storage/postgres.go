package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
)

var ErrDecisionNotFound = errors.New("no decisions found for recipient")

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

type Liker struct {
	ID        uint64
	ActorID   string
	UpdatedAt uint64
}

func (s *Storage) PutDecision(ctx context.Context, recipientID, actorID string, liked bool) error {
	_, err := s.db.ExecContext(ctx,
		`
		INSERT INTO decisions (recipient_id, actor_id, liked, updated_at) 
		VALUES ($1, $2, $3, now())
		ON CONFLICT (recipient_id, actor_id) 
		DO UPDATE 
		SET (recipient_id, actor_id, liked, updated_at) = ($1, $2, $3, now());
		`,
		recipientID, actorID, liked)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) PutMutualDecisions(ctx context.Context, recipientID, actorID string, actorLiked, recipientLiked bool) error {
	// In one transaction, insert new decision for recipient and update mutual decision on actor. We use a transaction so we don't go out of sync on error
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`
		INSERT INTO decisions (recipient_id, actor_id, liked, mutually_liked, updated_at) 
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (recipient_id, actor_id) 
		DO UPDATE 
		SET (recipient_id, actor_id, liked, mutually_liked, updated_at) = ($1, $2, $3, $4, now());
		`,
		recipientID, actorID, actorLiked, recipientLiked)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`
		UPDATE decisions 
		SET mutually_liked = $3,
		updated_at = now()
		WHERE recipient_id = $1 AND actor_id = $2;
		`,
		actorID, recipientID, recipientLiked)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetLikedDecisions(ctx context.Context, recipientID string, liked bool, token *uint64, limit *uint32) ([]*Liker, error) {
	queryBuilder := sq.Select("id", "actor_id", "updated_at").
		From("decisions").
		Where(sq.Eq{
			"recipient_id": recipientID,
		}).
		Where(sq.Eq{
			"liked": liked,
		}).
		PlaceholderFormat(sq.Dollar).OrderBy("id DESC")

	if token != nil {
		queryBuilder = queryBuilder.Where(sq.Lt{
			"id": *token,
		})
	}

	if limit != nil {
		queryBuilder = queryBuilder.Limit(uint64(*limit))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err

	}
	defer rows.Close()

	var likers []*Liker
	for rows.Next() {
		var id uint64
		var updated_at time.Time
		var actor_id string
		err := rows.Scan(&id, &actor_id, &updated_at)
		if err != nil {
			return nil, err
		}

		likers = append(likers, &Liker{
			ID:        id,
			ActorID:   actor_id,
			UpdatedAt: uint64(updated_at.Unix()),
		})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return likers, nil
}

func (s *Storage) GetNewLikedDecisions(ctx context.Context, recipientID string, liked bool, token *uint64, limit *uint32) ([]*Liker, error) {
	queryBuilder := sq.Select("id", "actor_id", "updated_at").
		From("decisions").
		Where(sq.Eq{
			"recipient_id": recipientID,
		}).
		Where(sq.Eq{
			"liked": liked,
		}).
		Where(sq.Eq{
			"mutually_liked": nil,
		}).
		PlaceholderFormat(sq.Dollar).OrderBy("id DESC")

	if token != nil {
		queryBuilder = queryBuilder.Where(sq.Lt{
			"id": *token,
		})
	}

	if limit != nil {
		queryBuilder = queryBuilder.Limit(uint64(*limit))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likers []*Liker
	for rows.Next() {
		var id uint64
		var updated_at time.Time
		var actor_id string
		err := rows.Scan(&id, &actor_id, &updated_at)
		if err != nil {
			return nil, err
		}

		likers = append(likers, &Liker{
			ID:        id,
			ActorID:   actor_id,
			UpdatedAt: uint64(updated_at.Unix()),
		})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return likers, nil
}

func (s *Storage) GetLikedDecisionsCount(ctx context.Context, recipientID string, liked bool) (int, error) {
	row := s.db.QueryRowContext(ctx, "SELECT count(*) FROM decisions WHERE recipient_id = $1 AND liked = $2;", recipientID, liked)

	var count int
	err := row.Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrDecisionNotFound
		}
		return 0, err
	}

	return count, nil
}

func (s *Storage) GetLikedDecision(ctx context.Context, recipientID, actorID string) (bool, error) {
	row := s.db.QueryRowContext(ctx, "SELECT liked FROM decisions WHERE recipient_id = $1 AND actor_id = $2;", recipientID, actorID)

	var liked bool
	err := row.Scan(&liked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrDecisionNotFound
		}
		return false, err
	}

	return liked, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
