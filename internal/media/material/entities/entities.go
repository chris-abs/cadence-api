package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Media struct {
	ID          models.MediaID    `json:"id"`
	Name        string            `json:"name"`
	Type        MediaType         `json:"type"`
	Genre       string            `json:"genre"`
	ReleaseYear int               `json:"releaseYear"`
	Runtime     int               `json:"runtime"` 
	PosterURL   string            `json:"posterUrl"`
	SourceIDs   []models.SourceID `json:"sourceIds"`
	Sources     []Source          `json:"sources,omitempty"` 
	WatchWith   WatchWith         `json:"watchWith"`
	Status      Status            `json:"status"`
	Priority    Priority          `json:"priority"`
	Notes       string            `json:"notes"`
	ProfileID   models.ProfileID  `json:"profileId"`
	FamilyID    models.FamilyID   `json:"familyId"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeShow  MediaType = "show"
)

type WatchWith string

const (
	WatchWithAlone   WatchWith = "alone"
	WatchWithPartner WatchWith = "partner"
	WatchWithFamily  WatchWith = "family"
)

type Status string

const (
	StatusToWatch        Status = "to_watch"
	StatusInProgress     Status = "in_progress"
	StatusWatching       Status = "watching"
	StatusAwaitingRelease Status = "awaiting_release"
	StatusWatched        Status = "watched"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Source struct {
    ID       models.SourceID `json:"id"`
    Name     string          `json:"name"`
    Color    string          `json:"color"`   
    Category string          `json:"category"`
}

var (
	ValidGenres = []string{
		"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime",
		"Documentary", "Drama", "Family", "Fantasy", "History", "Horror",
		"Music", "Mystery", "Romance", "Sci-Fi", "Sport", "Thriller",
		"War", "Western",
	}
	
// TODO: this is dogshit - we should provide a base set of soruces and allow the user to create any additional options.	
DefaultSources = []Source{
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000001"), Name: "Netflix", Color: "red", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000002"), Name: "Disney+", Color: "blue", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000003"), Name: "Prime Video", Color: "cyan", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000004"), Name: "HBO Max", Color: "purple", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000005"), Name: "Apple TV+", Color: "slate", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000006"), Name: "Hulu", Color: "green", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000007"), Name: "Paramount+", Color: "indigo", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000008"), Name: "Peacock", Color: "violet", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000009"), Name: "BBC iPlayer", Color: "orange", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000010"), Name: "ITV Hub", Color: "yellow", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000011"), Name: "All 4", Color: "pink", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000012"), Name: "Now TV", Color: "sky", Category: "streaming"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000013"), Name: "Cinema", Color: "amber", Category: "cinema"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000014"), Name: "DVD/Blu-ray", Color: "zinc", Category: "physical"},
    {ID: models.SourceID("018f-8e2e-1000-abcd-000000000015"), Name: "Other", Color: "gray", Category: "other"},
}
)