package models

import (
	. "image"
	"time"

	"github.com/nu7hatch/gouuid"
)

// Trigger triggers for the world
type Trigger struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SpeedLimit
type SpeedLimit struct {
	Location string `json:"location"`
	Value    int    `json:"value"`
}

// Settings settings for the world
type Settings struct {
	ID                      *uuid.UUID   `json:"id"`
	ViolentCrimeRate        float32      `json:"violentCrimeRate"`
	MurderRate              float32      `json:"murderRate"`
	CarAccidentFatalityRate float32      `json:"carAccidentFatalityRate"`
	Diseases                []Disease    `json:"diseases"`
	WorldSpeed              int          `json:"worldSpeed"`
	LastTime                time.Time    `json:"lastTime"`
	Triggers                []Trigger    `json:"triggers"`
	SpeedLimits             []SpeedLimit `json:"speedLimits"`
}

// WorldQueueMessage Messages sent to World Queue
type WorldQueueMessage struct {
	Controller string `json:"controller"`
	Status     string `json:"status"`
	Detail     string `json:"detail"`
}

// WorldTrafficQueueMessage Messages sent to World Traffic Queue
type WorldTrafficQueueMessage struct {
	WorldSettings Settings `json:"worldSettings"`
}

// WorldCityQueueMessage Messages sent to World City Queue
type WorldCityQueueMessage struct {
	WorldSettings Settings `json:"worldSettings"`
	City          string   `json:"city"`
}

type CityWorkerQueueMessage struct {
	WorldSettings Settings   `json:"worldSettings"`
	BuildingID    *uuid.UUID `json:"buildingid"`
}

type TrafficWorkerQueueMessage struct {
	WorldSettings Settings   `json:"worldSettings"`
	PersonID      *uuid.UUID `json:"personid"`
}

type ControllerType int

const (
	TrafficController ControllerType = iota + 1
	CityController
)

// Controller a controller
type Controller struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Ready bool   `json:"ready"`
	Exit  bool   `json:"exit"`
}

type Worker struct {
	ID string `json:"id"`
}

// City city
type City struct {
	ID          *uuid.UUID `json:"id"`
	Name        string     `json:"name"`
	TopLeft     Point      `json:"topleft"`
	BottomRight Point      `json:"bottomright"`
	Established time.Time  `json:"established"`
}

// BuildingType used to identify the type of building
type BuildingType int

const (
	House BuildingType = iota + 1
	Apartment
	School
	Office
	Warehouse
	Retail
	Entertainment
	Hospital
	Police
)

// Building a building in a city
type Building struct {
	ID           *uuid.UUID   `json:"id"`
	Name         string       `json:"name"`
	TopLeft      Point        `json:"topleft"`
	BottomRight  Point        `json:"bottomright"`
	BuildDate    time.Time    `json:"builddate"`
	Type         BuildingType `json:"type"`
	Floors       int          `json:"floors"`
	MaxOccupancy int          `json:"maxoccupancy"`
	CityID       *uuid.UUID   `json:"cityid"`
}

// DeathType used to identify how person died
type DeathType int

const (
	Natural DeathType = iota + 1
	Accident
	Murder
	Illness
)

// Person a person
type Person struct {
	ID              *uuid.UUID   `json:"id"`
	Birthdate       time.Time    `json:"birthdate"`
	FirstName       string       `json:"firstname"`
	LastName        string       `json:"lastname"`
	ChildrenIDs     []*uuid.UUID `json:"childrenIDs"`
	CurrentBuilding *uuid.UUID   `json:"currentbuilding"`
	CurrentXY       Point        `json:"currentxy"`
	Traveling       bool         `json:"traveling"`
	NewToBuilding   bool         `json:"newtobuilding"`
	HomeBuilding    *uuid.UUID   `json:"homebuilding"`
	WorkBuilding    *uuid.UUID   `json:"workbuilding"`
	Health          int          `json:"health"`
	Illness         *uuid.UUID   `json:"illness"`
	Happiness       int          `json:"happiness"`
	DeathDate       time.Time    `json:"deathdate"`
	CauseOfDeath    DeathType    `json:"causeofdeath"`
	Spouse          *uuid.UUID   `json:"spouse"`
}

// Disease types of diseases
type Disease struct {
	Name            string  `json:"name"`
	DaysDetected    int     `json:"daysDetected"`
	AvgDaysIll      int     `json:"avgDaysIll"`
	LethalityRate   float32 `json:"lethalityRate"`
	Infectious      bool    `json:"infectious"`
	InfectionChance float32 `json:"infectionChance"`
	Severity        float32 `json:"severity"`
}
