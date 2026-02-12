package edsm

import (
	"fmt"
	"errors"
)

type SystemData struct {
	Name string `json:"name"`
	ID int64 `json:"id"`
	Coords *Coordinates `json:"coords"`
}

type Coordinates struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}


func (e *EDSM) Systems(names []string) ([]SystemData, error) {
	req, err := e.newRequest("GET", "/api-v1/systems")
	if err != nil {
		return []SystemData{}, errors.Join(err, fmt.Errorf("Unable to query systems %v", names))
	}

	q := req.URL.Query()
	q.Set("showId", "1")
	q.Set("showCoordinates", "1")

	for _, name := range names {
		q.Add("systemName[]", name)
	}

	req.URL.RawQuery = q.Encode()

	retval := make([]SystemData, 0, len(names))
	_, err = e.call(req, &retval)
	if err != nil {
		return []SystemData{}, errors.Join(err, fmt.Errorf("Unable to query systems %v", names))
	}

	return retval, nil
}
