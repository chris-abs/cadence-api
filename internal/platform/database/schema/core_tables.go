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

    if err := createCalendarTable(db); err != nil {
        return fmt.Errorf("failed to create calendar table: %v", err)
    }

    if err := createNotificationTable(db); err != nil {
        return fmt.Errorf("failed to create notification table: %v", err)
    }

    return nil
}

func createFamilyAccountTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS family_account (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password TEXT NOT NULL,
        family_name VARCHAR(100) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by INTEGER REFERENCES family_account(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_family_account_email ON family_account(email);
    `
    _, err := db.Exec(query)
    return err
}

func createProfileTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS profile (
        id SERIAL PRIMARY KEY,
        family_id INTEGER REFERENCES family_account(id) ON DELETE CASCADE,
        name VARCHAR(100) NOT NULL,
        role user_role NOT NULL,
        pin VARCHAR(6),
        image_url TEXT,
        is_owner BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by INTEGER REFERENCES profile(id)
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
        family_id INTEGER PRIMARY KEY REFERENCES family_account(id) ON DELETE CASCADE,
        modules JSONB NOT NULL DEFAULT '{
            "storage": {
                "isEnabled": true
            },
            "meals": {
                "isEnabled": false
            },
            "chores": {
                "isEnabled": false
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
        deleted_by INTEGER REFERENCES profile(id)
    );
    `
    _, err := db.Exec(query)
    return err
}

func createCalendarTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS calendar_event (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        start_time TIMESTAMP WITH TIME ZONE NOT NULL,
        end_time TIMESTAMP WITH TIME ZONE NOT NULL,
        all_day BOOLEAN DEFAULT FALSE,
        source_module VARCHAR(50) NOT NULL,
        source_id INTEGER NOT NULL,
        profile_id INTEGER REFERENCES profile(id),
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by INTEGER REFERENCES profile(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_calendar_event_family ON calendar_event(family_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_profile ON calendar_event(profile_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_source ON calendar_event(source_module, source_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_date ON calendar_event(start_time, end_time);
    `
    _, err := db.Exec(query)
    return err
}

func createNotificationTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS notification (
        id SERIAL PRIMARY KEY,
        profile_id INTEGER REFERENCES profile(id),
        family_id INTEGER REFERENCES family_account(id),
        title VARCHAR(255) NOT NULL,
        message TEXT NOT NULL,
        type VARCHAR(50) NOT NULL,
        source_id INTEGER,
        is_read BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        is_deleted BOOLEAN NOT NULL DEFAULT false,
        deleted_at TIMESTAMP WITH TIME ZONE,
        deleted_by INTEGER REFERENCES profile(id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_notification_profile ON notification(profile_id);
    CREATE INDEX IF NOT EXISTS idx_notification_family ON notification(family_id);
    CREATE INDEX IF NOT EXISTS idx_notification_read ON notification(is_read);
    `
    _, err := db.Exec(query)
    return err
}
