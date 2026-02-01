package cmd

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

// Arrival represents a single arrival event
type Arrival struct {
	StopID  string
	RouteID string
	Arrival time.Time
}

// fetchFeed fetches and parses the MTA GTFS-Realtime feed
// Filters for routes 1, 2, and 3
func fetchFeed() ([]Arrival, error) {
	// MTA endpoint for A Division (1, 2, 3, 4, 5, 6, S)
	url := "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs"

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse protobuf
	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(data, feed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	// Extract arrivals for routes 1, 2, 3
	var arrivals []Arrival
	now := time.Now()

	for _, entity := range feed.GetEntity() {
		tripUpdate := entity.GetTripUpdate()
		if tripUpdate == nil {
			continue
		}

		trip := tripUpdate.GetTrip()
		if trip == nil {
			continue
		}

		routeID := trip.GetRouteId()
		// Filter for routes 1, 2, 3
		if routeID != "1" && routeID != "2" && routeID != "3" {
			continue
		}

		// Process stop time updates
		for _, stopTimeUpdate := range tripUpdate.GetStopTimeUpdate() {
			arrivalEvent := stopTimeUpdate.GetArrival()
			if arrivalEvent == nil {
				continue
			}

			arrivalTime := arrivalEvent.GetTime()
			if arrivalTime == 0 {
				continue
			}

			// Convert Unix timestamp to time.Time
			t := time.Unix(arrivalTime, 0)

			// Filter out past arrivals
			if t.Before(now) {
				continue
			}

			stopID := stopTimeUpdate.GetStopId()
			if stopID == "" {
				continue
			}

			arrivals = append(arrivals, Arrival{
				StopID:  stopID,
				RouteID: routeID,
				Arrival: t,
			})
		}
	}

	return arrivals, nil
}

// filterArrivals filters the list of arrivals by station name or stop ID
func filterArrivals(arrivals []Arrival, station string, nameToIDs map[string][]string) []Arrival {
	var filtered []Arrival
	
	// Check if station looks like a stop ID (alphanumeric, possibly with N/S suffix)
	// If it matches a stop ID directly, use it
	// Otherwise, treat it as a station name and lookup associated stop IDs
	
	var targetStopIDs map[string]bool
	
	// First, check if it's a direct stop ID match
	isDirectMatch := false
	for _, arrival := range arrivals {
		if arrival.StopID == station {
			isDirectMatch = true
			break
		}
	}
	
	if isDirectMatch {
		// Direct stop ID match
		targetStopIDs = map[string]bool{station: true}
	} else {
		// Try to find by station name
		stopIDs := nameToIDs[station]
		if len(stopIDs) == 0 {
			// No match found
			return filtered
		}
		targetStopIDs = make(map[string]bool)
		for _, id := range stopIDs {
			targetStopIDs[id] = true
		}
	}
	
	// Filter arrivals
	for _, arrival := range arrivals {
		if targetStopIDs[arrival.StopID] {
			filtered = append(filtered, arrival)
		}
	}
	
	return filtered
}

// displayArrivals displays the arrivals in a formatted table
func displayArrivals(arrivals []Arrival, stopIDToName map[string]string) {
	// Sort by arrival time
	sort.Slice(arrivals, func(i, j int) bool {
		return arrivals[i].Arrival.Before(arrivals[j].Arrival)
	})

	// Display arrivals with station names
	fmt.Printf("%-10s %-8s %-35s %s\n", "STOP_ID", "ROUTE", "STATION", "ARRIVAL_TIME")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, arrival := range arrivals {
		stationName := stopIDToName[arrival.StopID]
		if stationName == "" {
			stationName = "(unknown)"
		}
		fmt.Printf("%-10s %-8s %-35s %s\n",
			arrival.StopID,
			arrival.RouteID,
			stationName,
			arrival.Arrival.Format("3:04 PM"),
		)
	}
	fmt.Printf("\nTotal: %d upcoming arrivals\n", len(arrivals))
}


var watchMode bool

var arrivalsCmd = &cobra.Command{
	Use:   "arrivals [station]",
	Short: "Fetch real-time arrival data for subway lines 1, 2, and 3",
	Long: `Fetches and displays real-time arrival information for NYC Subway lines 1, 2, and 3.
Shows stop IDs and arrival times for upcoming trains.

Optionally filter by station name or stop ID:
  mta-cli arrivals                              # Show all arrivals
  mta-cli arrivals "116 St-Columbia University" # Filter by station name
  mta-cli arrivals 116N                         # Filter by stop ID
  mta-cli arrivals 116N --watch                 # Watch mode: continuous updates`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If watch mode is enabled, require a station argument
		if watchMode && len(args) == 0 {
			fmt.Println("Error: watch mode requires a station name or stop ID")
			fmt.Println("Usage: mta-cli arrivals [station] --watch")
			return
		}

		// Load stop mappings
		stopIDToName, nameToIDs, err := LoadStopMaps("gtfs_subway/stops.csv")
		if err != nil {
			fmt.Printf("Warning: Could not load stop names: %v\n", err)
			fmt.Println("Will display stop IDs only.")
		}

		// Get station filter if provided
		var station string
		if len(args) > 0 {
			station = args[0]
		}

		// Function to fetch, filter, and display arrivals
		fetchAndDisplay := func() {
			// Fetch the feed
			arrivals, err := fetchFeed()
			if err != nil {
				fmt.Printf("Error fetching feed: %v\n", err)
				return
			}

			if len(arrivals) == 0 {
				fmt.Println("No upcoming arrivals found.")
				return
			}

			// Apply filtering if station argument provided
			var filteredArrivals []Arrival
			if station != "" {
				filteredArrivals = filterArrivals(arrivals, station, nameToIDs)
				if len(filteredArrivals) == 0 {
					fmt.Printf("No arrivals found for station: %s\n", station)
					return
				}
			} else {
				filteredArrivals = arrivals
			}

			// Display arrivals
			displayArrivals(filteredArrivals, stopIDToName)
		}

		if watchMode {
			// Watch mode: continuous updates
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			// Clear screen function
			clearScreen := func() {
				fmt.Print("\033[H\033[2J") // ANSI escape codes to clear terminal
			}

			// Initial fetch and display
			fetchAndDisplay()
			fmt.Printf("\nLast updated: %s\n", time.Now().Format("3:04:05 PM"))
			fmt.Println("Watch mode active. Press Ctrl+C to exit.")
			fmt.Println("Refreshing every 30 seconds...")

			// Continuous updates
			for range ticker.C {
				clearScreen()
				fetchAndDisplay()
				fmt.Printf("\nLast updated: %s\n", time.Now().Format("3:04:05 PM"))
				fmt.Println("Watch mode active. Press Ctrl+C to exit.")
				fmt.Println("Refreshing every 30 seconds...")
			}
		} else {
			// One-time fetch and display
			fetchAndDisplay()
		}
	},
}

func init() {
	rootCmd.AddCommand(arrivalsCmd)
	arrivalsCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Watch mode: continuously update arrivals every 30 seconds")
}
