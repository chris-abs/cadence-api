package models

import "github.com/google/uuid"

type FamilyID string
type ProfileID string

type MediaID string
type ChoreID string
type EventID string
type ContainerID string
type ItemID string
type WorkspaceID string
type TagID string

func NewFamilyID() FamilyID {
    return FamilyID(uuid.Must(uuid.NewV7()).String())
}

func NewProfileID() ProfileID {
    return ProfileID(uuid.Must(uuid.NewV7()).String())
}

func NewMediaID() MediaID {
    return MediaID(uuid.Must(uuid.NewV7()).String())
}

func NewChoreID() ChoreID {
    return ChoreID(uuid.Must(uuid.NewV7()).String())
}

func NewEventID() EventID {
    return EventID(uuid.Must(uuid.NewV7()).String())
}

func NewContainerID() ContainerID {
    return ContainerID(uuid.Must(uuid.NewV7()).String())
}

func NewItemID() ItemID {
    return ItemID(uuid.Must(uuid.NewV7()).String())
}

func NewWorkspaceID() WorkspaceID {
    return WorkspaceID(uuid.Must(uuid.NewV7()).String())
}

func NewTagID() TagID {
    return TagID(uuid.Must(uuid.NewV7()).String())
}

func (id FamilyID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id ProfileID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id FamilyID) String() string {
    return string(id)
}

func (id ProfileID) String() string {
    return string(id)
}