package database

import "fmt"

func (db *PostgresDB) createUsersTable() error {
    query := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password TEXT NOT NULL,
        first_name VARCHAR(100),
        last_name VARCHAR(100),
        image_url TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
    `
    _, err := db.Exec(query)
    return err
}

func (db *PostgresDB) createWorkspaceTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS workspace (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            description TEXT,
            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_workspace_user ON workspace(user_id);
    `

    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("error creating workspace tables: %v", err)
    }

    return nil
}

func (db *PostgresDB) createContainerTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS container (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50),
            qr_code VARCHAR(100) UNIQUE,           
            qr_code_image TEXT,             
            number INTEGER,         
            location VARCHAR(50),
            user_id INTEGER REFERENCES users(id) NOT NULL,
            workspace_id INTEGER REFERENCES workspace(id),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_container_qr_code ON container(qr_code);
    `
    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("error executing create table query: %v", err)
    }

    return nil
}

func (db *PostgresDB) createItemTables() error {
    query := `
        CREATE TABLE IF NOT EXISTS tag (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) UNIQUE,
            colour TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS item (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100),
            description TEXT,
            quantity INTEGER,
            container_id INTEGER REFERENCES container(id) ON DELETE CASCADE NULL,
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
        CREATE INDEX IF NOT EXISTS idx_item_tag_item ON item_tag(item_id);
        CREATE INDEX IF NOT EXISTS idx_item_tag_tag ON item_tag(tag_id);
        CREATE INDEX IF NOT EXISTS idx_item_image_item_id ON item_image(item_id);
        CREATE INDEX IF NOT EXISTS idx_item_image_display_order ON item_image(item_id, display_order);
    `
    _, err := db.Exec(query)
    return err
}