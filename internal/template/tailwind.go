package template

import (
	htmltemplate "html/template"
	"net/http"
)

// TailwindHelpers provides TailwindCSS-specific helper functions
type TailwindHelpers struct {
	request *http.Request
}

// NewTailwindHelpers creates a new Tailwind helpers instance
func NewTailwindHelpers(r *http.Request) *TailwindHelpers {
	return &TailwindHelpers{request: r}
}

// GetTailwindHelpers returns TailwindCSS helper functions for Fin templates
func GetTailwindHelpers(r *http.Request) htmltemplate.FuncMap {
	helpers := NewTailwindHelpers(r)
	return htmltemplate.FuncMap{
		// TailwindCSS utility classes
		"btn":        helpers.Btn,
		"btn_primary": helpers.BtnPrimary,
		"btn_secondary": helpers.BtnSecondary,
		"btn_success": helpers.BtnSuccess,
		"btn_danger": helpers.BtnDanger,
		"btn_warning": helpers.BtnWarning,
		"btn_info": helpers.BtnInfo,
		"card":      helpers.Card,
		"container": helpers.Container,
		"input":     helpers.Input,
		"label":     helpers.Label,
		"alert":     helpers.Alert,
		"badge":     helpers.Badge,
	}
}

// Btn returns Tailwind button classes
func (t *TailwindHelpers) Btn(variant string) string {
	base := "px-4 py-2 rounded font-medium transition-colors"
	switch variant {
	case "primary":
		return base + " bg-blue-600 text-white hover:bg-blue-700"
	case "secondary":
		return base + " bg-gray-600 text-white hover:bg-gray-700"
	case "success":
		return base + " bg-green-600 text-white hover:bg-green-700"
	case "danger":
		return base + " bg-red-600 text-white hover:bg-red-700"
	case "warning":
		return base + " bg-yellow-600 text-white hover:bg-yellow-700"
	case "info":
		return base + " bg-cyan-600 text-white hover:bg-cyan-700"
	case "outline":
		return base + " border-2 border-gray-300 text-gray-700 hover:bg-gray-50"
	default:
		return base + " bg-gray-600 text-white hover:bg-gray-700"
	}
}

// BtnPrimary returns primary button classes
func (t *TailwindHelpers) BtnPrimary() string {
	return t.Btn("primary")
}

// BtnSecondary returns secondary button classes
func (t *TailwindHelpers) BtnSecondary() string {
	return t.Btn("secondary")
}

// BtnSuccess returns success button classes
func (t *TailwindHelpers) BtnSuccess() string {
	return t.Btn("success")
}

// BtnDanger returns danger button classes
func (t *TailwindHelpers) BtnDanger() string {
	return t.Btn("danger")
}

// BtnWarning returns warning button classes
func (t *TailwindHelpers) BtnWarning() string {
	return t.Btn("warning")
}

// BtnInfo returns info button classes
func (t *TailwindHelpers) BtnInfo() string {
	return t.Btn("info")
}

// Card returns Tailwind card classes
func (t *TailwindHelpers) Card() string {
	return "bg-white rounded-lg shadow-md p-6"
}

// Container returns Tailwind container classes
func (t *TailwindHelpers) Container() string {
	return "max-w-7xl mx-auto px-4 sm:px-6 lg:px-8"
}

// Input returns Tailwind input classes
func (t *TailwindHelpers) Input() string {
	return "mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
}

// Label returns Tailwind label classes
func (t *TailwindHelpers) Label() string {
	return "block text-sm font-medium text-gray-700"
}

// Alert returns Tailwind alert classes
func (t *TailwindHelpers) Alert(variant string) string {
	base := "p-4 rounded"
	switch variant {
	case "success":
		return base + " bg-green-100 text-green-800 border border-green-200"
	case "error", "danger":
		return base + " bg-red-100 text-red-800 border border-red-200"
	case "warning":
		return base + " bg-yellow-100 text-yellow-800 border border-yellow-200"
	case "info":
		return base + " bg-blue-100 text-blue-800 border border-blue-200"
	default:
		return base + " bg-gray-100 text-gray-800 border border-gray-200"
	}
}

// Badge returns Tailwind badge classes
func (t *TailwindHelpers) Badge(variant string) string {
	base := "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
	switch variant {
	case "primary":
		return base + " bg-blue-100 text-blue-800"
	case "success":
		return base + " bg-green-100 text-green-800"
	case "danger":
		return base + " bg-red-100 text-red-800"
	case "warning":
		return base + " bg-yellow-100 text-yellow-800"
	case "info":
		return base + " bg-cyan-100 text-cyan-800"
	default:
		return base + " bg-gray-100 text-gray-800"
	}
}

