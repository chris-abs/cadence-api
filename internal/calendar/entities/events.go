package entities

import "time"

type Event struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Start       time.Time `json:"start"`
    End         time.Time `json:"end"`
    AllDay      bool      `json:"allDay"`
    
    CreatedBy   int       `json:"createdBy"` 
    AssigneeIDs []int     `json:"assigneeIds"`
    Color       *string   `json:"color,omitempty"`

    Type        string    `json:"type"` 
    ModuleID    *string   `json:"moduleId,omitempty"`
    EntityID    *int      `json:"entityId,omitempty"`
    
    FamilyID    int       `json:"familyId"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    
    IsDeleted   bool      `json:"isDeleted"`
    DeletedAt   *time.Time `json:"deletedAt,omitempty"`
    DeletedBy   *int      `json:"deletedBy,omitempty"`
}