package app

import (
	"announcer/config"
	"announcer/internal/adapter"
	"announcer/pkg/logger"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

// The runner's clock is UTC, which is still yesterday before 7AM in Vietnam.
// Vietnam has no DST, so a fixed offset needs no tzdata on the runner.
var vnTZ = time.FixedZone("UTC+7", 7*60*60)

func RunAnnounceBreakfast(cfg *config.BreakfastConfig) {
	// Get today's date in the format used in the CSV (dd/mm/yyyy)
	now := time.Now().In(vnTZ)
	today := now.Format("02/01/2006")

	// Fetch the CSV data
	resp, err := http.Get(cfg.BreakfastLink + "/export?format=csv")
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

	// Prepare the announcement message
	title := fmt.Sprintf("🍔 %s: Tới công ty ăn sáng thôi", today)
	color := 16753920 // Orange color for sidebar
	footerText := ""

	// Find the column with today's date
	todaysFood := "Nhịn"
	for i, date := range datesRow {
		if strings.TrimSpace(date) == today {
			todaysFood = strings.TrimSpace(foodRow[i])
			break
		}
	}

	if strings.Contains(strings.ToLower(todaysFood), "nghỉ") {
		logger.Info("Skipping breakfast announcement, today's food is a day off: %s", todaysFood)
		return
	}

	if todaysFood == "Nhịn" {
		description := getNotFoundMessage(cfg)

		if err := adapter.SendDiscordEmbed(cfg.DiscordWebhookURL, title, description, color, footerText); err != nil {
			logger.Error("Error sending Discord embed: %v", err)
		}

		return // Stop here if today's food is not found, no need to check tomorrow's food
	}

	tomorrow := now.Add(24 * time.Hour)
	tomorrowsFood := "Nhịn"
	if tomorrow.Weekday() == time.Saturday {
		tomorrowsFood = "Cuối tuần nghỉ ngơi thôi"
	}

	tmrStr := tomorrow.Format("02/01/2006")
	for i, date := range datesRow {
		if strings.TrimSpace(date) == tmrStr {
			tomorrowsFood = strings.TrimSpace(foodRow[i])
			break
		}
	}

	description := fmt.Sprintf(
		"**Hôm nay:** %s\n"+
			"**Ngày mai:** %s",
		todaysFood, tomorrowsFood,
	)

	blessing, err := getBlessingMessage(cfg)
	switch {
	case err != nil:
		logger.Error("Error getting blessing message: %v", err)
	case blessing == "":
		logger.Warn("All blessings have been sent, announcing with the default one")
		description += "\n\n" + getAllBlessingsSentMessage(cfg)
	default:
		description += "\n\n" + blessing
	}

	// Send announcement to Discord as embed with color sidebar
	if err := adapter.SendDiscordEmbed(cfg.DiscordWebhookURL, title, description, color, footerText); err != nil {
		logger.Error("Error sending Discord embed: %v", err)
		return
	}

	// Mark only after a successful send so a failed post doesn't use up a blessing
	if blessing != "" {
		if err := markBlessingSent(cfg, blessing); err != nil {
			logger.Error("Error marking blessing as sent: %v", err)
		}
	}
}

func getNotFoundMessage(cfg *config.BreakfastConfig) string {
	notFoundMessages := []string{
		"Hmm, có gì đó sai sai... [Cùng kiểm tra lại nhé!](%s) 🤔",
		"Bữa sáng hôm nay đang chơi trốn tìm 👀 [Ai rảnh ngó hộ với nhé!](%s)",
		"Menu hôm nay đi lạc rồi, ai nhặt được giúp với 🥺 [Link đây nè!](%s)",
		"Có vẻ menu chưa cập nhật, ai rảnh ngó hộ với 🙏 [Bấm vào đây nha!](%s)",
		"Bữa sáng hôm nay biến mất bí ẩn 🕵️ [Có ai biết tung tích không?](%s)",
		"Ơ kìa, menu đâu mất rồi nhỉ? 😅 [Cùng đi tìm nào!](%s)",
		"Báo động! Menu hôm nay vẫn còn ngủ quên 😴 [Đánh thức giùm với!](%s)",
	}
	msg := notFoundMessages[time.Now().In(vnTZ).Day()%len(notFoundMessages)]
	return fmt.Sprintf(msg, cfg.BreakfastLink)
}

// getAllBlessingsSentMessage nudges everyone to add more blessings to the sheet.
// It is not a sheet row, so the caller must not mark it sent.
func getAllBlessingsSentMessage(cfg *config.BreakfastConfig) string {
	return fmt.Sprintf(
		"Lời chúc nào cũng đã được gửi đi rồi 🎉 [Cùng góp thêm lời chúc mới cho ngày thêm vui nào!](%s) ✨",
		cfg.BlessingMessageLink,
	)
}

// getBlessingMessage returns a random blessing not yet marked Sent, or "" when
// every blessing has been sent.
func getBlessingMessage(cfg *config.BreakfastConfig) (string, error) {
	resp, err := http.Get(cfg.BlessingMessageLink + "/export?format=csv")
	if err != nil {
		return "", fmt.Errorf("fetching blessings: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error closing response body: %v", err)
		}
	}()

	// A sheet that isn't shared publicly answers with an HTML sign-in page instead of CSV
	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(contentType, "text/csv") {
		return "", fmt.Errorf("unexpected blessings response: status %d, content type %q", resp.StatusCode, contentType)
	}

	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("parsing blessings: %w", err)
	}

	// Columns: blessing, Sent. Skip the header row.
	if len(rows) > 0 {
		rows = rows[1:]
	}

	var unsent []string
	for _, row := range rows {
		blessing := strings.TrimSpace(row[0])
		sent := len(row) > 1 && strings.EqualFold(strings.TrimSpace(row[1]), "TRUE")
		if blessing != "" && !sent {
			unsent = append(unsent, blessing)
		}
	}

	if len(unsent) == 0 {
		return "", nil
	}

	return unsent[rand.IntN(len(unsent))], nil
}

// markBlessingSent ticks Sent for the blessing via the sheet's Apps Script web
// app (scripts/blessing.gs), since the CSV export link is read-only.
func markBlessingSent(cfg *config.BreakfastConfig, blessing string) error {
	payload, err := json.Marshal(map[string]string{"blessing": blessing})
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	resp, err := http.Post(cfg.BlessingUpdateLink, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("sending update: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error closing response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading update response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "ok" {
		return fmt.Errorf("unexpected update response: status %d, body %q", resp.StatusCode, body)
	}

	return nil
}
