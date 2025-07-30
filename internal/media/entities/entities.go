package entities

import "time"

type Media struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        MediaType `json:"type"`
	Genre       string    `json:"genre"`
	ReleaseYear int       `json:"releaseYear"`
	Runtime     int       `json:"runtime"` 
	PosterURL   string    `json:"posterUrl"`
	SourceIDs   []int     `json:"sourceIds"`
	Sources     []Source  `json:"sources,omitempty"` 
	WatchWith   WatchWith `json:"watchWith"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	Notes       string    `json:"notes"`
	ProfileID   int       `json:"profileId"`
	FamilyID    int       `json:"familyId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Color    string `json:"color"`   
    Category string `json:"category"`
}

var (
	ValidGenres = []string{
		"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime",
		"Documentary", "Drama", "Family", "Fantasy", "History", "Horror",
		"Music", "Mystery", "Romance", "Sci-Fi", "Sport", "Thriller",
		"War", "Western",
	}
	
DefaultSources = []Source{
    {ID: 1, Name: "Netflix", Color: "red", Category: "streaming"},
    {ID: 2, Name: "Disney+", Color: "blue", Category: "streaming"},
    {ID: 3, Name: "Prime Video", Color: "cyan", Category: "streaming"},
    {ID: 4, Name: "HBO Max", Color: "purple", Category: "streaming"},
    {ID: 5, Name: "Apple TV+", Color: "slate", Category: "streaming"},
    {ID: 6, Name: "Hulu", Color: "green", Category: "streaming"},
    {ID: 7, Name: "Paramount+", Color: "indigo", Category: "streaming"},
    {ID: 8, Name: "Peacock", Color: "violet", Category: "streaming"},
    {ID: 9, Name: "BBC iPlayer", Color: "orange", Category: "streaming"},
    {ID: 10, Name: "ITV Hub", Color: "yellow", Category: "streaming"},
    {ID: 11, Name: "All 4", Color: "pink", Category: "streaming"},
    {ID: 12, Name: "Now TV", Color: "sky", Category: "streaming"},
    {ID: 13, Name: "Cinema", Color: "amber", Category: "cinema"},
    {ID: 14, Name: "DVD/Blu-ray", Color: "zinc", Category: "physical"},
    {ID: 15, Name: "Other", Color: "gray", Category: "other"},
}
)
