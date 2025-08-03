package models

import "github.com/google/uuid"

type FamilyID    string
type ProfileID   string

type MediaID     string
type ChoreID     string
type EventID     string
type ContainerID string
type ItemID      string
type WorkspaceID string
type TagID       string

func NewFamilyID() FamilyID {
    return FamilyID(uuid.New().String())
}

func NewProfileID() ProfileID {
    return ProfileID(uuid.New().String())
}
