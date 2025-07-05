package timezone

import (
	"fmt"
	"time"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ConvertLocalToUTC(localTime time.Time, timezoneName string) (time.Time, error) {
	if timezoneName == "" || timezoneName == "UTC" {
		return localTime.UTC(), nil
	}
	
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", timezoneName)
	}
	
	if localTime.Location() != time.UTC && localTime.Location().String() != "Local" {
		localTime = localTime.In(loc)
	}
	
	
	timeInTZ := time.Date(
		localTime.Year(), localTime.Month(), localTime.Day(),
		localTime.Hour(), localTime.Minute(), localTime.Second(),
		localTime.Nanosecond(), loc,
	)
	
	return timeInTZ.UTC(), nil
}

func (c *Converter) ConvertUTCToLocal(utcTime time.Time, timezoneName string) time.Time {
	if timezoneName == "" || timezoneName == "UTC" {
		return utcTime.UTC()
	}
	
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		return utcTime.UTC()
	}
	
	return utcTime.In(loc)
}

func (c *Converter) ValidateTimezone(timezone string) error {
	if timezone == "" {
		return nil 
	}
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %s", timezone)
	}
	return nil
}

func (c *Converter) GetUserTimezone(timezone string) string {
	if timezone == "" {
		return "UTC"
	}
	if err := c.ValidateTimezone(timezone); err != nil {
		return "UTC"
	}
	return timezone
}

func (c *Converter) GetDateBoundaryInTimezone(date time.Time, timezoneName string) (time.Time, error) {
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", timezoneName)
	}
	
	startOfDay := time.Date(
		date.Year(), date.Month(), date.Day(),
		0, 0, 0, 0, loc,
	)
	
	return startOfDay, nil
}

func (c *Converter) ConvertDateToUTCBoundary(date time.Time, timezoneName string) (time.Time, error) {
	boundary, err := c.GetDateBoundaryInTimezone(date, timezoneName)
	if err != nil {
		return time.Time{}, err
	}
	
	return boundary.UTC(), nil
}