package seeds

import (
	"database/sql"
	"fmt"
)

type Seed struct {
	ID      string
	Enabled bool
	Run     func(*sql.Tx) error
}

type Manager struct {
	db    *sql.DB
	seeds []Seed
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db: db,
		seeds: []Seed{
			{
				ID:      "001_media_sources",
				Enabled: false, 
				Run:     SeedMediaSources,
			},
		},
	}
}

func (m *Manager) EnableSeed(id string) {
	for i := range m.seeds {
		if m.seeds[i].ID == id {
			m.seeds[i].Enabled = true
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

	for _, seed := range m.seeds {
		if seed.Enabled {
			fmt.Printf("Running seed: %s\n", seed.ID)
			if err := seed.Run(tx); err != nil {
				return fmt.Errorf("seed %s failed: %v", seed.ID, err)
			}
		}
	}

	return tx.Commit()
}