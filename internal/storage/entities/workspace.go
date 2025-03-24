package entities

import "time"

type Workspace struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    ProfileID   int         `json:"profileId"`
    FamilyID    int         `json:"familyId"`
    Containers  []Container `json:"containers"`
    CreatedAt   time.Time   `json:"createdAt"`
    UpdatedAt   time.Time   `json:"updatedAt"`
}