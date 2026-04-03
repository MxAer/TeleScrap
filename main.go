package storage

import (
	"context"
	"fmt"
	"setka/models"
	"time"
)

func (s *Storage) CreatePostTable(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	query := `CREATE TABLE IF NOT EXISTS posts (
        id           VARCHAR(255) PRIMARY KEY NOT NULL,
        creator_id   VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        group_id     VARCHAR(255) REFERENCES groups(id) ON DELETE SET NULL,
        text         VARCHAR(3000) NOT NULL,
        with_files   BOOL NOT NULL DEFAULT false,
        signature    VARCHAR(255) NOT NULL,
        payload_hash VARCHAR(255) NOT NULL,
        created_at   TIMESTAMPTZ DEFAULT NOW()  -- ← Добавили колонку
    )`

	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_posts_creator ON posts(creator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_group ON posts(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC)`,
	}

	for _, q := range indexes {
		if _, err := tx.Exec(ctx, q); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) SavePost(ctx context.Context, post *models.Post) error { // ← pointer
	query := `INSERT INTO posts (
        id, creator_id, group_id, text, with_files, signature, payload_hash, created_at
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    ON CONFLICT (id) DO UPDATE SET 
        text = EXCLUDED.text,
        with_files = EXCLUDED.with_files,
        updated_at = NOW()
    WHERE posts.created_at IS NOT DISTINCT FROM EXCLUDED.created_at`

	_, err := s.pool.Exec(ctx, query,
		post.ID,
		post.CreatorID,
		post.GroupID,
		post.Text,
		post.WithFiles,
		post.Signature,
		post.PayloadHash,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save post: %w", err)
	}

	return nil
}
