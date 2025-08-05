package workspace

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/storage/entities"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(workspace *entities.Workspace) error {
    workspace.ID = models.NewWorkspaceID() 
    
    query := `
        INSERT INTO workspace (id, name, description, profile_id, family_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)` 

    _, err := r.db.Exec(  
        query,
        workspace.ID,
        workspace.Name,
        workspace.Description,
        workspace.ProfileID,
        workspace.FamilyID,
        time.Now().UTC(),
        time.Now().UTC(),
    )

    if err != nil {
        return fmt.Errorf("error creating workspace: %v", err)
    }

    return nil
}

func (r *Repository) GetByID(id models.WorkspaceID, familyID models.FamilyID) (*entities.Workspace, error) {
    workspaceQuery := `
        SELECT w.id, w.name, w.description, w.profile_id, w.family_id, w.created_at, w.updated_at
        FROM workspace w
        WHERE w.id = $1 AND w.family_id = $2 AND w.is_deleted = false`

    workspace := new(entities.Workspace)
    err := r.db.QueryRow(workspaceQuery, id, familyID).Scan(
        &workspace.ID,
        &workspace.Name,
        &workspace.Description,
        &workspace.ProfileID,
        &workspace.FamilyID,
        &workspace.CreatedAt,
        &workspace.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("workspace not found")
    }
    if err != nil {
        return nil, err
    }

    containersQuery := `
        SELECT 
            id, name, description, qr_code, qr_code_image, number, location, 
            profile_id, family_id, workspace_id, created_at, updated_at
        FROM container
        WHERE workspace_id = $1 AND family_id = $2 AND is_deleted = false
        ORDER BY created_at DESC`

    rows, err := r.db.Query(containersQuery, id, familyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    workspace.Containers = make([]entities.Container, 0)
    for rows.Next() {
        var container entities.Container
        var workspaceID sql.NullString
        err := rows.Scan(
            &container.ID,
            &container.Name,
            &container.Description,
            &container.QRCode,
            &container.QRCodeImage,
            &container.Number,
            &container.Location,
            &container.ProfileID, 
            &container.FamilyID,
            &workspaceID,
            &container.CreatedAt,
            &container.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }

        if workspaceID.Valid {
            wsID := models.WorkspaceID(workspaceID.String)
            container.WorkspaceID = &wsID
        }

        workspace.Containers = append(workspace.Containers, container)
    }

    return workspace, nil
}

func (r *Repository) GetByFamilyID(familyID models.FamilyID, profileID models.ProfileID) ([]*entities.Workspace, error) {
    query := `
        SELECT id, name, description, profile_id, family_id, created_at, updated_at 
        FROM workspace
        WHERE family_id = $1 AND is_deleted = false
        ORDER BY created_at DESC`

    rows, err := r.db.Query(query, familyID)
    if err != nil {
        return nil, fmt.Errorf("error querying workspaces: %v", err)
    }
    defer rows.Close()

    var workspaces []*entities.Workspace
    for rows.Next() {
        workspace := new(entities.Workspace)
        err := rows.Scan(
            &workspace.ID,
            &workspace.Name,
            &workspace.Description,
            &workspace.ProfileID,
            &workspace.FamilyID,
            &workspace.CreatedAt,
            &workspace.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning workspace: %v", err)
        }

        containersQuery := `
            SELECT 
                id, name, description, qr_code, qr_code_image, number, location, 
                profile_id, family_id, workspace_id, created_at, updated_at
            FROM container
            WHERE workspace_id = $1 AND family_id = $2 AND is_deleted = false
            ORDER BY created_at DESC`

        containerRows, err := r.db.Query(containersQuery, workspace.ID, familyID)
        if err != nil {
            return nil, fmt.Errorf("error querying containers: %v", err)
        }

        workspace.Containers = make([]entities.Container, 0)
        func() {
            defer containerRows.Close()
            for containerRows.Next() {
                var container entities.Container
                var workspaceID sql.NullString
                err := containerRows.Scan(
                    &container.ID,
                    &container.Name,
                    &container.Description,
                    &container.QRCode,
                    &container.QRCodeImage,
                    &container.Number,
                    &container.Location,
                    &container.ProfileID,
                    &container.FamilyID,
                    &workspaceID,
                    &container.CreatedAt,
                    &container.UpdatedAt,
                )
                if err != nil {
                    return
                }

                if workspaceID.Valid {
                    wsID := models.WorkspaceID(workspaceID.String)
                    container.WorkspaceID = &wsID
                }

                workspace.Containers = append(workspace.Containers, container)
            }
        }()

        workspaces = append(workspaces, workspace)
    }

    return workspaces, nil
}

func (r *Repository) Update(workspace *entities.Workspace) error {
    query := `
        UPDATE workspace
        SET name = $2, description = $3, updated_at = $4
        WHERE id = $1 AND family_id = $5 AND is_deleted = false`

    result, err := r.db.Exec(
        query,
        workspace.ID,
        workspace.Name,
        workspace.Description,
        time.Now().UTC(),
        workspace.FamilyID,
    )
    if err != nil {
        return fmt.Errorf("error updating workspace: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking update result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("workspace not found")
    }

    return nil
}

func (r *Repository) UpdateContainers(workspaceID models.WorkspaceID, familyID models.FamilyID, containerIDs []models.ContainerID) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    if err := r.clearWorkspaceContainers(tx, workspaceID, familyID); err != nil {
        return err
    }

    if err := r.assignContainersToWorkspace(tx, workspaceID, familyID, containerIDs); err != nil {
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("error committing transaction: %v", err)
    }

    return nil
}

func (r *Repository) clearWorkspaceContainers(tx *sql.Tx, workspaceID models.WorkspaceID, familyID models.FamilyID) error {
    query := `
        UPDATE container 
        SET workspace_id = NULL, updated_at = $3
        WHERE workspace_id = $1 AND family_id = $2`

    _, err := tx.Exec(query, workspaceID, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error clearing workspace containers: %v", err)
    }

    return nil
}

func (r *Repository) assignContainersToWorkspace(tx *sql.Tx, workspaceID models.WorkspaceID, familyID models.FamilyID, containerIDs []models.ContainerID) error {
    if len(containerIDs) == 0 {
        return nil
    }

    args := make([]interface{}, len(containerIDs))
    for i, id := range containerIDs {
        args[i] = string(id)
    }

    query := `
        UPDATE container 
        SET workspace_id = $1, updated_at = $2
        WHERE id = ANY($3) AND family_id = $4`

    _, err := tx.Exec(query, workspaceID, time.Now().UTC(), "{" + fmt.Sprintf(`"%s"`, containerIDs) + "}", familyID)
    if err != nil {
        return fmt.Errorf("error assigning containers to workspace: %v", err)
    }

    return nil
}

func (r *Repository) Delete(id models.WorkspaceID, familyID models.FamilyID, deletedBy models.ProfileID) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback()

    orphanQuery := `
        UPDATE container 
        SET workspace_id = NULL, updated_at = $3
        WHERE workspace_id = $1 AND family_id = $2 AND is_deleted = false`
    
    _, err = tx.Exec(orphanQuery, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error orphaning containers: %v", err)
    }
    
    deleteQuery := `
        UPDATE workspace
        SET is_deleted = true, deleted_at = $3, deleted_by = $4, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = false`
    
    result, err := tx.Exec(deleteQuery, id, familyID, time.Now().UTC(), deletedBy)
    if err != nil {
        return fmt.Errorf("error soft deleting workspace: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking delete result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("workspace not found or already deleted")
    }

    return tx.Commit()
}

func (r *Repository) RestoreDeleted(id models.WorkspaceID, familyID models.FamilyID) error {
    query := `
        UPDATE workspace
        SET is_deleted = false, deleted_at = NULL, deleted_by = NULL, updated_at = $3
        WHERE id = $1 AND family_id = $2 AND is_deleted = true`
    
    result, err := r.db.Exec(query, id, familyID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("error restoring workspace: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking restore result: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("workspace not found or not deleted")
    }

    return nil
}