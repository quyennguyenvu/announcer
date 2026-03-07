package app

import (
	"announcer/config"
	"announcer/internal/adapter"
	"announcer/pkg/logger"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func RunAnnounceBreakfast(cfg *config.BreakfastConfig) {
	// Get today's date in the format used in the CSV (dd/mm/yyyy)
	today := time.Now().Add(-24 * time.Hour).Format("02/01/2006")

	// Fetch the CSV data
	resp, err := http.Get(cfg.BreakfastLink)
	if err != nil {
		logger.Error("Error fetching CSV data: %v", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error closing response body: %v", err)
		}
	}()

	// Parse CSV - read only needed rows to optimize memory
	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true    // Allow bare quotes in fields
	reader.FieldsPerRecord = -1 // Allow variable number of fields per row

	// Skip first 2 rows
	for i := range 2 {
		if _, err := reader.Read(); err != nil {
			logger.Error("Error skipping row %d: %v", i, err)
			return
		}
	}

	// Read dates row (row index 2)
	datesRow, err := reader.Read()
	if err != nil {
		logger.Error("Error reading dates row: %v", err)
		return
	}

	// Read food row (row index 3)
	foodRow, err := reader.Read()
	if err != nil {
		logger.Error("Error reading food row: %v", err)
		return
	}

	// Find the column with today's date
	dateColumn := -1
	for i, date := range datesRow {
		if strings.TrimSpace(date) == today {
			dateColumn = i
			break
		}
	}

	if dateColumn == -1 {
		logger.Error("Today's date (%s) not found in the breakfast menu", today)

		// Show available dates for debugging
		var availableDates []string
		for _, date := range datesRow {
			if strings.Contains(date, "/") && len(strings.TrimSpace(date)) > 0 {
				availableDates = append(availableDates, strings.TrimSpace(date))
			}
		}
		logger.Error("Available dates: %v", availableDates)
		return
	}

	// Get the corresponding food item
	if dateColumn >= len(foodRow) {
		logger.Error("Food data not available for column %d", dateColumn)
		return
	}

	todaysFood := strings.TrimSpace(foodRow[dateColumn])
	if todaysFood == "" {
		logger.Error("No food item specified for today (%s)", today)
		return
	}

	tomorrow := time.Now().Add(24 * time.Hour)
	tomorrowFood := "Hối VNPAY cập nhật thực đơn"
	if tomorrow.Weekday() == time.Saturday {
		tomorrowFood = "Cuối tuần nghỉ ngơi thôi"
	}
	if dateColumn+1 < len(foodRow) {
		tomorrowFood = strings.TrimSpace(foodRow[dateColumn+1])
	}

	// Send announcement to Discord as embed with color sidebar
	title := fmt.Sprintf("🍽️ %s: Tới công ty ăn sáng thôi", today)
	description := fmt.Sprintf(
		"**Hôm nay:** %s\n"+
			"**Ngày mai:** %s\n\n"+
			"Chúc ngon miệng! 😋",
		todaysFood, tomorrowFood,
	)
	color := 16753920 // Orange color for sidebar
	footerText := ""

	if err := adapter.SendDiscordEmbed(cfg.DiscordWebhookURL, title, description, color, footerText); err != nil {
		logger.Error("Error sending Discord embed: %v", err)
	}
}
