package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Slug              string     `json:"slug"`
	Type              *string    `json:"type"`
	Style             *string    `json:"style"`
	Status            string     `json:"status"`
	Beds              *int       `json:"beds"`
	Baths             *int       `json:"baths"`
	HalfBaths         *int       `json:"half_baths"`
	MainSF            *int       `json:"main_sf"`
	UpperSF           *int       `json:"upper_sf"`
	LowerSF           *int       `json:"lower_sf"`
	PorchDeckSF       *int       `json:"porch_deck_sf"`
	GarageSF          *int       `json:"garage_sf"`
	GarageApartmentSF *int       `json:"garage_apartment_sf"`
	UnfinishedSF      *int       `json:"unfinished_sf"`
	HeatedSF          *int       `json:"heated_sf"`
	TotalSF           *int       `json:"total_sf"`
	Notes             *string    `json:"notes"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CreatedBy         *uuid.UUID `json:"created_by"`
	UpdatedBy         *uuid.UUID `json:"updated_by"`
	CreatedByUser     *User      `json:"created_by_user,omitempty"`
	UpdatedByUser     *User      `json:"updated_by_user,omitempty"`
}

type PlanListResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Type      *string   `json:"type"`
	Style     *string   `json:"style"`
	Status    string    `json:"status"`
	Beds      *int      `json:"beds"`
	Baths     *int      `json:"baths"`
	HalfBaths *int      `json:"half_baths"`
	HeatedSF  *int      `json:"heated_sf"`
	TotalSF   *int      `json:"total_sf"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlanFilesResponse struct {
	Website   map[string]*File `json:"website"`
	Reference []File           `json:"reference"`
	Technical []File           `json:"technical"`
	ThreeD    []File           `json:"3d"`
	Other     []File           `json:"other"`
}
