package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateMediaSourceNameConstraint(tx *sql.Tx) error {
	// Drop the old unique constraint on name (if it exists)
	// Find and drop any unique constraint that only involves the 'name' column
	dropOldConstraintQuery := `
		DO $$ 
		DECLARE
			constraint_name TEXT;
		BEGIN
			-- Find the unique constraint on just the 'name' column
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
		return fmt.Errorf("failed to drop old unique constraint on name: %v", err)
	}

	// Add the new composite unique constraint on (family_id, name)
	addNewConstraintQuery := `
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint 
				WHERE conrelid = 'media_source'::regclass 
				AND conname = 'media_source_family_id_name_key'
			) THEN
				ALTER TABLE media_source
				ADD CONSTRAINT media_source_family_id_name_key 
				UNIQUE (family_id, name);
			END IF;
		END $$;
	`
	if _, err := tx.Exec(addNewConstraintQuery); err != nil {
		return fmt.Errorf("failed to add composite unique constraint: %v", err)
	}

	return nil
}

