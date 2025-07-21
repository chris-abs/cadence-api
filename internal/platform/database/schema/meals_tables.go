package schema

import (
	"database/sql"
	"fmt"
)

func InitMealsSchema(db *sql.DB) error {
	if err := createMealTables(db); err != nil {
		return fmt.Errorf("failed to create meal tables: %v", err)
	}

	if err := createMealPlanTables(db); err != nil {
		return fmt.Errorf("failed to create meal plan tables: %v", err)
	}

	if err := createShoppingListTables(db); err != nil {
		return fmt.Errorf("failed to create shopping list tables: %v", err)
	}

	return nil
}

func createMealTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS meal (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		image_url TEXT,
		cuisine VARCHAR(50) NOT NULL,
		subcategory VARCHAR(50),
		prep_time_minutes INTEGER NOT NULL DEFAULT 0,
		cook_time_minutes INTEGER NOT NULL DEFAULT 0,
		instructions JSONB NOT NULL DEFAULT '[]',
		macros JSONB NOT NULL DEFAULT '{}',
		dietary_info JSONB NOT NULL DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	CREATE TABLE IF NOT EXISTS meal_option (
		id SERIAL PRIMARY KEY,
		meal_id INTEGER REFERENCES meal(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		is_default BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	CREATE TABLE IF NOT EXISTS meal_ingredient (
		id SERIAL PRIMARY KEY,
		meal_id INTEGER REFERENCES meal(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		quantity DECIMAL(10,3) NOT NULL,
		unit VARCHAR(20) NOT NULL,
		category VARCHAR(50) NOT NULL,
		ingredient_type VARCHAR(20) NOT NULL CHECK (ingredient_type IN ('primary', 'secondary', 'side')),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	-- Indexes for meal tables
	CREATE INDEX IF NOT EXISTS idx_meal_cuisine ON meal(cuisine);
	CREATE INDEX IF NOT EXISTS idx_meal_subcategory ON meal(subcategory);
	CREATE INDEX IF NOT EXISTS idx_meal_name_pattern ON meal USING gin (name gin_trgm_ops);
	CREATE INDEX IF NOT EXISTS idx_meal_name_fts ON meal USING gin (to_tsvector('english', name || ' ' || COALESCE(description, '')));
	CREATE INDEX IF NOT EXISTS idx_meal_is_deleted ON meal(is_deleted);
	CREATE INDEX IF NOT EXISTS idx_meal_dietary_info ON meal USING gin (dietary_info);

	CREATE INDEX IF NOT EXISTS idx_meal_option_meal_id ON meal_option(meal_id);
	CREATE INDEX IF NOT EXISTS idx_meal_option_is_deleted ON meal_option(is_deleted);

	CREATE INDEX IF NOT EXISTS idx_meal_ingredient_meal_id ON meal_ingredient(meal_id);
	CREATE INDEX IF NOT EXISTS idx_meal_ingredient_type ON meal_ingredient(ingredient_type);
	CREATE INDEX IF NOT EXISTS idx_meal_ingredient_category ON meal_ingredient(category);
	CREATE INDEX IF NOT EXISTS idx_meal_ingredient_is_deleted ON meal_ingredient(is_deleted);
	`
	_, err := db.Exec(query)
	return err
}

func createMealPlanTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS meal_plan (
		id SERIAL PRIMARY KEY,
		family_id INTEGER REFERENCES family_account(id) NOT NULL,
		start_date DATE NOT NULL,
		end_date DATE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		created_by INTEGER REFERENCES profile(id) NOT NULL,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	CREATE TABLE IF NOT EXISTS meal_plan_day (
		id SERIAL PRIMARY KEY,
		meal_plan_id INTEGER REFERENCES meal_plan(id) ON DELETE CASCADE,
		date DATE NOT NULL,
		meal_id INTEGER REFERENCES meal(id),
		adult_count INTEGER NOT NULL DEFAULT 2,
		child_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	-- Indexes for meal plan tables
	CREATE INDEX IF NOT EXISTS idx_meal_plan_family_id ON meal_plan(family_id);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_start_date ON meal_plan(start_date);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_end_date ON meal_plan(end_date);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_family_dates ON meal_plan(family_id, start_date, end_date);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_is_deleted ON meal_plan(is_deleted);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_family_deleted ON meal_plan(family_id, is_deleted);

	CREATE INDEX IF NOT EXISTS idx_meal_plan_day_plan_id ON meal_plan_day(meal_plan_id);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_day_date ON meal_plan_day(date);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_day_meal_id ON meal_plan_day(meal_id);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_day_plan_date ON meal_plan_day(meal_plan_id, date);
	CREATE INDEX IF NOT EXISTS idx_meal_plan_day_is_deleted ON meal_plan_day(is_deleted);

	-- Unique constraint to prevent duplicate meals on same day within a plan
	CREATE UNIQUE INDEX IF NOT EXISTS idx_meal_plan_day_unique
	ON meal_plan_day(meal_plan_id, date) WHERE is_deleted = false;
	`
	_, err := db.Exec(query)
	return err
}

func createShoppingListTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS shopping_list (
		id SERIAL PRIMARY KEY,
		family_id INTEGER REFERENCES family_account(id) NOT NULL,
		meal_plan_id INTEGER REFERENCES meal_plan(id) ON DELETE CASCADE,
		start_date DATE NOT NULL,
		end_date DATE NOT NULL,
		generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	CREATE TABLE IF NOT EXISTS shopping_list_item (
		id SERIAL PRIMARY KEY,
		shopping_list_id INTEGER REFERENCES shopping_list(id) ON DELETE CASCADE,
		ingredient_name VARCHAR(100) NOT NULL,
		total_quantity DECIMAL(10,3) NOT NULL,
		unit VARCHAR(20) NOT NULL,
		category VARCHAR(50) NOT NULL,
		is_completed BOOLEAN NOT NULL DEFAULT false,
		display_order INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		deleted_at TIMESTAMP WITH TIME ZONE,
		deleted_by INTEGER REFERENCES profile(id)
	);

	-- Indexes for shopping list tables
	CREATE INDEX IF NOT EXISTS idx_shopping_list_family_id ON shopping_list(family_id);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_meal_plan_id ON shopping_list(meal_plan_id);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_start_date ON shopping_list(start_date);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_end_date ON shopping_list(end_date);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_family_dates ON shopping_list(family_id, start_date, end_date);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_is_deleted ON shopping_list(is_deleted);

	CREATE INDEX IF NOT EXISTS idx_shopping_list_item_list_id ON shopping_list_item(shopping_list_id);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_item_category ON shopping_list_item(category);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_item_completed ON shopping_list_item(is_completed);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_item_display_order ON shopping_list_item(shopping_list_id, display_order);
	CREATE INDEX IF NOT EXISTS idx_shopping_list_item_is_deleted ON shopping_list_item(is_deleted);

	-- Unique constraint to prevent duplicate shopping lists for same meal plan
	CREATE UNIQUE INDEX IF NOT EXISTS idx_shopping_list_meal_plan_unique
	ON shopping_list(meal_plan_id) WHERE is_deleted = false;
	`
	_, err := db.Exec(query)
	return err
}