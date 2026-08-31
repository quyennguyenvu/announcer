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
	now := time.Now()
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

	// Send announcement to Discord as embed with color sidebar
	blessing := getBlessingMessage()

	description := fmt.Sprintf(
		"**Hôm nay:** %s\n"+
			"**Ngày mai:** %s\n\n"+
			"%s",
		todaysFood, tomorrowsFood, blessing,
	)

	if err := adapter.SendDiscordEmbed(cfg.DiscordWebhookURL, title, description, color, footerText); err != nil {
		logger.Error("Error sending Discord embed: %v", err)
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
	msg := notFoundMessages[time.Now().Day()%len(notFoundMessages)]
	return fmt.Sprintf(msg, cfg.BreakfastLink)
}

func getBlessingMessage() string {
	blessings := []string{
		// Hài hước
		"Ăn sáng đi, không thì cái bụng réo nguyên buổi họp đó nha! 😂",
		"Đừng ăn quá no nhé, lát họp ngủ gật là sếp biết liền đó! 😴",
		"Cấm vừa ăn vừa code! ... ờ mà thôi, kệ, miễn không rớt phím là được. 🤷",
		"Nhớ nhai 30 lần... à thôi, 5 lần cũng được, kẻo nguội mất ngon! 🍽️",
		"Bữa sáng miễn phí, nhưng deadline thì không - tỉnh táo lên anh em! ⏰",
		// Bề trên / uy quyền
		"Đứng dậy! Ăn sáng! Cày cuốc! Đó là kỷ luật của kẻ thành công! 👑",
		"Im lặng. Ăn. Cày. Đừng than. Đó là công thức! 🎖️",
		"Ăn cho hết, không bỏ thừa - đời này không nuôi kẻ phí của! 🦁",
		"Toàn đội tập hợp! Nạp năng lượng, ra trận, không lùi bước! 🪖",
		"Đàn ông là phải ăn no, làm việc khỏe, không kêu mệt! 💪",
		// Thơ mộng
		"Sáng nay gió nhẹ, nắng hiền, bữa ăn ấm áp đón ngày mới 🌷",
		"Một bữa sáng, một nụ cười, đủ để cả ngày dài thêm thương 💕",
		"Một sớm bình yên, một bữa đầy đủ, lòng người cũng nhẹ tênh 🌸",
		"Hôm nay trời đẹp, đồ ăn ngon, mong lòng người cũng vui 🍃",
		// Gen-Z
		"Bữa sáng slay quá, vibe hôm nay chắc chắn lên top trending! 💅",
		"Ăn sáng xong là vibe của em phất lên ngay, không flop nổi! ✨",
		"Bữa sáng tuyệt cú mèo, hôm nay chắc chắn không có ngày tệ! 🔥",
		// Hiền lành / mẹ hiền
		"Ăn no nhé các con, mẹ thương! 🤱",
		"Nhớ uống nước nữa nha, đừng có khô cổ cả ngày! 💧",
		"Ăn từ từ thôi, không ai giành đâu! 😊",
		// Hiền triết / cụ ông
		"Cổ nhân dạy: bụng no thì đầu mới sáng. 📜",
		"Đói thì cáu, no thì khôn. Cả nhà nên chọn khôn. 🦉",
		"Người ăn sáng đầy đủ là người làm chủ vận mệnh mình. 🧙",
		// Mọt công nghệ
		"Bữa sáng load thành công, hệ thống sẵn sàng deploy bản thân! 🔋",
		"Bữa sáng = git pull năng lượng cho cả ngày, đừng quên commit! 🖥️",
		"Chúc mọi người ăn xong là tỉnh táo, code không sai một dòng! 🧠",
		// Châm biếm / kiêu
		"Bữa sáng đỉnh thế này, không xuất sắc cả ngày mới lạ! 🏆",
		"Ăn sáng xong rồi thì... đến lượt deadline ăn các bạn nha! 😈",
		// Ấm áp / đoàn kết
		"Cùng nhau ăn sáng, cùng nhau cố gắng, cùng nhau về sớm! 🤝",
		"Chúc cả nhà một ngày bình an, may mắn và nhiều niềm vui! 🌈",
	}
	return blessings[time.Now().Day()%len(blessings)] // Add a random blessing based on the day of the month
}
