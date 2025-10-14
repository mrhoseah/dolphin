package frontend

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

// FrontendManager handles frontend integration
type FrontendManager struct {
	templates *template.Template
	assets    embed.FS
}

// NewFrontendManager creates a new frontend manager
func NewFrontendManager(assets embed.FS) *FrontendManager {
	templates := template.New("")

	return &FrontendManager{
		templates: templates,
		assets:    assets,
	}
}

// LoadTemplates loads HTML templates
func (fm *FrontendManager) LoadTemplates(pattern string) error {
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return err
	}

	fm.templates = tmpl
	return nil
}

// Render renders a template with data
func (fm *FrontendManager) Render(w http.ResponseWriter, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return fm.templates.ExecuteTemplate(w, name, data)
}

// VueJSIntegration provides Vue.js integration
type VueJSIntegration struct {
	appName string
	version string
}

// NewVueJSIntegration creates a new Vue.js integration
func NewVueJSIntegration(appName, version string) *VueJSIntegration {
	if version == "" {
		version = "3.3.4" // Default Vue 3 version
	}

	return &VueJSIntegration{
		appName: appName,
		version: version,
	}
}

// GenerateVueApp generates a Vue.js application structure
func (v *VueJSIntegration) GenerateVueApp() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/vue@%s/dist/vue.global.js"></script>
</head>
<body>
    <div id="app">
        <div class="min-h-screen bg-gray-100">
            <nav class="bg-white shadow">
                <div class="max-w-7xl mx-auto px-4">
                    <div class="flex justify-between h-16">
                        <div class="flex items-center">
                            <h1 class="text-xl font-semibold">🐬 Dolphin Framework</h1>
                        </div>
                        <div class="flex items-center space-x-4">
                            <a href="#" class="text-gray-500 hover:text-gray-700">Dashboard</a>
                            <a href="#" class="text-gray-500 hover:text-gray-700">Profile</a>
                            <button @click="logout" class="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600">Logout</button>
                        </div>
                    </div>
                </div>
            </nav>
            
            <main class="max-w-7xl mx-auto py-6 px-4">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div v-for="item in items" :key="item.id" class="bg-white rounded-lg shadow p-6">
                        <h3 class="text-lg font-medium text-gray-900">{{ item.title }}</h3>
                        <p class="text-gray-600">{{ item.description }}</p>
                        <div class="mt-4">
                            <button @click="editItem(item)" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 mr-2">Edit</button>
                            <button @click="deleteItem(item.id)" class="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600">Delete</button>
                        </div>
                    </div>
                </div>
                
                <div class="mt-8">
                    <button @click="showAddForm = !showAddForm" class="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600">
                        Add New Item
                    </button>
                    
                    <div v-if="showAddForm" class="mt-4 bg-white rounded-lg shadow p-6">
                        <h3 class="text-lg font-medium text-gray-900 mb-4">Add New Item</h3>
                        <form @submit.prevent="addItem">
                            <div class="mb-4">
                                <label class="block text-sm font-medium text-gray-700 mb-2">Title</label>
                                <input v-model="newItem.title" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                            </div>
                            <div class="mb-4">
                                <label class="block text-sm font-medium text-gray-700 mb-2">Description</label>
                                <textarea v-model="newItem.description" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" rows="3"></textarea>
                            </div>
                            <div class="flex space-x-4">
                                <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Save</button>
                                <button type="button" @click="showAddForm = false" class="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600">Cancel</button>
                            </div>
                        </form>
                    </div>
                </div>
            </main>
        </div>
    </div>

    <script>
        const { createApp } = Vue;
        
        createApp({
            data() {
                return {
                    items: [],
                    showAddForm: false,
                    newItem: {
                        title: '',
                        description: ''
                    }
                }
            },
            mounted() {
                this.loadItems();
            },
            methods: {
                async loadItems() {
                    try {
                        const response = await fetch('/api/v1/items');
                        this.items = await response.json();
                    } catch (error) {
                        console.error('Error loading items:', error);
                    }
                },
                async addItem() {
                    try {
                        const response = await fetch('/api/v1/items', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify(this.newItem)
                        });
                        
                        if (response.ok) {
                            this.newItem = { title: '', description: '' };
                            this.showAddForm = false;
                            this.loadItems();
                        }
                    } catch (error) {
                        console.error('Error adding item:', error);
                    }
                },
                async editItem(item) {
                    // Implement edit functionality
                    console.log('Edit item:', item);
                },
                async deleteItem(id) {
                    if (confirm('Are you sure you want to delete this item?')) {
                        try {
                            await fetch('/api/v1/items/' + id, {
                                method: 'DELETE'
                            });
                            this.loadItems();
                        } catch (error) {
                            console.error('Error deleting item:', error);
                        }
                    }
                },
                async logout() {
                    try {
                        await fetch('/api/v1/auth/logout', {
                            method: 'POST'
                        });
                        window.location.href = '/';
                    } catch (error) {
                        console.error('Error logging out:', error);
                    }
                }
            }
        }).mount('#app');
    </script>
</body>
</html>`, v.appName, v.version)
}

// ReactJSIntegration provides React.js integration
type ReactJSIntegration struct {
	appName string
	version string
}

// NewReactJSIntegration creates a new React.js integration
func NewReactJSIntegration(appName, version string) *ReactJSIntegration {
	if version == "" {
		version = "18.2.0" // Default React 18 version
	}

	return &ReactJSIntegration{
		appName: appName,
		version: version,
	}
}

// GenerateReactApp generates a React.js application structure
func (r *ReactJSIntegration) GenerateReactApp() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script crossorigin src="https://unpkg.com/react@%s/umd/react.production.min.js"></script>
    <script crossorigin src="https://unpkg.com/react-dom@%s/umd/react-dom.production.min.js"></script>
    <script src="https://unpkg.com/@babel/standalone/babel.min.js"></script>
</head>
<body>
    <div id="root"></div>

    <script type="text/babel">
        const { useState, useEffect } = React;
        
        function App() {
            const [items, setItems] = useState([]);
            const [showAddForm, setShowAddForm] = useState(false);
            const [newItem, setNewItem] = useState({ title: '', description: '' });
            
            useEffect(() => {
                loadItems();
            }, []);
            
            const loadItems = async () => {
                try {
                    const response = await fetch('/api/v1/items');
                    const data = await response.json();
                    setItems(data);
                } catch (error) {
                    console.error('Error loading items:', error);
                }
            };
            
            const addItem = async () => {
                try {
                    const response = await fetch('/api/v1/items', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify(newItem)
                    });
                    
                    if (response.ok) {
                        setNewItem({ title: '', description: '' });
                        setShowAddForm(false);
                        loadItems();
                    }
                } catch (error) {
                    console.error('Error adding item:', error);
                }
            };
            
            const deleteItem = async (id) => {
                if (window.confirm('Are you sure you want to delete this item?')) {
                    try {
                        await fetch('/api/v1/items/' + id, {
                            method: 'DELETE'
                        });
                        loadItems();
                    } catch (error) {
                        console.error('Error deleting item:', error);
                    }
                }
            };
            
            const logout = async () => {
                try {
                    await fetch('/api/v1/auth/logout', {
                        method: 'POST'
                    });
                    window.location.href = '/';
                } catch (error) {
                    console.error('Error logging out:', error);
                }
            };
            
            return (
                <div className="min-h-screen bg-gray-100">
                    <nav className="bg-white shadow">
                        <div className="max-w-7xl mx-auto px-4">
                            <div className="flex justify-between h-16">
                                <div className="flex items-center">
                                    <h1 className="text-xl font-semibold">🐬 Dolphin Framework</h1>
                                </div>
                                <div className="flex items-center space-x-4">
                                    <a href="#" className="text-gray-500 hover:text-gray-700">Dashboard</a>
                                    <a href="#" className="text-gray-500 hover:text-gray-700">Profile</a>
                                    <button onClick={logout} className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600">Logout</button>
                                </div>
                            </div>
                        </div>
                    </nav>
                    
                    <main className="max-w-7xl mx-auto py-6 px-4">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                            {items.map(item => (
                                <div key={item.id} className="bg-white rounded-lg shadow p-6">
                                    <h3 className="text-lg font-medium text-gray-900">{item.title}</h3>
                                    <p className="text-gray-600">{item.description}</p>
                                    <div className="mt-4">
                                        <button className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 mr-2">Edit</button>
                                        <button onClick={() => deleteItem(item.id)} className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600">Delete</button>
                                    </div>
                                </div>
                            ))}
                        </div>
                        
                        <div className="mt-8">
                            <button onClick={() => setShowAddForm(!showAddForm)} className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600">
                                Add New Item
                            </button>
                            
                            {showAddForm && (
                                <div className="mt-4 bg-white rounded-lg shadow p-6">
                                    <h3 className="text-lg font-medium text-gray-900 mb-4">Add New Item</h3>
                                    <form onSubmit={(e) => { e.preventDefault(); addItem(); }}>
                                        <div className="mb-4">
                                            <label className="block text-sm font-medium text-gray-700 mb-2">Title</label>
                                            <input 
                                                type="text" 
                                                value={newItem.title}
                                                onChange={(e) => setNewItem({...newItem, title: e.target.value})}
                                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
                                                required 
                                            />
                                        </div>
                                        <div className="mb-4">
                                            <label className="block text-sm font-medium text-gray-700 mb-2">Description</label>
                                            <textarea 
                                                value={newItem.description}
                                                onChange={(e) => setNewItem({...newItem, description: e.target.value})}
                                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
                                                rows="3"
                                            />
                                        </div>
                                        <div className="flex space-x-4">
                                            <button type="submit" className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Save</button>
                                            <button type="button" onClick={() => setShowAddForm(false)} className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600">Cancel</button>
                                        </div>
                                    </form>
                                </div>
                            )}
                        </div>
                    </main>
                </div>
            );
        }
        
        ReactDOM.render(<App />, document.getElementById('root'));
    </script>
</body>
</html>`, r.appName, r.version, r.version)
}

// TailwindCSSIntegration provides Tailwind CSS integration
type TailwindCSSIntegration struct {
	version string
}

// NewTailwindCSSIntegration creates a new Tailwind CSS integration
func NewTailwindCSSIntegration(version string) *TailwindCSSIntegration {
	if version == "" {
		version = "3.3.0" // Default Tailwind CSS version
	}

	return &TailwindCSSIntegration{
		version: version,
	}
}

// GenerateTailwindConfig generates a Tailwind CSS configuration
func (t *TailwindCSSIntegration) GenerateTailwindConfig() string {
	return `module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./resources/**/*.{js,ts,jsx,tsx}",
    "./public/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui'],
      },
    },
  },
  plugins: [],
}`
}

// GenerateTailwindCSS generates Tailwind CSS styles
func (t *TailwindCSSIntegration) GenerateTailwindCSS() string {
	return `@tailwind base;
@tailwind components;
@tailwind utilities;

@layer components {
  .btn {
    @apply px-4 py-2 rounded font-medium transition-colors duration-200;
  }
  
  .btn-primary {
    @apply bg-blue-500 text-white hover:bg-blue-600;
  }
  
  .btn-secondary {
    @apply bg-gray-500 text-white hover:bg-gray-600;
  }
  
  .btn-danger {
    @apply bg-red-500 text-white hover:bg-red-600;
  }
  
  .btn-success {
    @apply bg-green-500 text-white hover:bg-green-600;
  }
  
  .card {
    @apply bg-white rounded-lg shadow p-6;
  }
  
  .form-input {
    @apply w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  
  .form-label {
    @apply block text-sm font-medium text-gray-700 mb-2;
  }
}`
}
