
package meals

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chrisabs/cadence/internal/meals/entities"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}



func (r *Repository) GetMealByID(id int) (*entities.Meal, error) {
	query := `
		SELECT m.id, m.name, m.description, m.image_url, m.cuisine, m.subcategory,
			   m.prep_time_minutes, m.cook_time_minutes, m.instructions, m.macros, 
			   m.dietary_info, m.created_at, m.updated_at
		FROM meal m
		WHERE m.id = $1 AND m.is_deleted = false`

	meal := new(entities.Meal)
	var instructionsJSON, macrosJSON, dietaryInfoJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&meal.ID, &meal.Name, &meal.Description, &meal.ImageURL, &meal.Cuisine,
		&meal.Subcategory, &meal.PrepTimeMinutes, &meal.CookTimeMinutes,
		&instructionsJSON, &macrosJSON, &dietaryInfoJSON,
		&meal.CreatedAt, &meal.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("meal not found")
	}
	if err != nil {
		return nil, err
	}

	
	if err := json.Unmarshal(instructionsJSON, &meal.Instructions); err != nil {
		return nil, fmt.Errorf("error parsing instructions: %v", err)
	}
	if err := json.Unmarshal(macrosJSON, &meal.Macros); err != nil {
		return nil, fmt.Errorf("error parsing macros: %v", err)
	}
	if err := json.Unmarshal(dietaryInfoJSON, &meal.DietaryInfo); err != nil {
		return nil, fmt.Errorf("error parsing dietary info: %v", err)
	}

	
	if err := r.loadMealOptions(meal); err != nil {
		return nil, err
	}
	if err := r.loadMealIngredients(meal); err != nil {
		return nil, err
	}

	return meal, nil
}

func (r *Repository) SearchMeals(req *MealSearchRequest) ([]entities.Meal, int, error) {
	
	conditions := []string{"m.is_deleted = false"}
	args := []interface{}{}
	argIndex := 1

	if req.Query != "" {
		conditions = append(conditions, fmt.Sprintf("to_tsvector('english', m.name || ' ' || COALESCE(m.description, '')) @@ plainto_tsquery('english', $%d)", argIndex))
		args = append(args, req.Query)
		argIndex++
	}

	if req.Cuisine != "" {
		conditions = append(conditions, fmt.Sprintf("m.cuisine = $%d", argIndex))
		args = append(args, req.Cuisine)
		argIndex++
	}

	if req.Subcategory != "" {
		conditions = append(conditions, fmt.Sprintf("m.subcategory = $%d", argIndex))
		args = append(args, req.Subcategory)
		argIndex++
	}

	if req.MaxPrepTime > 0 {
		conditions = append(conditions, fmt.Sprintf("m.prep_time_minutes <= $%d", argIndex))
		args = append(args, req.MaxPrepTime)
		argIndex++
	}

	if req.MaxCookTime > 0 {
		conditions = append(conditions, fmt.Sprintf("m.cook_time_minutes <= $%d", argIndex))
		args = append(args, req.MaxCookTime)
		argIndex++
	}

	if len(req.ExcludeAllergens) > 0 {
		for _, allergen := range req.ExcludeAllergens {
			conditions = append(conditions, fmt.Sprintf("NOT (m.dietary_info->'allergens' ? $%d)", argIndex))
			args = append(args, allergen)
			argIndex++
		}
	}

	if len(req.DietaryTags) > 0 {
		for _, tag := range req.DietaryTags {
			conditions = append(conditions, fmt.Sprintf("m.dietary_info->'dietaryTags' ? $%d", argIndex))
			args = append(args, tag)
			argIndex++
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM meal m WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT m.id, m.name, m.description, m.image_url, m.cuisine, m.subcategory,
			   m.prep_time_minutes, m.cook_time_minutes, m.instructions, m.macros, 
			   m.dietary_info, m.created_at, m.updated_at
		FROM meal m
		WHERE %s
		ORDER BY m.name
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var meals []entities.Meal
	for rows.Next() {
		var meal entities.Meal
		var instructionsJSON, macrosJSON, dietaryInfoJSON []byte

		err := rows.Scan(
			&meal.ID, &meal.Name, &meal.Description, &meal.ImageURL, &meal.Cuisine,
			&meal.Subcategory, &meal.PrepTimeMinutes, &meal.CookTimeMinutes,
			&instructionsJSON, &macrosJSON, &dietaryInfoJSON,
			&meal.CreatedAt, &meal.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		
		if err := json.Unmarshal(instructionsJSON, &meal.Instructions); err != nil {
			return nil, 0, fmt.Errorf("error parsing instructions: %v", err)
		}
		if err := json.Unmarshal(macrosJSON, &meal.Macros); err != nil {
			return nil, 0, fmt.Errorf("error parsing macros: %v", err)
		}
		if err := json.Unmarshal(dietaryInfoJSON, &meal.DietaryInfo); err != nil {
			return nil, 0, fmt.Errorf("error parsing dietary info: %v", err)
		}

		
		if err := r.loadMealOptions(&meal); err != nil {
			return nil, 0, err
		}
		if err := r.loadMealIngredients(&meal); err != nil {
			return nil, 0, err
		}

		meals = append(meals, meal)
	}

	return meals, total, nil
}

func (r *Repository) GetCuisines() ([]CuisineInfo, error) {
	query := `
		SELECT cuisine, subcategory, COUNT(*) as meal_count
		FROM meal
		WHERE is_deleted = false
		GROUP BY cuisine, subcategory
		ORDER BY cuisine, subcategory`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cuisineMap := make(map[string]*CuisineInfo)
	for rows.Next() {
		var cuisine, subcategory string
		var mealCount int
		var subcategoryPtr sql.NullString

		err := rows.Scan(&cuisine, &subcategoryPtr, &mealCount)
		if err != nil {
			return nil, err
		}

		if subcategoryPtr.Valid {
			subcategory = subcategoryPtr.String
		}

		if cuisineMap[cuisine] == nil {
			cuisineMap[cuisine] = &CuisineInfo{
				Name:          cuisine,
				Subcategories: []string{},
				MealCount:     0,
			}
		}

		cuisineMap[cuisine].MealCount += mealCount
		if subcategory != "" {
			cuisineMap[cuisine].Subcategories = append(cuisineMap[cuisine].Subcategories, subcategory)
		}
	}

	var cuisines []CuisineInfo
	for _, cuisine := range cuisineMap {
		cuisines = append(cuisines, *cuisine)
	}

	return cuisines, nil
}

func (r *Repository) loadMealOptions(meal *entities.Meal) error {
	query := `
		SELECT id, name, description, is_default
		FROM meal_option
		WHERE meal_id = $1 AND is_deleted = false
		ORDER BY is_default DESC, name`

	rows, err := r.db.Query(query, meal.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	meal.Options = make([]entities.MealOption, 0)
	for rows.Next() {
		var option entities.MealOption
		err := rows.Scan(&option.ID, &option.Name, &option.Description, &option.IsDefault)
		if err != nil {
			return err
		}
		meal.Options = append(meal.Options, option)
	}

	return nil
}

func (r *Repository) loadMealIngredients(meal *entities.Meal) error {
	query := `
		SELECT id, name, quantity, unit, category, ingredient_type
		FROM meal_ingredient
		WHERE meal_id = $1 AND is_deleted = false
		ORDER BY ingredient_type, name`

	rows, err := r.db.Query(query, meal.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	meal.PrimaryIngredients = make([]entities.MealIngredient, 0)
	meal.SecondaryIngredients = make([]entities.MealIngredient, 0)
	meal.Sides = make([]entities.MealIngredient, 0)

	for rows.Next() {
		var ingredient entities.MealIngredient
		var ingredientType string

		err := rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.Quantity,
			&ingredient.Unit, &ingredient.Category, &ingredientType)
		if err != nil {
			return err
		}

		switch ingredientType {
		case "primary":
			meal.PrimaryIngredients = append(meal.PrimaryIngredients, ingredient)
		case "secondary":
			meal.SecondaryIngredients = append(meal.SecondaryIngredients, ingredient)
		case "side":
			meal.Sides = append(meal.Sides, ingredient)
		}
	}

	return nil
}



func (r *Repository) CreateMealPlan(familyID int, createdBy int, req *CreateMealPlanRequest) (*entities.MealPlan, error) {
	query := `
		INSERT INTO meal_plan (family_id, start_date, end_date, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now().UTC()
	var mealPlanID int

	err := r.db.QueryRow(query, familyID, req.StartDate, req.EndDate, createdBy, now, now).Scan(&mealPlanID)
	if err != nil {
		return nil, fmt.Errorf("error creating meal plan: %v", err)
	}

	return r.GetMealPlanByID(mealPlanID, familyID)
}

func (r *Repository) GetMealPlanByID(id int, familyID int) (*entities.MealPlan, error) {
	query := `
		SELECT id, family_id, start_date, end_date, created_at, updated_at, created_by
		FROM meal_plan
		WHERE id = $1 AND family_id = $2 AND is_deleted = false`

	plan := new(entities.MealPlan)
	err := r.db.QueryRow(query, id, familyID).Scan(
		&plan.ID, &plan.FamilyID, &plan.StartDate, &plan.EndDate,
		&plan.CreatedAt, &plan.UpdatedAt, &plan.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("meal plan not found")
	}
	if err != nil {
		return nil, err
	}

	
	if err := r.loadMealPlanDays(plan); err != nil {
		return nil, err
	}

	return plan, nil
}

func (r *Repository) GetMealPlansByFamilyID(familyID int) ([]*entities.MealPlan, error) {
	query := `
		SELECT id, family_id, start_date, end_date, created_at, updated_at, created_by
		FROM meal_plan
		WHERE family_id = $1 AND is_deleted = false
		ORDER BY start_date DESC`

	rows, err := r.db.Query(query, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*entities.MealPlan
	for rows.Next() {
		plan := new(entities.MealPlan)
		err := rows.Scan(
			&plan.ID, &plan.FamilyID, &plan.StartDate, &plan.EndDate,
			&plan.CreatedAt, &plan.UpdatedAt, &plan.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		
		if err := r.loadMealPlanDays(plan); err != nil {
			return nil, err
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

func (r *Repository) UpdateMealPlanDay(mealPlanID int, familyID int, date time.Time, req *UpdateMealPlanDayRequest) error {
	
	checkQuery := `SELECT 1 FROM meal_plan WHERE id = $1 AND family_id = $2 AND is_deleted = false`
	var exists int
	err := r.db.QueryRow(checkQuery, mealPlanID, familyID).Scan(&exists)
	if err == sql.ErrNoRows {
		return fmt.Errorf("meal plan not found")
	}
	if err != nil {
		return err
	}

	
	query := `
		INSERT INTO meal_plan_day (meal_plan_id, date, meal_id, adult_count, child_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (meal_plan_id, date) WHERE is_deleted = false
		DO UPDATE SET
			meal_id = EXCLUDED.meal_id,
			adult_count = EXCLUDED.adult_count,
			child_count = EXCLUDED.child_count,
			updated_at = EXCLUDED.updated_at`

	now := time.Now().UTC()
	_, err = r.db.Exec(query, mealPlanID, date, req.MealID, req.AdultCount, req.ChildCount, now, now)
	if err != nil {
		return fmt.Errorf("error updating meal plan day: %v", err)
	}

	return nil
}

func (r *Repository) DeleteMealPlan(id int, familyID int, deletedBy int) error {
	query := `
		UPDATE meal_plan
		SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
		WHERE id = $1 AND family_id = $2 AND is_deleted = false`

	result, err := r.db.Exec(query, id, familyID, time.Now().UTC(), deletedBy)
	if err != nil {
		return fmt.Errorf("error deleting meal plan: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("meal plan not found")
	}

	return nil
}

func (r *Repository) loadMealPlanDays(plan *entities.MealPlan) error {
	query := `
		SELECT mpd.id, mpd.meal_plan_id, mpd.date, mpd.meal_id, mpd.adult_count, 
			   mpd.child_count, mpd.created_at, mpd.updated_at,
			   m.id, m.name, m.description, m.image_url, m.cuisine, m.subcategory,
			   m.prep_time_minutes, m.cook_time_minutes
		FROM meal_plan_day mpd
		LEFT JOIN meal m ON mpd.meal_id = m.id AND m.is_deleted = false
		WHERE mpd.meal_plan_id = $1 AND mpd.is_deleted = false
		ORDER BY mpd.date`

	rows, err := r.db.Query(query, plan.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	plan.Days = make([]entities.MealPlanDay, 0)
	for rows.Next() {
		var day entities.MealPlanDay
		var mealFields struct {
			ID              sql.NullInt64
			Name            sql.NullString
			Description     sql.NullString
			ImageURL        sql.NullString
			Cuisine         sql.NullString
			Subcategory     sql.NullString
			PrepTimeMinutes sql.NullInt64
			CookTimeMinutes sql.NullInt64
		}

		err := rows.Scan(
			&day.ID, &day.MealPlanID, &day.Date, &day.MealID, &day.AdultCount,
			&day.ChildCount, &day.CreatedAt, &day.UpdatedAt,
			&mealFields.ID, &mealFields.Name, &mealFields.Description,
			&mealFields.ImageURL, &mealFields.Cuisine, &mealFields.Subcategory,
			&mealFields.PrepTimeMinutes, &mealFields.CookTimeMinutes,
		)
		if err != nil {
			return err
		}
		
		if day.MealID != nil && mealFields.ID.Valid {
			day.Meal = &entities.Meal{
				ID:              int(mealFields.ID.Int64),
				Name:            mealFields.Name.String,
				Description:     mealFields.Description.String,
				ImageURL:        mealFields.ImageURL.String,
				Cuisine:         mealFields.Cuisine.String,
				Subcategory:     mealFields.Subcategory.String,
				PrepTimeMinutes: int(mealFields.PrepTimeMinutes.Int64),
				CookTimeMinutes: int(mealFields.CookTimeMinutes.Int64),
			}
		}

		plan.Days = append(plan.Days, day)
	}

	return nil
}



func (r *Repository) CreateShoppingList(familyID int, mealPlanID int, startDate, endDate time.Time) (*entities.ShoppingList, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()
	
	deleteQuery := `
		UPDATE shopping_list 
		SET is_deleted = true, deleted_at = $2, updated_at = $2
		WHERE meal_plan_id = $1 AND is_deleted = false`
	_, err = tx.Exec(deleteQuery, mealPlanID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("error deleting existing shopping list: %v", err)
	}

	
	createQuery := `
		INSERT INTO shopping_list (family_id, meal_plan_id, start_date, end_date, generated_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now().UTC()
	var shoppingListID int
	err = tx.QueryRow(createQuery, familyID, mealPlanID, startDate, endDate, now, now).Scan(&shoppingListID)
	if err != nil {
		return nil, fmt.Errorf("error creating shopping list: %v", err)
	}

	
	if err := r.generateShoppingListItems(tx, shoppingListID, mealPlanID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return r.GetShoppingListByMealPlanID(mealPlanID, familyID)
}

func (r *Repository) GetShoppingListByMealPlanID(mealPlanID int, familyID int) (*entities.ShoppingList, error) {
	query := `
		SELECT id, family_id, meal_plan_id, start_date, end_date, generated_at, updated_at
		FROM shopping_list
		WHERE meal_plan_id = $1 AND family_id = $2 AND is_deleted = false`

	list := new(entities.ShoppingList)
	err := r.db.QueryRow(query, mealPlanID, familyID).Scan(
		&list.ID, &list.FamilyID, &list.MealPlanID, &list.StartDate,
		&list.EndDate, &list.GeneratedAt, &list.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shopping list not found")
	}
	if err != nil {
		return nil, err
	}

	
	if err := r.loadShoppingListItems(list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Repository) UpdateShoppingListItem(itemID int, familyID int, req *UpdateShoppingListItemRequest) error {
	query := `
		UPDATE shopping_list_item
		SET is_completed = $3, updated_at = $4
		WHERE id = $1 AND shopping_list_id IN (
			SELECT id FROM shopping_list WHERE family_id = $2 AND is_deleted = false
		) AND is_deleted = false`

	result, err := r.db.Exec(query, itemID, familyID, req.IsCompleted, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error updating shopping list item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("shopping list item not found")
	}

	return nil
}

func (r *Repository) generateShoppingListItems(tx *sql.Tx, shoppingListID int, mealPlanID int) error {
	query := `
		WITH meal_portions AS (
			SELECT mpd.meal_id, 
				   (mpd.adult_count * 1.0 + mpd.child_count * 0.5) as multiplier
			FROM meal_plan_day mpd
			WHERE mpd.meal_plan_id = $1 AND mpd.meal_id IS NOT NULL AND mpd.is_deleted = false
		),
		aggregated_ingredients AS (
			SELECT mi.name as ingredient_name,
				   SUM(mi.quantity * mp.multiplier) as total_quantity,
				   mi.unit,
				   mi.category
			FROM meal_ingredient mi
			JOIN meal_portions mp ON mi.meal_id = mp.meal_id
			WHERE mi.is_deleted = false
			GROUP BY mi.name, mi.unit, mi.category
		)
		INSERT INTO shopping_list_item (
			shopping_list_id, ingredient_name, total_quantity, unit, category, 
			display_order, created_at, updated_at
		)
		SELECT $2, ingredient_name, total_quantity, unit, category,
			   ROW_NUMBER() OVER (ORDER BY category, ingredient_name),
			   $3, $3
		FROM aggregated_ingredients
		ORDER BY category, ingredient_name`

	now := time.Now().UTC()
	_, err := tx.Exec(query, mealPlanID, shoppingListID, now)
	if err != nil {
		return fmt.Errorf("error generating shopping list items: %v", err)
	}

	return nil
}

func (r *Repository) loadShoppingListItems(list *entities.ShoppingList) error {
	query := `
		SELECT id, ingredient_name, total_quantity, unit, category, is_completed
		FROM shopping_list_item
		WHERE shopping_list_id = $1 AND is_deleted = false
		ORDER BY category, ingredient_name`

	rows, err := r.db.Query(query, list.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	list.Items = make([]entities.ShoppingListItem, 0)
	categoryMap := make(map[string][]entities.ShoppingListItem)

	for rows.Next() {
		var item entities.ShoppingListItem
		err := rows.Scan(&item.ID, &item.IngredientName, &item.TotalQuantity,
			&item.Unit, &item.Category, &item.IsCompleted)
		if err != nil {
			return err
		}

		list.Items = append(list.Items, item)
		categoryMap[item.Category] = append(categoryMap[item.Category], item)
	}

	list.Categories = make([]entities.ShoppingListCategory, 0)
	displayOrder := 0
	for category, items := range categoryMap {
		list.Categories = append(list.Categories, entities.ShoppingListCategory{
			Name:         category,
			DisplayOrder: displayOrder,
			Items:        items,
		})
		displayOrder++
	}

	return nil
}