package main

import (
	"fmt"

	"github.com/golang/geo/s2"
)

type City struct {
	Name string
	Lat  float64
	Lon  float64
}

func main() {
	cities := []City{
		{"Jakarta", -6.2088, 106.8456},
		{"Surabaya", -7.2575, 112.7521},
		{"Bandung", -6.9175, 107.6191},
		{"Medan", 3.5952, 98.6722},
		{"Semarang", -6.9667, 110.4167},
		{"Makassar", -5.1476, 119.4327},
		{"Palembang", -2.9761, 104.7754},
		{"Tangerang", -6.1783, 106.6319},
		{"South Tangerang", -6.2886, 106.7179},
		{"Depok", -6.4025, 106.7942},
		{"Batam", 1.1281, 104.0322},
		{"Bogor", -6.5971, 106.7986},
		{"Padang", -0.9471, 100.4172},
		{"Pekanbaru", 0.5071, 101.4478},
		{"Malang", -7.9839, 112.6214},
		{"Samarinda", -0.4949, 117.1492},
		{"Pontianak", -0.0263, 109.3425},
		{"Banjarmasin", -3.3167, 114.5917},
		{"Denpasar", -8.6705, 115.2126},
		{"Yogyakarta", -7.7956, 110.3695},
		{"Manado", 1.4748, 124.8421},
		{"Ambon", -3.6954, 128.1814},
		{"Jayapura", -2.5337, 140.7181},
	}

	fmt.Println("-- Seed major Indonesian cities as geo-regions")
	fmt.Println("INSERT INTO geo_regions (region_name, s2_cell_id, geometry) VALUES")

	for i, city := range cities {
		latlng := s2.LatLngFromDegrees(city.Lat, city.Lon)
		cellID := s2.CellIDFromLatLng(latlng).Parent(13)

		// Create a small polygon (roughly 1km around the city center)
		// For a real app, these would be city boundaries, but for a seed, a bbox is fine.
		d := 0.01 // half-size in degrees
		polygonWKT := fmt.Sprintf("POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
			city.Lon-d, city.Lat-d,
			city.Lon+d, city.Lat-d,
			city.Lon+d, city.Lat+d,
			city.Lon-d, city.Lat+d,
			city.Lon-d, city.Lat-d,
		)

		comma := ","
		if i == len(cities)-1 {
			comma = ";"
		}

		fmt.Printf("('%s', %d, ST_GeogFromText('%s'))%s\n",
			city.Name, uint64(cellID), polygonWKT, comma)
	}
}
