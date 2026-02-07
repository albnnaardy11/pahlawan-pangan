package geo

import (
	"github.com/golang/geo/s2"
)

// S2Engine implements Google's S2 Geometry for hyper-scale geographic indexing.
// Standard at Google, Uber, and Gojek.
type S2Engine struct {
	Level int // S2 Cell Level (Level 13 is ~1km2, Level 15 is ~200m2)
}

func NewS2Engine() *S2Engine {
	return &S2Engine{
		Level: 15, // High precision for urban food rescue
	}
}

// GetCellID converts latitude and longitude to a unique 64-bit S2 Cell ID.
func (e *S2Engine) GetCellID(lat, lon float64) uint64 {
	latlng := s2.LatLngFromDegrees(lat, lon)
	cellID := s2.CellIDFromLatLng(latlng).Parent(e.Level)
	return uint64(cellID)
}

// GetNearbyCells returns a list of cell IDs surrounding a location.
// Used for "Number Range" database queries instead of heavy math.
func (e *S2Engine) GetNearbyCells(lat, lon float64) []uint64 {
	latlng := s2.LatLngFromDegrees(lat, lon)
	centerCell := s2.CellIDFromLatLng(latlng).Parent(e.Level)
	
	neighbors := centerCell.AllNeighbors(e.Level)
	ids := make([]uint64, len(neighbors)+1)
	ids[0] = uint64(centerCell)
	for i, neighbor := range neighbors {
		ids[i+1] = uint64(neighbor)
	}
	return ids
}
