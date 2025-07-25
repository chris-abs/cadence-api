package seeds

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/media/entities"
)

func SeedMediaSources(tx *sql.Tx) error {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM media_source WHERE is_deleted = false").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing sources: %v", err)
	}

	if count > 0 {
		fmt.Printf("Media sources already seeded (%d sources found)\n", count)
		return nil 
	}

	query := `
		INSERT INTO media_source (name, color, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	now := time.Now().UTC()
	for _, source := range entities.DefaultSources {
		_, err := tx.Exec(query, source.Name, source.Color, source.Category, now, now)
		if err != nil {
			return fmt.Errorf("error seeding source %s: %v", source.Name, err)
		}
	}

	fmt.Printf("Successfully seeded %d media sources\n", len(entities.DefaultSources))
	return nil
}