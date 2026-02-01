# mta-cli

A CLI written in Go for using MTA subway information. In early development.

[![Go Version](https://img.shields.io/badge/go-1.25.5-blue.svg)](https://golang.org/doc/go1.25)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Last Commit](https://img.shields.io/github/last-commit/thosib/mta-cli/main)](https://github.com/thosib/mta-cli/commits/main)

## Features

- **Real-time arrival data** for NYC Subway lines 1, 2, and 3.

## Usage

### Basic Commands

**Show all upcoming arrivals for lines 1, 2, and 3:**

```bash
mta-cli arrivals
```

**Filter by station name:**

```bash
mta-cli arrivals "Times Sq-42 St"
mta-cli arrivals "116 St-Columbia University"
```

**Filter by stop ID:**

```bash
mta-cli arrivals 116N
mta-cli arrivals 127S
```

**Watch mode (auto-refresh every 30 seconds):**

```bash
mta-cli arrivals "Times Sq-42 St" --watch
mta-cli arrivals 116N -w
```

### Output Example

```
STOP_ID    ROUTE    STATION                              ARRIVAL_TIME
--------------------------------------------------------------------------------
116N       1        116 St-Columbia University           2:45 PM
116N       1        116 St-Columbia University           2:52 PM
116N       1        116 St-Columbia University           3:01 PM
116S       1        116 St-Columbia University           2:47 PM
116S       1        116 St-Columbia University           2:59 PM

Total: 5 upcoming arrivals
```

## How It Works

### Data Sources

- **GTFS-Realtime Feed**: `https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs`
  - Lines 1, 2, 3, 4, 5, 6, 7, and 42nd St Shuttle (S)
  - Currently filtering for lines 1, 2, 3 only
- **GTFS Static Data**: Included in `gtfs_subway/` directory
  - Station names, stop IDs, route information

### Architecture

```
mta-cli
├── main.go              # Entry point
├── cmd/
│   ├── root.go         # Cobra root command
│   ├── arrivals.go     # Arrivals command and logic
│   └── stops.go        # GTFS static data parsing
└── gtfs_subway/        # GTFS static reference data
    ├── stops.csv       # Station names and IDs
    └── ...
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [GTFS-Realtime Bindings](https://github.com/MobilityData/gtfs-realtime-bindings)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)

```

**Note**: This tool is not affiliated with or endorsed by the MTA. Real-time data is provided as-is by the MTA API.
```
