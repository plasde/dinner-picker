package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
    "math/rand"
)

type WeekState struct {
    WeekStart        time.Time `json:"week_start"`
    CurrentWeek      []string  `json:"current_week"`      // dinner names selected this week
    PreviousWeek     []string  `json:"previous_week"`     // dinner names from last week
}

const StateFileName = "dinner_state.json"

// LoadState reads the state file, creating a new one if it doesn't exist
func LoadState() (*WeekState, error) {
    // Check if state file exists
    if _, err := os.Stat(StateFileName); os.IsNotExist(err) {
        // Create new state
        state := &WeekState{
            WeekStart:    GetCurrentWeekStart(),
            CurrentWeek:  []string{},
            PreviousWeek: []string{},
        }
        return state, nil
    }

    // Read existing state
    file, err := os.ReadFile(StateFileName)
    if err != nil {
        return nil, fmt.Errorf("error reading state file: %w", err)
    }

    var state WeekState
    err = json.Unmarshal(file, &state)
    if err != nil {
        return nil, fmt.Errorf("error parsing state JSON: %w", err)
    }

    return &state, nil
}

// SaveState writes the current state to file
func (s *WeekState) SaveState() error {
    data, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return fmt.Errorf("error marshaling state: %w", err)
    }

    err = os.WriteFile(StateFileName, data, 0644)
    if err != nil {
        return fmt.Errorf("error writing state file: %w", err)
    }

    return nil
}

// CheckNewWeek determines if we've moved to a new week and updates state accordingly
func (s *WeekState) CheckNewWeek() {
    currentWeekStart := GetCurrentWeekStart()
    
    // If we're in a new week, rotate the selections
    if !s.WeekStart.Equal(currentWeekStart) {
        s.PreviousWeek = s.CurrentWeek
        s.CurrentWeek = []string{}
        s.WeekStart = currentWeekStart
    }
}

// GetCurrentWeekStart returns the start of the current week (Sunday)
func GetCurrentWeekStart() time.Time {
    now := time.Now()
    
    // Find the most recent Sunday
    daysFromSunday := int(now.Weekday())
    weekStart := now.AddDate(0, 0, -daysFromSunday)
    
    // Set to start of day
    return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
}

// IsAlreadySelected checks if a dinner was selected this week or last week
func (s *WeekState) IsAlreadySelected(dinnerName string) bool {
    // Check current week
    for _, name := range s.CurrentWeek {
        if name == dinnerName {
            return true
        }
    }
    
    // Check previous week
    for _, name := range s.PreviousWeek {
        if name == dinnerName {
            return true
        }
    }
    
    return false
}

// AddSelection adds a dinner to the current week's selections
func (s *WeekState) AddSelection(dinnerName string) {
    s.CurrentWeek = append(s.CurrentWeek, dinnerName)
}


func PickRandomDinner(dinners *DinnerData, categoryName string) Dinner {
    dinnerSlice := dinners.Dinners[categoryName]  // Get the slice directly
    i := rand.Intn(len(dinnerSlice))              // Random index
    return dinnerSlice[i]                         // Return the dinner (not print it)
}



type Dinner struct {
    Name        string   `json:"name"`
    Category    string   `json:"category"`
    Ingredients []string `json:"ingredients"`
}

type DinnerData struct {
    Dinners map[string][]Dinner `json:"dinners"`
}

// LoadDinners reads the JSON file and returns the dinner data
func LoadDinners(filename string) (*DinnerData, error) {
    file, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("error reading file: %w", err)
    }

    var data DinnerData
    err = json.Unmarshal(file, &data)
    if err != nil {
        return nil, fmt.Errorf("error parsing JSON: %w", err)
    }

    return &data, nil
}

// Example usage
func main() {
    dinners, err := LoadDinners("dinners.json")
    if err != nil {
        fmt.Printf("Error loading dinners: %v\n", err)
        return
    }

    // Print available categories
    fmt.Println("Available categories:")
    for category, meals := range dinners.Dinners {
        fmt.Printf("- %s: %d meals\n", category, len(meals))
    }
}
