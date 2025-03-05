package schema

import (
	"database/sql"
	"fmt"
)

func InitCoreSchema(db *sql.DB) error {
	if err := createUsersTable(db); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if err := createFamilyTables(db); err != nil {
		return fmt.Errorf("failed to create family tables: %v", err)
	}

	if err := createFamilyMembershipTable(db); err != nil {
		return fmt.Errorf("failed to create family membership table: %v", err)
	}

	if err := createFamilyInviteTable(db); err != nil {
		return fmt.Errorf("failed to create family invite table: %v", err)
	}

	if err := createCalendarTable(db); err != nil {
		return fmt.Errorf("failed to create calendar table: %v", err)
	}

	if err := createNotificationTable(db); err != nil {
		return fmt.Errorf("failed to create notification table: %v", err)
	}

	return nil
}

func createUsersTable(db *sql.DB) error {
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

func createFamilyTables(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS family (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
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
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    `

	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}

func createFamilyMembershipTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS family_membership (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        family_id INTEGER REFERENCES family(id) ON DELETE CASCADE,
        role user_role NOT NULL,
        is_owner BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_family_membership_user ON family_membership(user_id);
    CREATE INDEX IF NOT EXISTS idx_family_membership_family ON family_membership(family_id);
    CREATE INDEX IF NOT EXISTS idx_family_membership_owner ON family_membership(family_id, is_owner);
    `
	_, err := db.Exec(query)
	return err
}

func createFamilyInviteTable(db *sql.DB) error {
	query := `
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
        assignee_id INTEGER REFERENCES users(id),
        family_id INTEGER REFERENCES family(id) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_calendar_event_family ON calendar_event(family_id);
    CREATE INDEX IF NOT EXISTS idx_calendar_event_assignee ON calendar_event(assignee_id);
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
        user_id INTEGER REFERENCES users(id),
        family_id INTEGER REFERENCES family(id),
        title VARCHAR(255) NOT NULL,
        message TEXT NOT NULL,
        type VARCHAR(50) NOT NULL,
        source_id INTEGER,
        is_read BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_notification_user ON notification(user_id);
    CREATE INDEX IF NOT EXISTS idx_notification_family ON notification(family_id);
    CREATE INDEX IF NOT EXISTS idx_notification_read ON notification(is_read);
    `
    
    _, err := db.Exec(query)
    return err
}