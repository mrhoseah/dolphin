package main

import (
	"fmt"
	"log"
	"os"

	"dolphin/internal/template"
)

func RunFinTemplateExample() {
	fmt.Println("🐬 Fin Template Example")
	fmt.Println("=======================")

	// Create Fin template engine configuration
	config := &template.Config{
		ViewsPath:    "views",
		CachePath:    "storage/cache/views",
		CacheEnabled: true,
		DebugMode:    true,
		Extensions:   []string{".fin.html"}, // Only .fin.html is supported
	}

	// Initialize Fin template engine
	engine := template.NewFinEngine(config)

	// Register a layout
	engine.RegisterLayout("admin", `
<!DOCTYPE html>
<html>
<head>
    <title>@yield('title')</title>
    <style>
        .dashboard { padding: 20px; }
        .stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin: 20px 0; }
        .stat-card { background: #f5f5f5; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; color: #333; }
        .users-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .users-table th, .users-table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .users-table th { background-color: #f2f2f2; }
        .role-badge { padding: 4px 8px; border-radius: 4px; font-size: 0.8em; }
        .role-admin { background-color: #ff6b6b; color: white; }
        .role-user { background-color: #4ecdc4; color: white; }
        .status-active { color: #28a745; }
        .status-inactive { color: #dc3545; }
    </style>
</head>
<body>
    <div class="dashboard">
        @yield('content')
    </div>
    @yield('scripts')
</body>
</html>`)

	// Sample data for demonstration
	data := map[string]interface{}{
		"Title": "Admin Dashboard",
		"Dashboard": map[string]interface{}{
			"Stats": map[string]interface{}{
				"TotalUsers":     150,
				"ActiveSessions": 23,
				"SystemLoad":     45,
			},
			"Users": []map[string]interface{}{
				{
					"ID":       1,
					"Name":     "John Doe",
					"Email":    "john@example.com",
					"Role":     "admin",
					"IsActive": true,
				},
				{
					"ID":       2,
					"Name":     "Jane Smith",
					"Email":    "jane@example.com",
					"Role":     "user",
					"IsActive": true,
				},
				{
					"ID":       3,
					"Name":     "Bob Wilson",
					"Email":    "bob@example.com",
					"Role":     "user",
					"IsActive": false,
				},
			},
			"RecentActivities": []map[string]interface{}{
				{
					"Type":        "login",
					"Description": "User John Doe logged in",
					"Timestamp":   "2024-01-15 10:30:00",
				},
				{
					"Type":        "create",
					"Description": "New user registered",
					"Timestamp":   "2024-01-15 10:25:00",
				},
				{
					"Type":        "update",
					"Description": "User profile updated",
					"Timestamp":   "2024-01-15 10:20:00",
				},
			},
			"SystemInfo": map[string]interface{}{
				"ServerName":  "dolphin-server-01",
				"Version":     "1.0.0",
				"Uptime":      "15 days, 3 hours",
				"MemoryUsage": "2.1 GB / 8 GB",
			},
		},
		"CurrentUser": map[string]interface{}{
			"Name":        "Admin User",
			"Email":       "admin@example.com",
			"Role":        "admin",
			"LastLoginAt": "2024-01-15 10:30:00",
		},
	}

	// Example Fin template content
	finTemplate := `
@extends('admin')

@section('title')
    Admin Dashboard
@endsection

@section('content')
    <h1>Admin Dashboard</h1>
    
    <!-- User information with controller data -->
    <div class="user-info">
        <h2>Welcome, {{.CurrentUser.Name}}!</h2>
        <p>Email: {{.CurrentUser.Email}}</p>
        <p>Role: {{.CurrentUser.Role}}</p>
        <p>Last Login: {{.CurrentUser.LastLoginAt}}</p>
    </div>
    
    <!-- Statistics cards -->
    <div class="stats-grid">
        <div class="stat-card">
            <h3>Total Users</h3>
            <span class="stat-number">{{.Dashboard.Stats.TotalUsers}}</span>
        </div>
        <div class="stat-card">
            <h3>Active Sessions</h3>
            <span class="stat-number">{{.Dashboard.Stats.ActiveSessions}}</span>
        </div>
        <div class="stat-card">
            <h3>System Load</h3>
            <span class="stat-number">{{.Dashboard.Stats.SystemLoad}}%</span>
        </div>
    </div>
    
    <!-- Recent activities -->
    <div class="recent-activities">
        <h3>Recent Activities</h3>
        @foreach(.Dashboard.RecentActivities as activity)
            <div class="activity-item">
                <div class="activity-icon">
                    @if(activity.Type == "login")
                        🔐
                    @elseif(activity.Type == "logout")
                        🚪
                    @elseif(activity.Type == "create")
                        ➕
                    @elseif(activity.Type == "update")
                        ✏️
                    @elseif(activity.Type == "delete")
                        🗑️
                    @else
                        📝
                    @endif
                </div>
                <div class="activity-content">
                    <p>{{activity.Description}}</p>
                    <small>{{activity.Timestamp}}</small>
                </div>
            </div>
        @endforeach
    </div>
    
    <!-- User management table -->
    <div class="user-management">
        <h3>User Management</h3>
        <table class="users-table">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                @foreach(dashboard.Users as user)
                    <tr>
                        <td>{{user.Name}}</td>
                        <td>{{user.Email}}</td>
                        <td>
                            <span class="role-badge role-{{user.Role}}">
                                {{user.Role}}
                            </span>
                        </td>
                        <td>
                            @if(user.IsActive)
                                <span class="status-active">Active</span>
                            @else
                                <span class="status-inactive">Inactive</span>
                            @endif
                        </td>
                        <td>
                            <a href="/admin/users/{{user.ID}}/edit" class="btn-edit">Edit</a>
                            @if(user.IsActive)
                                <a href="/admin/users/{{user.ID}}/deactivate" class="btn-deactivate">Deactivate</a>
                            @else
                                <a href="/admin/users/{{user.ID}}/activate" class="btn-activate">Activate</a>
                            @endif
                        </td>
                    </tr>
                @endforeach
            </tbody>
        </table>
    </div>
    
    <!-- Nested model example -->
    @if(dashboard.SystemInfo)
        @model('SystemInfo', systemInfo)
        <div class="system-info">
            <h3>System Information</h3>
            <div class="info-grid">
                <div class="info-item">
                    <label>Server:</label>
                    <span>{{systemInfo.ServerName}}</span>
                </div>
                <div class="info-item">
                    <label>Version:</label>
                    <span>{{systemInfo.Version}}</span>
                </div>
                <div class="info-item">
                    <label>Uptime:</label>
                    <span>{{systemInfo.Uptime}}</span>
                </div>
                <div class="info-item">
                    <label>Memory Usage:</label>
                    <span>{{systemInfo.MemoryUsage}}</span>
                </div>
            </div>
        </div>
    @endif
@endsection

@section('scripts')
    <script>
        // Dashboard-specific JavaScript
        console.log('Dashboard loaded for user: {{currentUser.Name}}');
    </script>
@endsection`

	// For demonstration, we'll create a temporary template file
	tempFile := "temp_dashboard.fin.html"
	err := os.WriteFile(tempFile, []byte(finTemplate), 0644)
	if err != nil {
		log.Fatal("Failed to create temp template:", err)
	}
	defer os.Remove(tempFile)

	fmt.Println("Fin template content:")
	fmt.Println(finTemplate)
	fmt.Println("\nWould render to:")
	fmt.Printf("Welcome, %s!\n", data["CurrentUser"].(map[string]interface{})["Name"])
	fmt.Printf("Email: %s\n", data["CurrentUser"].(map[string]interface{})["Email"])
	fmt.Printf("Role: %s\n", data["CurrentUser"].(map[string]interface{})["Role"])
	fmt.Printf("Last Login: %s\n", data["CurrentUser"].(map[string]interface{})["LastLoginAt"])

	dashboard := data["Dashboard"].(map[string]interface{})
	stats := dashboard["Stats"].(map[string]interface{})
	fmt.Printf("Total Users: %v\n", stats["TotalUsers"])
	fmt.Printf("Active Sessions: %v\n", stats["ActiveSessions"])
	fmt.Printf("System Load: %v%%\n", stats["SystemLoad"])

	fmt.Println("\nRecent Activities:")
	activities := dashboard["RecentActivities"].([]map[string]interface{})
	for _, activity := range activities {
		fmt.Printf("- %s: %s (%s)\n", activity["Type"], activity["Description"], activity["Timestamp"])
	}

	fmt.Println("\nUsers:")
	users := dashboard["Users"].([]map[string]interface{})
	for _, user := range users {
		fmt.Printf("- %s (%s) - %s - %v\n", user["Name"], user["Email"], user["Role"], user["IsActive"])
	}

	fmt.Println("\n🎉 Fin Template Example completed successfully!")
	fmt.Println("📚 This demonstrates the Fin template syntax with model annotations, loops, and conditionals.")
}

func main() {
	RunFinTemplateExample()
}
