package schema

import (
	"database/sql"
	"fmt"
)

func InitMediaSchema(db *sql.DB) error {
	if err := createMediaTables(db); err != nil {
		return fmt.Errorf("failed to create media tables: %v", err)
	}

	return nil
}

func createMediaTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS media (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(20) NOT NULL CHECK (type IN ('movie', 'show')),
		genre VARCHAR(50) NOT NULL,
		release_year INTEGER,
		runtime INTEGER NOT NULL DEFAULT 0,
		poster_url TEXT,
		sources JSONB NOT NULL DEFAULT '[]',
		watch_with VARCHAR(20) NOT NULL CHECK (watch_with IN ('alone', 'partner', 'family')),
		status VARCHAR(30) NOT NULL CHECK (status IN ('to_watch', 'in_progress', 'watching', 'waiting_for_season', 'watched')),
		priority VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
		notes TEXT NOT NULL DEFAULT '',
		profile_id INTEGER REFERENCES profile(id) NOT NULL,
		family_id INTEGER REFERENCES family_account(id) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	-- Indexes for media table
	CREATE INDEX IF NOT EXISTS idx_media_profile_id ON media(profile_id);
	CREATE INDEX IF NOT EXISTS idx_media_family_id ON media(family_id);
	CREATE INDEX IF NOT EXISTS idx_media_type ON media(type);
	CREATE INDEX IF NOT EXISTS idx_media_genre ON media(genre);
	CREATE INDEX IF NOT EXISTS idx_media_status ON media(status);
	CREATE INDEX IF NOT EXISTS idx_media_priority ON media(priority);
	CREATE INDEX IF NOT EXISTS idx_media_watch_with ON media(watch_with);
	CREATE INDEX IF NOT EXISTS idx_media_sources ON media USING gin (sources);
	CREATE INDEX IF NOT EXISTS idx_media_name_pattern ON media USING gin (name gin_trgm_ops);
	CREATE INDEX IF NOT EXISTS idx_media_name_fts ON media USING gin (to_tsvector('english', name || ' ' || COALESCE(notes, '')));
	CREATE INDEX IF NOT EXISTS idx_media_is_deleted ON media(is_deleted);
	CREATE INDEX IF NOT EXISTS idx_media_profile_deleted ON media(profile_id, is_deleted);
	CREATE INDEX IF NOT EXISTS idx_media_family_deleted ON media(family_id, is_deleted);
	CREATE INDEX IF NOT EXISTS idx_media_profile_status ON media(profile_id, status, is_deleted);
	`
	_, err := db.Exec(query)
	return err
}