package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
)

func LoadStopData(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open stops file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	stopMap := make(map[string]string)

	// 0: stop_id, 1: stop_name
	
	for i, record := range records {
		// Skip header
		if i == 0 {
			continue
		}
		
		if len(record) < 2 {
			continue
		}

		stopID := record[0]
		stopName := record[1]
		stopMap[stopID] = stopName
	}

	return stopMap, nil
}

// LoadStopMaps reads a GTFS stops.txt file and returns both:
// stop_id -> stop_name map
// stop_name -> []stop_id map (for reverse lookup)
func LoadStopMaps(path string) (map[string]string, map[string][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open stops file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	stopMap := make(map[string]string)
	nameToIDs := make(map[string][]string)

	for i, record := range records {
		if i == 0 {
			continue
		}
		
		if len(record) < 2 {
			continue
		}

		stopID := record[0]
		stopName := record[1]
		
		stopMap[stopID] = stopName
		nameToIDs[stopName] = append(nameToIDs[stopName], stopID)
	}

	return stopMap, nameToIDs, nil
}
