package migrations

import (
	"database/sql"
	"fmt"
)

// MigrateMediaSourceProfileScopedConstraint changes the unique constraint on media_source
// from (family_id, name) to (profile_id, name) to support profile-scoped media sources.
// This migration ensures that source names are unique per profile, allowing different
// profiles in the same family to have sources with the same name (e.g., "Netflix").
func MigrateMediaSourceProfileScopedConstraint(tx *sql.Tx) error {
	// Drop the old unique constraint on (family_id, name) if it exists
	dropOldConstraintQuery := `
		DO $$ 
		DECLARE
			constraint_name TEXT;
		BEGIN
			-- Find the unique constraint on (family_id, name)
			SELECT conname INTO constraint_name
			FROM pg_constraint 
			WHERE conrelid = 'media_source'::regclass 
			AND contype = 'u' 
			AND array_length(conkey, 1) = 2
			AND conkey[1] = (
				SELECT attnum FROM pg_attribute 
				WHERE attrelid = 'media_source'::regclass 
				AND attname = 'family_id'
			)
			AND conkey[2] = (
				SELECT attnum FROM pg_attribute 
				WHERE attrelid = 'media_source'::regclass 
				AND attname = 'name'
			);
			
			-- Drop it if found
			IF constraint_name IS NOT NULL THEN
				EXECUTE format('ALTER TABLE media_source DROP CONSTRAINT IF EXISTS %I', constraint_name);
			END IF;
			
			-- Also check for old single-column unique constraint on name (global uniqueness)
			SELECT conname INTO constraint_name
			FROM pg_constraint 
			WHERE conrelid = 'media_source'::regclass 
			AND contype = 'u' 
			AND array_length(conkey, 1) = 1
			AND conkey[1] = (
				SELECT attnum FROM pg_attribute 
				WHERE attrelid = 'media_source'::regclass 
				AND attname = 'name'
			);
			
			-- Drop it if found
			IF constraint_name IS NOT NULL THEN
				EXECUTE format('ALTER TABLE media_source DROP CONSTRAINT IF EXISTS %I', constraint_name);
			END IF;
		END $$;
	`
	if _, err := tx.Exec(dropOldConstraintQuery); err != nil {
		return fmt.Errorf("failed to drop old unique constraint: %v", err)
	}

	// Add the new composite unique constraint on (profile_id, name)
	addNewConstraintQuery := `
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint 
				WHERE conrelid = 'media_source'::regclass 
				AND conname = 'media_source_profile_id_name_key'
			) THEN
				ALTER TABLE media_source
				ADD CONSTRAINT media_source_profile_id_name_key 
				UNIQUE (profile_id, name);
			END IF;
		END $$;
	`
	if _, err := tx.Exec(addNewConstraintQuery); err != nil {
		return fmt.Errorf("failed to add composite unique constraint: %v", err)
	}

	return nil
}

