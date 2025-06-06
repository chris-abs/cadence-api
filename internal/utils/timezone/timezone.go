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
	
	timeInTZ := time.Date(
		localTime.Year(), localTime.Month(), localTime.Day(),
		localTime.Hour(), localTime.Minute(), localTime.Second(),
		0, loc,
	)
	
	return timeInTZ.UTC(), nil
}

func (c *Converter) ConvertUTCToLocal(utcTime time.Time, timezoneName string) time.Time {
	if timezoneName == "" || timezoneName == "UTC" {
		return utcTime
	}
	
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		return utcTime 
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