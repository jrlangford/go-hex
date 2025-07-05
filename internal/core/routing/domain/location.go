package domain

import (
	"go_hex/internal/support/basedomain"
	"go_hex/internal/support/validation"
)

// UnLocode represents a UN/LOCODE (United Nations Code for Trade and Transport Locations)
type UnLocode struct {
	Code string `json:"code" validate:"required,len=5"`
}

// NewUnLocode creates a new UnLocode with validation
func NewUnLocode(code string) (UnLocode, error) {
	unLocode := UnLocode{Code: code}

	if err := validation.Validate(unLocode); err != nil {
		return UnLocode{}, NewDomainValidationError("invalid UN/LOCODE format", err)
	}

	return unLocode, nil
}

// String returns the string representation of the UnLocode
func (u UnLocode) String() string {
	return u.Code
}

// Location represents a physical point in the transport network
type Location struct {
	basedomain.BaseEntity[UnLocode] `json:",inline"`

	// Location data
	Data LocationData `json:"data"`
}

// LocationData represents the value object containing location's business data
type LocationData struct {
	UnLocode UnLocode `json:"unlocode"`
	Name     string   `json:"name" validate:"required,min=1,max=100"`
	Country  string   `json:"country" validate:"required,len=2"` // ISO 3166-1 alpha-2 country code
}

// NewLocation creates a new Location with validation
func NewLocation(code, name, country string) (Location, error) {
	unLocode, err := NewUnLocode(code)
	if err != nil {
		return Location{}, err
	}

	data := LocationData{
		UnLocode: unLocode,
		Name:     name,
		Country:  country,
	}

	if err := validation.Validate(data); err != nil {
		return Location{}, NewDomainValidationError("location data validation failed", err)
	}

	return Location{
		BaseEntity: basedomain.NewBaseEntity(unLocode),
		Data:       data,
	}, nil
}

// GetUnLocode returns the location's UN/LOCODE
func (l Location) GetUnLocode() UnLocode {
	return l.Data.UnLocode
}

// GetName returns the location's name
func (l Location) GetName() string {
	return l.Data.Name
}

// GetCountry returns the location's country code
func (l Location) GetCountry() string {
	return l.Data.Country
}
