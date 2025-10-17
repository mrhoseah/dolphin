package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	dolphinTime "github.com/mrhoseah/dolphin/internal/time"
)

func main() {
	fmt.Println("🕒 Dolphin Framework - Time Moment Example")
	fmt.Println("==========================================")

	// Demonstrate various time functions
	now := time.Now()
	fmt.Printf("Current time: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Create some example times
	times := []time.Time{
		now.Add(-2 * time.Minute),      // 2 minutes ago
		now.Add(-1 * time.Hour),        // 1 hour ago
		now.Add(-24 * time.Hour),       // 1 day ago
		now.Add(-7 * 24 * time.Hour),   // 1 week ago
		now.Add(-30 * 24 * time.Hour),  // 1 month ago
		now.Add(-365 * 24 * time.Hour), // 1 year ago
		now.Add(2 * time.Minute),       // 2 minutes from now
		now.Add(1 * time.Hour),         // 1 hour from now
	}

	fmt.Println("📅 Time Formatting Examples:")
	fmt.Println("============================")

	for i, t := range times {
		moment := dolphinTime.NewMoment(t)
		fmt.Printf("%d. %s\n", i+1, t.Format("2006-01-02 15:04:05"))
		fmt.Printf("   From Now: %s\n", moment.FromNow())
		fmt.Printf("   Calendar: %s\n", moment.Calendar())
		fmt.Printf("   Humanize: %s\n", moment.Humanize())
		fmt.Printf("   Relative: %s\n", moment.RelativeTime())
		fmt.Println()
	}

	// Demonstrate time comparisons
	fmt.Println("🔍 Time Comparison Examples:")
	fmt.Println("============================")

	pastTime := now.Add(-2 * time.Hour)
	futureTime := now.Add(2 * time.Hour)

	pastMoment := dolphinTime.NewMoment(pastTime)
	futureMoment := dolphinTime.NewMoment(futureTime)
	nowMoment := dolphinTime.Now()

	fmt.Printf("Past time: %s\n", pastTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Future time: %s\n", futureTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Now: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Printf("Is past time before now? %t\n", pastMoment.IsBefore(nowMoment))
	fmt.Printf("Is future time after now? %t\n", futureMoment.IsAfter(nowMoment))
	fmt.Printf("Is past time today? %t\n", pastMoment.IsToday())
	fmt.Printf("Is future time today? %t\n", futureMoment.IsToday())
	fmt.Println()

	// Demonstrate time manipulation
	fmt.Println("⚙️ Time Manipulation Examples:")
	fmt.Println("==============================")

	baseTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	baseMoment := dolphinTime.NewMoment(baseTime)

	fmt.Printf("Base time: %s\n", baseTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Start of day: %s\n", baseMoment.StartOfDay().Format("2006-01-02 15:04:05"))
	fmt.Printf("End of day: %s\n", baseMoment.EndOfDay().Format("2006-01-02 15:04:05"))
	fmt.Printf("Start of week: %s\n", baseMoment.StartOfWeek().Format("2006-01-02 15:04:05"))
	fmt.Printf("End of week: %s\n", baseMoment.EndOfWeek().Format("2006-01-02 15:04:05"))
	fmt.Printf("Start of month: %s\n", baseMoment.StartOfMonth().Format("2006-01-02 15:04:05"))
	fmt.Printf("End of month: %s\n", baseMoment.EndOfMonth().Format("2006-01-02 15:04:05"))
	fmt.Printf("Start of year: %s\n", baseMoment.StartOfYear().Format("2006-01-02 15:04:05"))
	fmt.Printf("End of year: %s\n", baseMoment.EndOfYear().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Demonstrate duration formatting
	fmt.Println("⏱️ Duration Formatting Examples:")
	fmt.Println("================================")

	durations := []time.Duration{
		30 * time.Second,
		5 * time.Minute,
		2 * time.Hour,
		3 * 24 * time.Hour,
		2 * 7 * 24 * time.Hour,
		6 * 30 * 24 * time.Hour,
		2 * 365 * 24 * time.Hour,
	}

	for _, d := range durations {
		fmt.Printf("Duration: %v\n", d)
		fmt.Printf("  Humanized: %s\n", dolphinTime.FormatDuration(d))
		fmt.Printf("  Ago: %s\n", dolphinTime.FormatDurationAgo(d))
		fmt.Printf("  In: %s\n", dolphinTime.FormatDurationIn(d))
		fmt.Println()
	}

	// Create a simple HTTP server to demonstrate template usage
	fmt.Println("🌐 Starting HTTP server to demonstrate template usage...")
	fmt.Println("Open http://localhost:8082 to see the time helpers in action")
	fmt.Println()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some data with timestamps
		activities := []struct {
			Action    string
			Timestamp time.Time
		}{
			{"User logged in", now.Add(-2 * time.Minute)},
			{"New post created", now.Add(-15 * time.Minute)},
			{"Comment added", now.Add(-1 * time.Hour)},
			{"Profile updated", now.Add(-3 * time.Hour)},
			{"Password changed", now.Add(-1 * 24 * time.Hour)},
			{"Account created", now.Add(-7 * 24 * time.Hour)},
		}

		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dolphin Time Moment Example</title>
    <style>
        body { 
            font-family: system-ui, -apple-system, sans-serif; 
            background: #f6f7fb; 
            color: #111827;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 1px solid #e5e7eb;
        }
        .activity {
            display: flex;
            align-items: center;
            padding: 15px;
            background: #f8fafc;
            border-radius: 8px;
            margin-bottom: 10px;
        }
        .activity-icon {
            width: 10px;
            height: 10px;
            background: #3b82f6;
            border-radius: 50%;
            margin-right: 15px;
        }
        .activity-content {
            flex: 1;
        }
        .activity-action {
            font-weight: 600;
            color: #374151;
            margin-bottom: 4px;
        }
        .activity-time {
            color: #6b7280;
            font-size: 14px;
        }
        .time-examples {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        .time-card {
            background: #f8fafc;
            padding: 15px;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
        }
        .time-card-title {
            font-weight: 600;
            color: #374151;
            margin-bottom: 8px;
        }
        .time-card-value {
            color: #6b7280;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🕒 Dolphin Time Moment Example</h1>
            <p>Demonstrating human-readable time formatting similar to moment.js</p>
        </div>

        <h2>Time Display Examples</h2>
        <div class="time-examples">
            <div class="time-card">
                <div class="time-card-title">Current Time</div>
                <div class="time-card-value">` + now.Format("2006-01-02 15:04:05") + `</div>
            </div>
            <div class="time-card">
                <div class="time-card-title">From Now</div>
                <div class="time-card-value">just now</div>
            </div>
            <div class="time-card">
                <div class="time-card-title">Calendar</div>
                <div class="time-card-value">Today at ` + now.Format("15:04") + `</div>
            </div>
        </div>

        <h2>Recent Activity</h2>`

		for _, activity := range activities {
			moment := dolphinTime.NewMoment(activity.Timestamp)
			html += fmt.Sprintf(`
        <div class="activity">
            <div class="activity-icon"></div>
            <div class="activity-content">
                <div class="activity-action">%s</div>
                <div class="activity-time">%s</div>
            </div>
        </div>`, activity.Action, moment.FromNow())
		}

		html += `
    </div>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Start server
	port := "8082"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	fmt.Printf("🚀 Starting server on port %s\n", port)
	fmt.Printf("📱 Open http://localhost:%s to see the time helpers\n", port)
	fmt.Println("⏹️  Press Ctrl+C to stop the server")
	fmt.Println()

	// Start server in goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for a moment to show the server is running
	time.Sleep(2 * time.Second)

	fmt.Println("✅ Server is running!")
	fmt.Println("✅ Time moment helpers are working!")
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("  • Human-readable time formatting")
	fmt.Println("  • Relative time display (e.g., '2 minutes ago')")
	fmt.Println("  • Calendar-style formatting")
	fmt.Println("  • Time manipulation and comparison")
	fmt.Println("  • Duration formatting")
	fmt.Println("  • Template helper functions")
}
