package database

import (
	"fmt"
)

func (db *PostgresDB) initializeDatabaseExtensions() error {
    query := `CREATE EXTENSION IF NOT EXISTS pg_trgm;`
    _, err := db.Exec(query)
    return err
}

func (db *PostgresDB) createEnums() error {
    query := `DO $$ 
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
            CREATE TYPE user_role AS ENUM ('PARENT', 'CHILD');
        END IF;
    END $$;`
    
    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("failed to create enum: %v", err)
    }

    return nil
}

func (db *PostgresDB) createUsersTable() error {
    query := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password TEXT NOT NULL,
        first_name VARCHAR(100),
        last_name VARCHAR(100),
        image_url TEXT,
        role user_role NULL, 
        family_id INTEGER REFERENCES family(id),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
    CREATE INDEX IF NOT EXISTS idx_users_family ON users(family_id);
    `
    _, err := db.Exec(query)
    return err
}


func (db *PostgresDB) createFamilyTables() error {
	query := `
    CREATE TABLE IF NOT EXISTS family (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        owner_id INTEGER, -- Remove the foreign key constraint initially
        modules JSONB NOT NULL DEFAULT '{
            "storage": {
                "isEnabled": true
            },
            "meals": {
                "isEnabled": false
            }
        }'::jsonb,
        status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS family_invite (
        id SERIAL PRIMARY KEY,
        family_id INTEGER REFERENCES family(id) ON DELETE CASCADE,
        email VARCHAR(255) NOT NULL,
        role user_role NOT NULL,
        token VARCHAR(255) UNIQUE NOT NULL,
        expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_family_invite_token ON family_invite(token);
    CREATE INDEX IF NOT EXISTS idx_family_invite_email ON family_invite(email);
    `

	if _, err := db.Exec(query); err != nil {
		return err
	}

	addForeignKeyQuery := `
    ALTER TABLE family
    ADD CONSTRAINT fk_family_owner
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT;
    `

	if _, err := db.Exec(addForeignKeyQuery); err != nil {
		return err
	}

	return nil
}


func (db *PostgresDB) createWorkspaceTable() error {
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
        CREATE INDEX IF NOT EXISTS idx_workspace_name_pattern ON workspace USING gin (name gin_trgm_ops);
        CREATE INDEX IF NOT EXISTS idx_workspace_name_fts 
        ON workspace USING gin (to_tsvector('english', name || ' ' || COALESCE(description, '')));
    `
    _, err := db.Exec(query)
    return err
}

func (db *PostgresDB) createContainerTable() error {
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

func (db *PostgresDB) createItemTables() error {
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

func (db *PostgresDB) InitializeTables() error {
    fmt.Println("Starting enum creation...")
    initFuncs := []func() error{
        db.initializeDatabaseExtensions,
        db.createEnums,
    }

    for _, fn := range initFuncs {
        if err := fn(); err != nil {
            return fmt.Errorf("initialization failed: %v", err)
        }
    }

    fmt.Println("Enums created successfully...")

    // Then create tables
    tableFuncs := []func() error{
        db.createUsersTable,
        db.createFamilyTables,
        db.createWorkspaceTable,
        db.createContainerTable,
        db.createItemTables,
    }

    for _, fn := range tableFuncs {
        if err := fn(); err != nil {
            return fmt.Errorf("table initialization failed: %v", err)
        }
    }

    return nil
}