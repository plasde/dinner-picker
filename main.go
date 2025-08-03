package main

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
    "time"
)

type Dinner struct {
    Name        string   `json:"name"`
    Category    string   `json:"category"`
    Ingredients []string `json:"ingredients"`
}

type DinnerData struct {
    Dinners map[string][]Dinner `json:"dinners"`
}

type WeekState struct {
    WeekStart    time.Time `json:"week_start"`
    CurrentWeek  []Dinner  `json:"current_week"`
    PreviousWeek []Dinner  `json:"previous_week"`
}

const StateFileName = "dinner_state.json"

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

// LoadState reads the state file, creating a new one if it doesn't exist
func LoadState() (*WeekState, error) {
    if _, err := os.Stat(StateFileName); os.IsNotExist(err) {
        state := &WeekState{
            WeekStart:    GetCurrentWeekStart(),
            CurrentWeek:  []Dinner{},
            PreviousWeek: []Dinner{},
        }
        return state, nil
    }

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
    
    if !s.WeekStart.Equal(currentWeekStart) {
        s.PreviousWeek = s.CurrentWeek
        s.CurrentWeek = []Dinner{}
        s.WeekStart = currentWeekStart
    }
}

// GetCurrentWeekStart returns the start of the current week (Sunday)
func GetCurrentWeekStart() time.Time {
    now := time.Now()
    daysFromSunday := int(now.Weekday())
    weekStart := now.AddDate(0, 0, -daysFromSunday)
    return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
}

// IsAlreadySelected checks if a dinner was selected this week or last week
func (s *WeekState) IsAlreadySelected(dinnerName string) bool {
    for _, dinner := range s.CurrentWeek {
        if dinner.Name == dinnerName {
            return true
        }
    }
    for _, dinner := range s.PreviousWeek {
        if dinner.Name == dinnerName {
            return true
        }
    }
    return false
}

// AddSelection adds a dinner to the current week's selections
func (s *WeekState) AddSelection(dinner Dinner) {
    s.CurrentWeek = append(s.CurrentWeek, dinner)
}

// PickRandomDinner selects a random dinner from a category
func PickRandomDinner(dinners *DinnerData, categoryName string) Dinner {
    dinnerSlice := dinners.Dinners[categoryName]
    if len(dinnerSlice) == 0 {
        panic(fmt.Sprintf("No dinners available in category: %s", categoryName))
    }
    i := rand.Intn(len(dinnerSlice))
    return dinnerSlice[i]
}

// pickDinnerFromCategory picks a dinner that hasn't been used recently
func pickDinnerFromCategory(dinners *DinnerData, state *WeekState, category string) Dinner {
    for {
        randomDinner := PickRandomDinner(dinners, category)
        if !state.IsAlreadySelected(randomDinner.Name) {
            return randomDinner
        }
    }
}

// SelectWeeklyDinners picks 5 dinners for the week
func SelectWeeklyDinners(dinners *DinnerData, state *WeekState) map[string]Dinner {
    selections := make(map[string]Dinner)
    
    // Sunday - always soup
    sundayDinner := pickDinnerFromCategory(dinners, state, "soup")
    selections["Sunday"] = sundayDinner
    state.AddSelection(sundayDinner)
    
    // Monday-Thursday - pick from remaining categories
    categories := []string{"noodles-rice", "pasta", "bread-y", "Salad"}
    days := []string{"Monday", "Tuesday", "Wednesday", "Thursday"}
    
    // Shuffle categories for variety
    rand.Shuffle(len(categories), func(i, j int) {
        categories[i], categories[j] = categories[j], categories[i]
    })
    
    for i, day := range days {
        dinner := pickDinnerFromCategory(dinners, state, categories[i])
        selections[day] = dinner
        state.AddSelection(dinner)
    }
    
    return selections
}

// PrintWeeklyMenu prints the selected dinners with ingredients
func PrintWeeklyMenu(selections map[string]Dinner) {
    days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"}
    
    fmt.Printf("=== DINNER PLAN FOR WEEK OF %s ===\n\n", time.Now().Format("January 2, 2006"))
    
    for _, day := range days {
        dinner := selections[day]
        fmt.Printf("%s - %s\n", day, dinner.Name)
        for _, ingredient := range dinner.Ingredients {
            fmt.Printf("  %s\n", ingredient)
        }
        fmt.Println()
    }
}

func main() {
    // Seed random number generator
    rand.Seed(time.Now().UnixNano())
    
    // Load dinner data
    dinners, err := LoadDinners("dinners.json")
    if err != nil {
        fmt.Printf("Error loading dinners: %v\n", err)
        return
    }
    
    // Load state
    state, err := LoadState()
    if err != nil {
        fmt.Printf("Error loading state: %v\n", err)
        return
    }
    
    // Check if it's a new week
    state.CheckNewWeek()
    
    // Select dinners for the week
    selections := SelectWeeklyDinners(dinners, state)
    
    // Save updated state
    err = state.SaveState()
    if err != nil {
        fmt.Printf("Error saving state: %v\n", err)
        return
    }
    
    // Print the menu
    PrintWeeklyMenu(selections)
}
