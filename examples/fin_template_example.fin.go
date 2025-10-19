<!-- examples/fin_template_example.fin.go -->
@extends('layouts.admin')
@model('Dashboard', dashboard)

@section('title')
    Admin Dashboard
@endsection

@section('content')
    <div class="dashboard">
        <h1>Admin Dashboard</h1>
        
        <!-- User information with model annotation -->
        @model('User', currentUser)
        <div class="user-info">
            <h2>Welcome, {{currentUser.Name}}!</h2>
            <p>Email: {{currentUser.Email}}</p>
            <p>Role: {{currentUser.Role}}</p>
            <p>Last Login: {{currentUser.LastLoginAt}}</p>
        </div>
        
        <!-- Statistics cards -->
        <div class="stats-grid">
            <div class="stat-card">
                <h3>Total Users</h3>
                <span class="stat-number">{{dashboard.Stats.TotalUsers}}</span>
            </div>
            <div class="stat-card">
                <h3>Active Sessions</h3>
                <span class="stat-number">{{dashboard.Stats.ActiveSessions}}</span>
            </div>
            <div class="stat-card">
                <h3>System Load</h3>
                <span class="stat-number">{{dashboard.Stats.SystemLoad}}%</span>
            </div>
        </div>
        
        <!-- Recent activities -->
        <div class="recent-activities">
            <h3>Recent Activities</h3>
            @foreach(dashboard.RecentActivities as activity)
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
    </div>
@endsection

@section('scripts')
    <script>
        // Dashboard-specific JavaScript
        console.log('Dashboard loaded for user: {{currentUser.Name}}');
    </script>
@endsection
