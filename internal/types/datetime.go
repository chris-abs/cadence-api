package datetime

import (
	"encoding/json"
	"fmt"
	"time"
)

const ISO8601 = "2006-01-02T15:04:05.000Z"

type DateTime struct {
    time.Time
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
    str := string(data[1:len(data)-1])
    parsed, err := time.Parse(ISO8601, str)
    if err != nil {
        return fmt.Errorf("invalid datetime format, expected ISO8601: %v", err)
    }
    dt.Time = parsed.UTC()
    return nil
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(dt.Time.Format(ISO8601))
}

type NullableDateTime struct {
    Time  time.Time
    Valid bool
}

func (ndt *NullableDateTime) UnmarshalJSON(data []byte) error {
    if string(data) == "null" {
        ndt.Valid = false
        return nil
    }
    
    str := string(data[1:len(data)-1])
    parsed, err := time.Parse(ISO8601, str)
    if err != nil {
        return fmt.Errorf("invalid datetime format, expected ISO8601: %v", err)
    }
    
    ndt.Time = parsed.UTC()
    ndt.Valid = true
    return nil
}

func (ndt NullableDateTime) MarshalJSON() ([]byte, error) {
    if !ndt.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(ndt.Time.Format(ISO8601))
}