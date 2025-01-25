package database

import "fmt"

func (db *PostgresDB) MigrateItemImages() error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    createTableQuery := `
        CREATE TABLE IF NOT EXISTS item_image (
            id SERIAL PRIMARY KEY,
            item_id INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
            url TEXT NOT NULL,
            display_order INTEGER NOT NULL DEFAULT 0,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_item_image_item_id ON item_image(item_id);
        CREATE INDEX IF NOT EXISTS idx_item_image_display_order ON item_image(item_id, display_order);
    `
    if _, err := tx.Exec(createTableQuery); err != nil {
        return fmt.Errorf("failed to create item_image table: %v", err)
    }

    migrateDataQuery := `
        INSERT INTO item_image (item_id, url, display_order)
        SELECT id, image_url, 0
        FROM item
        WHERE image_url IS NOT NULL AND image_url != '';
    `
    if _, err := tx.Exec(migrateDataQuery); err != nil {
        return fmt.Errorf("failed to migrate existing images: %v", err)
    }

    dropColumnQuery := `ALTER TABLE item DROP COLUMN IF EXISTS image_url;`
    if _, err := tx.Exec(dropColumnQuery); err != nil {
        return fmt.Errorf("failed to drop image_url column: %v", err)
    }

    return tx.Commit()
}