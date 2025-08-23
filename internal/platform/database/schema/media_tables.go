package schema

import (
	"database/sql"
	"fmt"
)

func InitMediaSchema(db *sql.DB) error {
	if err := createMediaSourceTables(db); err != nil {
		return fmt.Errorf("failed to create media source tables: %v", err)
	}

	if err := createMediaClassificationTables(db); err != nil {
		return fmt.Errorf("failed to create media classification tables: %v", err)
	}

	if err := createMaterialTables(db); err != nil {
		return fmt.Errorf("failed to create material tables: %v", err)
	}

	return nil
}

func createMediaSourceTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS media_source (
		id VARCHAR(36) PRIMARY KEY,  
		name VARCHAR(100) NOT NULL UNIQUE,
		color VARCHAR(20) NOT NULL,
		category VARCHAR(50) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by VARCHAR(36) REFERENCES profile(id)  
	);

	CREATE INDEX IF NOT EXISTS idx_media_source_name ON media_source(name);
	CREATE INDEX IF NOT EXISTS idx_media_source_category ON media_source(category);
	CREATE INDEX IF NOT EXISTS idx_media_source_is_deleted ON media_source(is_deleted);
	`
	_, err := db.Exec(query)
	return err
}

func createMediaClassificationTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS media_classification (
		id VARCHAR(36) PRIMARY KEY,  
		name VARCHAR(100) NOT NULL,
		description TEXT,
		color VARCHAR(20) NOT NULL,
		image_url TEXT,
		family_id VARCHAR(36) REFERENCES family_account(id) NOT NULL,
		created_by VARCHAR(36) REFERENCES profile(id) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by VARCHAR(36) REFERENCES profile(id)  
	);

	CREATE INDEX IF NOT EXISTS idx_media_classification_name ON media_classification(name);
	CREATE INDEX IF NOT EXISTS idx_media_classification_family_id ON media_classification(family_id);
	CREATE INDEX IF NOT EXISTS idx_media_classification_created_by ON media_classification(created_by);
	CREATE INDEX IF NOT EXISTS idx_media_classification_is_deleted ON media_classification(is_deleted);
	CREATE INDEX IF NOT EXISTS idx_media_classification_family_deleted ON media_classification(family_id, is_deleted);
	`
	_, err := db.Exec(query)
	return err
}

func createMaterialTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS material (
		id VARCHAR(36) PRIMARY KEY,  
		name VARCHAR(255) NOT NULL,
		type VARCHAR(20) NOT NULL CHECK (type IN ('movie', 'show')),
		genre VARCHAR(50) NOT NULL,
		release_year INTEGER,
		runtime INTEGER NOT NULL DEFAULT 0,
		poster_url TEXT,
		source_ids JSONB NOT NULL DEFAULT '[]',
		classification_id VARCHAR(36) REFERENCES media_classification(id),
		watch_with VARCHAR(20) NOT NULL CHECK (watch_with IN ('alone', 'partner', 'family')),
		status VARCHAR(30) NOT NULL CHECK (status IN ('to_watch', 'in_progress', 'watching', 'awaiting_release', 'watched')),
		priority VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
		notes TEXT NOT NULL DEFAULT '',
		profile_id VARCHAR(36) REFERENCES profile(id) NOT NULL,  
		family_id VARCHAR(36) REFERENCES family_account(id) NOT NULL,  
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by VARCHAR(36) REFERENCES profile(id)  
	);

	CREATE INDEX IF NOT EXISTS idx_material_profile_id ON material(profile_id);
	CREATE INDEX IF NOT EXISTS idx_material_family_id ON material(family_id);
	CREATE INDEX IF NOT EXISTS idx_material_type ON material(type);
	CREATE INDEX IF NOT EXISTS idx_material_genre ON material(genre);
	CREATE INDEX IF NOT EXISTS idx_material_status ON material(status);
	CREATE INDEX IF NOT EXISTS idx_material_priority ON material(priority);
	CREATE INDEX IF NOT EXISTS idx_material_watch_with ON material(watch_with);
	CREATE INDEX IF NOT EXISTS idx_material_classification_id ON material(classification_id);
	CREATE INDEX IF NOT EXISTS idx_material_source_ids ON material USING gin (source_ids);
	CREATE INDEX IF NOT EXISTS idx_material_name_pattern ON material USING gin (name gin_trgm_ops);
	CREATE INDEX IF NOT EXISTS idx_material_name_fts ON material USING gin (to_tsvector('english', name || ' ' || COALESCE(notes, '')));
	CREATE INDEX IF NOT EXISTS idx_material_is_deleted ON material(is_deleted);
	CREATE INDEX IF NOT EXISTS idx_material_profile_deleted ON material(profile_id, is_deleted);
	CREATE INDEX IF NOT EXISTS idx_material_family_deleted ON material(family_id, is_deleted);
	CREATE INDEX IF NOT EXISTS idx_material_profile_status ON material(profile_id, status, is_deleted);
	CREATE INDEX IF NOT EXISTS idx_material_classification_deleted ON material(classification_id, is_deleted);
	`
	_, err := db.Exec(query)
	return err
}