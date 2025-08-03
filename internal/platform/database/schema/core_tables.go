package schema

import (
	"database/sql"
	"fmt"
)

func InitCoreSchema(db *sql.DB) error {
    if err := createFamilyAccountTable(db); err != nil {
        return fmt.Errorf("failed to create family account table: %v", err)
    }

    if err := createProfileTable(db); err != nil {
        return fmt.Errorf("failed to create profile table: %v", err)
    }

    if err := createFamilySettingsTable(db); err != nil {
        return fmt.Errorf("failed to create family settings table: %v", err)
    }

    if err := createNotificationTable(db); err != nil {
        return fmt.Errorf("failed to create notification table: %v", err)
    }

    return nil
}

func createFamilyAccountTable(db *sql.DB) error {
    query := `
        CREATE TABLE IF NOT EXISTS family_account (
        id VARCHAR(36) PRIMARY KEY, 
        email VARCHAR(255) UNIQUE NOT NULL,
        password TEXT NOT NULL,
        family_name VARCHAR(100) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by VARCHAR(36) REFERENCES profile(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_family_account_email ON family_account(email);
    `
    _, err := db.Exec(query)
    return err
}

func createProfileTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS profile (
        id VARCHAR(36) PRIMARY KEY,  -- Changed from SERIAL to VARCHAR(36) for UUID
        family_id VARCHAR(36) REFERENCES family_account(id) ON DELETE CASCADE, 
        name VARCHAR(100) NOT NULL,
        role profile_role NOT NULL,
        pin VARCHAR(6),
        image_url TEXT,
        is_owner BOOLEAN NOT NULL DEFAULT false,
        timezone_name VARCHAR(50) NOT NULL DEFAULT 'UTC',
        colour VARCHAR(50) NOT NULL DEFAULT 'blue', 
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by VARCHAR(36) REFERENCES profile(id) 
    );
    
    CREATE INDEX IF NOT EXISTS idx_profile_family ON profile(family_id);
    CREATE INDEX IF NOT EXISTS idx_profile_owner ON profile(family_id, is_owner);
    `
    _, err := db.Exec(query)
    return err
}

func createFamilySettingsTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS family_settings (
        family_id VARCHAR(36) PRIMARY KEY REFERENCES family_account(id) ON DELETE CASCADE,  
        modules JSONB NOT NULL DEFAULT '{
            "storage": {
                "isEnabled": true
            },
            "chores": {
                "isEnabled": false
            },
            "meals": {
                "isEnabled": false
            },
            "media": {
                "isEnabled": true
            },
            "services": {
                "isEnabled": false
            }
        }'::jsonb,
        status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by VARCHAR(36) REFERENCES profile(id)  
    );
    `
    _, err := db.Exec(query)
    return err
}

func createNotificationTable(db *sql.DB) error {
    query := `
     CREATE TABLE IF NOT EXISTS notification (
        id VARCHAR(36) PRIMARY KEY,
        profile_id VARCHAR(36) REFERENCES profile(id), 
        family_id VARCHAR(36) REFERENCES family_account(id), 
        title VARCHAR(255) NOT NULL,
        message TEXT NOT NULL,
        type VARCHAR(50) NOT NULL,
        source_id VARCHAR(36), 
        is_read BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by VARCHAR(36) REFERENCES profile(id) 
    );
    
    CREATE INDEX IF NOT EXISTS idx_family_account_email ON family_account(email);
    CREATE INDEX IF NOT EXISTS idx_profile_family ON profile(family_id);
    CREATE INDEX IF NOT EXISTS idx_profile_owner ON profile(family_id, is_owner);
    CREATE INDEX IF NOT EXISTS idx_notification_profile ON notification(profile_id);
    CREATE INDEX IF NOT EXISTS idx_notification_family ON notification(family_id);
    CREATE INDEX IF NOT EXISTS idx_notification_read ON notification(is_read);
    `
    _, err := db.Exec(query)
    return err
}
