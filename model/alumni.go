package model

import (
	"time"

	"github.com/google/uuid"
)

type AlumniProfile struct {
	Base
	AccountID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"account_id"`
	Account    Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	BatchYear  int       `gorm:"index" json:"batch_year"`
	City       string    `json:"city"`
	Occupation string    `json:"occupation"`
	Bio        string    `gorm:"type:text" json:"bio"`
	Status     string    `gorm:"size:20;not null;default:pending;index" json:"status"`
}

type NewsArticle struct {
	Base
	AuthorID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id"`
	Author      Account    `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"-"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Slug        string     `gorm:"size:220;not null;uniqueIndex" json:"slug"`
	Category    string     `gorm:"size:80;not null;index" json:"category"`
	Excerpt     string     `gorm:"type:text" json:"excerpt"`
	Body        string     `gorm:"type:text;not null" json:"body"`
	Status      string     `gorm:"size:20;not null;default:draft;index" json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type AlumniEvent struct {
	Base
	AuthorID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id"`
	Author               Account    `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"-"`
	Title                string     `gorm:"size:200;not null" json:"title"`
	Description          string     `gorm:"type:text" json:"description"`
	Category             string     `gorm:"size:80;not null;index" json:"category"`
	Venue                string     `gorm:"size:200" json:"venue"`
	StartsAt             time.Time  `gorm:"not null;index" json:"starts_at"`
	EndsAt               time.Time  `json:"ends_at"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	Status               string     `gorm:"size:20;not null;default:draft;index" json:"status"`
}

type EventRegistration struct {
	Base
	EventID         uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_event_alumni" json:"event_id"`
	Event           AlumniEvent   `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"-"`
	AlumniProfileID uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_event_alumni" json:"alumni_profile_id"`
	AlumniProfile   AlumniProfile `gorm:"foreignKey:AlumniProfileID;constraint:OnDelete:CASCADE" json:"-"`
	Status          string        `gorm:"size:20;not null;default:registered" json:"status"`
}

type BusinessCareerListing struct {
	Base
	AlumniProfileID uuid.UUID     `gorm:"type:uuid;not null;index" json:"alumni_profile_id"`
	AlumniProfile   AlumniProfile `gorm:"foreignKey:AlumniProfileID;constraint:OnDelete:CASCADE" json:"-"`
	ListingType     string        `gorm:"size:20;not null;index" json:"listing_type"`
	Name            string        `gorm:"size:200;not null" json:"name"`
	Category        string        `gorm:"size:80;not null;index" json:"category"`
	Description     string        `gorm:"type:text;not null" json:"description"`
	ContactURL      string        `gorm:"size:500" json:"contact_url"`
	Status          string        `gorm:"size:20;not null;default:pending;index" json:"status"`
	ReviewedBy      *uuid.UUID    `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer        *Account      `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"-"`
	ReviewedAt      *time.Time    `json:"reviewed_at,omitempty"`
}

type ContactMessage struct {
	Base
	Name    string `gorm:"size:160;not null" json:"name"`
	Email   string `gorm:"size:254;not null;index" json:"email"`
	Message string `gorm:"type:text;not null" json:"message"`
	Status  string `gorm:"size:20;not null;default:new;index" json:"status"`
}
