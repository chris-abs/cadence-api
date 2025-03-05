package schema

import (
	"database/sql"
	"fmt"
)

func InitStorageSchema(db *sql.DB) error {
	if err := createWorkspaceTable(db); err != nil {
		return fmt.Errorf("failed to create workspace table: %v", err)
	}

	if err := createContainerTable(db); err != nil {
		return fmt.Errorf("failed to create container table: %v", err)
	}

	if err := createItemTables(db); err != nil {
		return fmt.Errorf("failed to create item tables: %v", err)
	}

	return nil
}

func createWorkspaceTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS workspace (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        description TEXT,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        family_id INTEGER REFERENCES family(id) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_workspace_user ON workspace(user_id);
    CREATE INDEX IF NOT EXISTS idx_workspace_family ON workspace(family_id);
    CREATE INDEX IF NOT EXISTS idx_workspace_name_pattern ON workspace USING gin (name gin_trgm_ops);
    CREATE INDEX IF NOT EXISTS idx_workspace_name_fts 
    ON workspace USING gin (to_tsvector('english', name || ' ' || COALESCE(description, '')));
    `
	_, err := db.Exec(query)
	return err
}

func createContainerTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS container (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50),
        description TEXT NOT NULL DEFAULT '',
        qr_code VARCHAR(100) UNIQUE,           
        qr_code_image TEXT,             
        number INTEGER,         
        location VARCHAR(50),
        user_id INTEGER REFERENCES users(id) NOT NULL,
        family_id INTEGER REFERENCES family(id) NOT NULL,
        workspace_id INTEGER REFERENCES workspace(id),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_container_qr_code ON container(qr_code);
    CREATE INDEX IF NOT EXISTS idx_container_family ON container(family_id);
    CREATE INDEX IF NOT EXISTS idx_container_workspace_id ON container(workspace_id);
    CREATE INDEX IF NOT EXISTS idx_container_workspace_user ON container(workspace_id, user_id);
    CREATE INDEX IF NOT EXISTS idx_container_name_pattern ON container USING gin (name gin_trgm_ops);
    CREATE INDEX IF NOT EXISTS idx_container_combined_search 
    ON container USING gin (to_tsvector('english', 
        name || ' ' || COALESCE(description, '') || ' ' || COALESCE(location, '')));
    `
	_, err := db.Exec(query)
	return err
}

func createItemTables(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS tag (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50),
        description TEXT NOT NULL DEFAULT '',
        colour TEXT,
        family_id INTEGER REFERENCES family(id) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_tag_family ON tag(family_id);
    CREATE INDEX IF NOT EXISTS idx_tag_name_pattern ON tag USING gin (name gin_trgm_ops);
    CREATE INDEX IF NOT EXISTS idx_tag_name_fts 
    ON tag USING gin (to_tsvector('english', name || ' ' || COALESCE(description, '')));

    CREATE TABLE IF NOT EXISTS item (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100),
        description TEXT,
        quantity INTEGER,
        container_id INTEGER REFERENCES container(id) ON DELETE CASCADE NULL,
        family_id INTEGER REFERENCES family(id) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS item_image (
        id SERIAL PRIMARY KEY,
        item_id INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
        url TEXT NOT NULL,
        display_order INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS item_tag (
        item_id INTEGER REFERENCES item(id) ON DELETE CASCADE,
        tag_id INTEGER REFERENCES tag(id) ON DELETE CASCADE,
        PRIMARY KEY (item_id, tag_id)
    );

    CREATE INDEX IF NOT EXISTS idx_item_container ON item(container_id);
    CREATE INDEX IF NOT EXISTS idx_item_family ON item(family_id);
    CREATE INDEX IF NOT EXISTS idx_item_container_combined ON item(container_id, created_at DESC);
    CREATE INDEX IF NOT EXISTS idx_item_name_pattern ON item USING gin (name gin_trgm_ops);
    CREATE INDEX IF NOT EXISTS idx_item_name_fts 
    ON item USING gin (to_tsvector('english', name || ' ' || COALESCE(description, '')));
    CREATE INDEX IF NOT EXISTS idx_item_tag_item ON item_tag(item_id);
    CREATE INDEX IF NOT EXISTS idx_item_tag_tag ON item_tag(tag_id);
    CREATE INDEX IF NOT EXISTS idx_item_image_item_id ON item_image(item_id);
    CREATE INDEX IF NOT EXISTS idx_item_image_display_order ON item_image(item_id, display_order);
    `
	_, err := db.Exec(query)
	return err
}