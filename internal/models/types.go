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
    return FamilyID(uuid.New().String())
}

func NewProfileID() ProfileID {
    return ProfileID(uuid.New().String())
}

func NewMediaID() MediaID {
    return MediaID(uuid.New().String())
}

func NewChoreID() ChoreID {
    return ChoreID(uuid.New().String())
}

func NewEventID() EventID {
    return EventID(uuid.New().String())
}

func NewContainerID() ContainerID {
    return ContainerID(uuid.New().String())
}

func NewItemID() ItemID {
    return ItemID(uuid.New().String())
}

func NewWorkspaceID() WorkspaceID {
    return WorkspaceID(uuid.New().String())
}

func NewTagID() TagID {
    return TagID(uuid.New().String())
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