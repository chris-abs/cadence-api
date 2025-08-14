package models

import "github.com/google/uuid"

type FamilyID string
type ProfileID string

type MediaID string
type SourceID string
type ChoreID string
type ChoreInstanceID string
type EventID string

type WorkspaceID string
type ContainerID string
type ItemID string
type ItemImageID string
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

func NewSourceID() SourceID {
    return SourceID(uuid.Must(uuid.NewV7()).String())
}

func NewChoreID() ChoreID {
    return ChoreID(uuid.Must(uuid.NewV7()).String())
}

func NewChoreInstanceID() ChoreInstanceID {
    return ChoreInstanceID(uuid.Must(uuid.NewV7()).String())
}

func NewEventID() EventID {
    return EventID(uuid.Must(uuid.NewV7()).String())
}

func NewWorkspaceID() WorkspaceID {
    return WorkspaceID(uuid.Must(uuid.NewV7()).String())
}

func NewContainerID() ContainerID {
    return ContainerID(uuid.Must(uuid.NewV7()).String())
}

func NewItemID() ItemID {
    return ItemID(uuid.Must(uuid.NewV7()).String())
}

func NewItemImageID() ItemImageID {
    return ItemImageID(uuid.Must(uuid.NewV7()).String())
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

func (id MediaID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id SourceID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id ChoreID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id ChoreInstanceID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id EventID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id WorkspaceID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id ContainerID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id ItemID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id ItemImageID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id TagID) IsValid() bool {
    _, err := uuid.Parse(string(id))
    return err == nil
}

func (id FamilyID) String() string {
    return string(id)
}

func (id ProfileID) String() string {
    return string(id)
}

func (id MediaID) String() string {
    return string(id)
}

func (id SourceID) String() string {
    return string(id)
}

func (id ChoreID) String() string {
    return string(id)
}

func (id ChoreInstanceID) String() string {
    return string(id)
}

func (id EventID) String() string {
    return string(id)
}

func (id WorkspaceID) String() string {
    return string(id)
}

func (id ContainerID) String() string {
    return string(id)
}

func (id ItemID) String() string {
    return string(id)
}

func (id ItemImageID) String() string {
    return string(id)
}

func (id TagID) String() string {
    return string(id)
}