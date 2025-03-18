package schema

import (
	"database/sql"
	"fmt"
)

func InitMealsSchema(db *sql.DB) error {
	if err := createRecipeTable(db); err != nil {
		return fmt.Errorf("failed to create recipe table: %v", err)
	}

	if err := createMealPlanTable(db); err != nil {
		return fmt.Errorf("failed to create meal plan table: %v", err)
	}

	if err := createShoppingListTable(db); err != nil {
		return fmt.Errorf("failed to create shopping list table: %v", err)
	}

	return nil
}

// todo: very much boilerplate - will need significant rework.

func createRecipeTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS recipe (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        instructions TEXT,
        prep_time INTEGER,
        cook_time INTEGER,
        serving_size INTEGER,
        image_url TEXT,
        creator_id INTEGER REFERENCES profile(id),
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        ingredients JSONB NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_recipe_family ON recipe(family_id);
    CREATE INDEX IF NOT EXISTS idx_recipe_creator ON recipe(creator_id);
    CREATE INDEX IF NOT EXISTS idx_recipe_name ON recipe USING gin (name gin_trgm_ops);
    `
    
    _, err := db.Exec(query)
    return err
}

func createMealPlanTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS meal_plan (
        id SERIAL PRIMARY KEY,
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        date DATE NOT NULL,
        meal_type VARCHAR(50) NOT NULL,
        recipe_id INTEGER REFERENCES recipe(id),
        notes TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS meal_plan_assignee (
        meal_plan_id INTEGER REFERENCES meal_plan(id) ON DELETE CASCADE,
        profile_id INTEGER REFERENCES profile(id) ON DELETE CASCADE,
        role VARCHAR(50) NOT NULL,
        PRIMARY KEY (meal_plan_id, profile_id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_meal_plan_family ON meal_plan(family_id);
    CREATE INDEX IF NOT EXISTS idx_meal_plan_date ON meal_plan(date);
    CREATE INDEX IF NOT EXISTS idx_meal_plan_recipe ON meal_plan(recipe_id);
    CREATE INDEX IF NOT EXISTS idx_meal_plan_assignee_profile ON meal_plan_assignee(profile_id);
    `
    
    _, err := db.Exec(query)
    return err
}

func createShoppingListTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS shopping_list (
        id SERIAL PRIMARY KEY,
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        name VARCHAR(255) NOT NULL,
        start_date DATE,
        end_date DATE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS shopping_list_item (
        id SERIAL PRIMARY KEY,
        shopping_list_id INTEGER REFERENCES shopping_list(id) ON DELETE CASCADE,
        item_name VARCHAR(255) NOT NULL,
        quantity VARCHAR(100),
        is_purchased BOOLEAN DEFAULT FALSE,
        purchased_by INTEGER REFERENCES profile(id),
        recipe_id INTEGER REFERENCES recipe(id),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_shopping_list_family ON shopping_list(family_id);
    CREATE INDEX IF NOT EXISTS idx_shopping_list_date ON shopping_list(start_date, end_date);
    CREATE INDEX IF NOT EXISTS idx_shopping_list_item_list ON shopping_list_item(shopping_list_id);
    CREATE INDEX IF NOT EXISTS idx_shopping_list_item_purchased ON shopping_list_item(is_purchased);
    `
    
    _, err := db.Exec(query)
    return err
}
