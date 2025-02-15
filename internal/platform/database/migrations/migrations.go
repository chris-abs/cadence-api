package migrations

import (
	"database/sql"
	"fmt"
)

type Migration struct {
    ID      string
    Enabled bool
    Run     func(*sql.Tx) error
}

type Manager struct {
    db *sql.DB
    migrations []Migration
}

func NewManager(db *sql.DB) *Manager {
    return &Manager{
        db: db,
        migrations: []Migration{
            {
                ID:      "001_item_images",
                Enabled: false,
                Run:     MigrateItemImages,
            },
            {
                ID:      "002_search_indexes",
                Enabled: true,
                Run:     MigrateSearchIndexes,
            },
            {
                ID:      "003_workspace_relationships",
                Enabled: false,
                Run:     MigrateWorkspaceRelationships,
            },
            {
                ID:      "004_container_description",
                Enabled: true,  
                Run:     MigrateContainerDescription,
            },
            {
                ID:      "005_tag_description",
                Enabled: true,
                Run:     MigrateTagDescription,
            },
        },
    }
}

func (m *Manager) EnableMigration(id string) {
    for i := range m.migrations {
        if m.migrations[i].ID == id {
            m.migrations[i].Enabled = true
            return
        }
    }
}

func (m *Manager) Run() error {
    tx, err := m.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    for _, migration := range m.migrations {
        if migration.Enabled {
            if err := migration.Run(tx); err != nil {
                return fmt.Errorf("migration %s failed: %v", migration.ID, err)
            }
        }
    }

    return tx.Commit()
}