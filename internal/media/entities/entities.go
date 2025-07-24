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
	LogoURL  string `json:"logoUrl"`
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
			{ID: 1, Name: "Netflix", LogoURL: "/logos/netflix.png", Category: "streaming"},
			{ID: 2, Name: "Disney+", LogoURL: "/logos/disney.png", Category: "streaming"},
			{ID: 3, Name: "Prime Video", LogoURL: "/logos/prime.png", Category: "streaming"},
			{ID: 4, Name: "HBO Max", LogoURL: "/logos/hbo.png", Category: "streaming"},
			{ID: 5, Name: "Apple TV+", LogoURL: "/logos/apple.png", Category: "streaming"},
			{ID: 6, Name: "Hulu", LogoURL: "/logos/hulu.png", Category: "streaming"},
			{ID: 7, Name: "Paramount+", LogoURL: "/logos/paramount.png", Category: "streaming"},
			{ID: 8, Name: "Peacock", LogoURL: "/logos/peacock.png", Category: "streaming"},
			{ID: 9, Name: "BBC iPlayer", LogoURL: "/logos/bbc.png", Category: "streaming"},
			{ID: 10, Name: "ITV Hub", LogoURL: "/logos/itv.png", Category: "streaming"},
			{ID: 11, Name: "All 4", LogoURL: "/logos/all4.png", Category: "streaming"},
			{ID: 12, Name: "Now TV", LogoURL: "/logos/now.png", Category: "streaming"},
			{ID: 13, Name: "Cinema", LogoURL: "/logos/cinema.png", Category: "cinema"},
			{ID: 14, Name: "DVD/Blu-ray", LogoURL: "/logos/disc.png", Category: "physical"},
			{ID: 15, Name: "Other", LogoURL: "/logos/other.png", Category: "other"},
		}
	)