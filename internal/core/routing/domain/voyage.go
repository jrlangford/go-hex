package domain

import (
	"go_hex/internal/support/basedomain"
	"go_hex/internal/support/validation"
	"time"
)

// CarrierMovement represents a single movement in a voyage schedule
type CarrierMovement struct {
	DepartureLocation UnLocode  `json:"departure_location"`
	ArrivalLocation   UnLocode  `json:"arrival_location"`
	DepartureTime     time.Time `json:"departure_time" validate:"required"`
	ArrivalTime       time.Time `json:"arrival_time" validate:"required"`
}

// NewCarrierMovement creates a new CarrierMovement with validation
func NewCarrierMovement(depLoc, arrLoc UnLocode, depTime, arrTime time.Time) (CarrierMovement, error) {
	movement := CarrierMovement{
		DepartureLocation: depLoc,
		ArrivalLocation:   arrLoc,
		DepartureTime:     depTime,
		ArrivalTime:       arrTime,
	}
	// Validate business rules
	if depLoc == arrLoc {
		return CarrierMovement{}, NewDomainValidationError("departure and arrival locations cannot be the same", nil)
	}
	if !arrTime.After(depTime) {
		return CarrierMovement{}, NewDomainValidationError("arrival time must be after departure time", nil)
	}
	if err := validation.Validate(movement); err != nil {
		return CarrierMovement{}, NewDomainValidationError("carrier movement validation failed", err)
	}

	return movement, nil
}

// Schedule represents the set of planned carrier movements for a voyage
type Schedule struct {
	Movements []CarrierMovement `json:"movements" validate:"required,min=1,dive"`
}

// NewSchedule creates a new Schedule with validation
func NewSchedule(movements []CarrierMovement) (Schedule, error) {
	if len(movements) == 0 {
		return Schedule{}, NewDomainValidationError("schedule must contain at least one movement", nil)
	}

	schedule := Schedule{Movements: movements}

	// Validate movement connectivity - each movement must connect to the next
	for i := 0; i < len(movements)-1; i++ {
		currentMovement := movements[i]
		nextMovement := movements[i+1]

		if currentMovement.ArrivalLocation != nextMovement.DepartureLocation {
			return Schedule{}, NewDomainValidationError("movements must be connected - arrival location must match next movement's departure location", nil)
		}

		if !nextMovement.DepartureTime.After(currentMovement.ArrivalTime) {
			return Schedule{}, NewDomainValidationError("insufficient time between movements", nil)
		}
	}

	if err := validation.Validate(schedule); err != nil {
		return Schedule{}, NewDomainValidationError("schedule validation failed", err)
	}

	return schedule, nil
}

// InitialDepartureLocation returns the departure location of the first movement
func (s Schedule) InitialDepartureLocation() UnLocode {
	if len(s.Movements) == 0 {
		return UnLocode{}
	}
	return s.Movements[0].DepartureLocation
}

// FinalArrivalLocation returns the arrival location of the last movement
func (s Schedule) FinalArrivalLocation() UnLocode {
	if len(s.Movements) == 0 {
		return UnLocode{}
	}
	return s.Movements[len(s.Movements)-1].ArrivalLocation
}

// InitialDepartureTime returns the departure time of the first movement
func (s Schedule) InitialDepartureTime() time.Time {
	if len(s.Movements) == 0 {
		return time.Time{}
	}
	return s.Movements[0].DepartureTime
}

// FinalArrivalTime returns the arrival time of the last movement
func (s Schedule) FinalArrivalTime() time.Time {
	if len(s.Movements) == 0 {
		return time.Time{}
	}
	return s.Movements[len(s.Movements)-1].ArrivalTime
}

// Voyage represents a carrier movement from one port to another on a given schedule
type Voyage struct {
	basedomain.BaseEntity[VoyageNumber] `json:",inline"`

	// Voyage data
	Data VoyageData `json:"data"`
}

// VoyageData represents the value object containing voyage's business data
type VoyageData struct {
	Schedule Schedule `json:"schedule"`
}

// NewVoyage creates a new Voyage with validation
func NewVoyage(movements []CarrierMovement) (Voyage, error) {
	voyageNumber := NewVoyageNumber()

	schedule, err := NewSchedule(movements)
	if err != nil {
		return Voyage{}, err
	}

	data := VoyageData{
		Schedule: schedule,
	}

	if err := validation.Validate(data); err != nil {
		return Voyage{}, NewDomainValidationError("voyage data validation failed", err)
	}

	return Voyage{
		BaseEntity: basedomain.NewBaseEntity(voyageNumber),
		Data:       data,
	}, nil
}

// GetVoyageNumber returns the voyage's number
func (v Voyage) GetVoyageNumber() VoyageNumber {
	return v.Id
}

// GetSchedule returns the voyage's schedule
func (v Voyage) GetSchedule() Schedule {
	return v.Data.Schedule
}
