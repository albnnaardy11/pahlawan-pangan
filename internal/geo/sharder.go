package geo

import (
	"github.com/golang/geo/s2"
)

const (
	// S2Level for sharding (Level 13 is ~1.27 km2)
	// Perfect for dense urban restaurant clusters
	DefaultShardLevel = 13
)

// ShardID calculates the S2 Cell ID for a given lat/lon at a specific level
func GetShardID(lat, lon float64) uint64 {
	latlng := s2.LatLngFromDegrees(lat, lon)
	cellID := s2.CellIDFromLatLng(latlng).Parent(DefaultShardLevel)
	return uint64(cellID)
}

// GetNearbyShards returns the cell IDs of neighboring shards
// Useful for geo-queries that cross shard boundaries
func GetNearbyShards(lat, lon float64) []uint64 {
	latlng := s2.LatLngFromDegrees(lat, lon)
	cellID := s2.CellIDFromLatLng(latlng).Parent(DefaultShardLevel)
	
	neighbors := cellID.EdgeNeighbors()
	ids := make([]uint64, len(neighbors)+1)
	ids[0] = uint64(cellID)
	for i, n := range neighbors {
		ids[i+1] = uint64(n)
	}
	return ids
}
