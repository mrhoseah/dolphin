package template

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strings"
)

// HTMXHelpers provides HTMX-specific helper functions for Fin templates
type HTMXHelpers struct {
	request *http.Request
}

// NewHTMXHelpers creates a new HTMX helpers instance
func NewHTMXHelpers(r *http.Request) *HTMXHelpers {
	return &HTMXHelpers{request: r}
}

// GetHTMXHelpers returns HTMX helper functions for Fin templates
func GetHTMXHelpers(r *http.Request) htmltemplate.FuncMap {
	helpers := NewHTMXHelpers(r)
	return htmltemplate.FuncMap{
		// HTMX attributes
		"hx_get":      helpers.HXGet,
		"hx_post":     helpers.HXPost,
		"hx_put":      helpers.HXPut,
		"hx_patch":    helpers.HXPatch,
		"hx_delete":   helpers.HXDelete,
		"hx_trigger":  helpers.HXTrigger,
		"hx_target":   helpers.HXTarget,
		"hx_swap":     helpers.HXSwap,
		"hx_swap_oob": helpers.HXSwapOOB,
		"hx_boost":    helpers.HXBoost,
		"hx_confirm":  helpers.HXConfirm,
		"hx_disable":  helpers.HXDisable,
		"hx_ext":      helpers.HXExt,
		"hx_headers":  helpers.HXHeaders,
		"hx_history":  helpers.HXHistory,
		"hx_history_elt": helpers.HXHistoryElt,
		"hx_include":  helpers.HXInclude,
		"hx_indicator": helpers.HXIndicator,
		"hx_params":   helpers.HXParams,
		"hx_preserve": helpers.HXPreserve,
		"hx_prompt":   helpers.HXPrompt,
		"hx_push_url": helpers.HXPushURL,
		"hx_replace_url": helpers.HXReplaceURL,
		"hx_request":  helpers.HXRequest,
		"hx_select":   helpers.HXSelect,
		"hx_select_oob": helpers.HXSelectOOB,
		"hx_sse":      helpers.HXSSE,
		"hx_sync":     helpers.HXSync,
		"hx_validate": helpers.HXValidate,
		"hx_vals":     helpers.HXVals,
		"hx_ws":       helpers.HXWS,
		
		// HTMX helpers
		"hx_attrs":    helpers.HXAttrs,
		"hx_button":   helpers.HXButton,
		"hx_link":    helpers.HXLink,
		"hx_form":    helpers.HXForm,
	},
}

// HXGet returns hx-get attribute
func (h *HTMXHelpers) HXGet(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-get="%s"`, url))
}

// HXPost returns hx-post attribute
func (h *HTMXHelpers) HXPost(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-post="%s"`, url))
}

// HXPut returns hx-put attribute
func (h *HTMXHelpers) HXPut(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-put="%s"`, url))
}

// HXPatch returns hx-patch attribute
func (h *HTMXHelpers) HXPatch(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-patch="%s"`, url))
}

// HXDelete returns hx-delete attribute
func (h *HTMXHelpers) HXDelete(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-delete="%s"`, url))
}

// HXTrigger returns hx-trigger attribute
func (h *HTMXHelpers) HXTrigger(trigger string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-trigger="%s"`, trigger))
}

// HXTarget returns hx-target attribute
func (h *HTMXHelpers) HXTarget(target string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-target="%s"`, target))
}

// HXSwap returns hx-swap attribute
func (h *HTMXHelpers) HXSwap(swap string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-swap="%s"`, swap))
}

// HXSwapOOB returns hx-swap-oob attribute
func (h *HTMXHelpers) HXSwapOOB(value string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-swap-oob="%s"`, value))
}

// HXBoost returns hx-boost attribute
func (h *HTMXHelpers) HXBoost(value bool) htmltemplate.HTMLAttr {
	if value {
		return htmltemplate.HTMLAttr(`hx-boost="true"`)
	}
	return ""
}

// HXConfirm returns hx-confirm attribute
func (h *HTMXHelpers) HXConfirm(message string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-confirm="%s"`, message))
}

// HXDisable returns hx-disable attribute
func (h *HTMXHelpers) HXDisable(value bool) htmltemplate.HTMLAttr {
	if value {
		return htmltemplate.HTMLAttr(`hx-disable="true"`)
	}
	return htmltemplate.HTMLAttr("")
}

// HXExt returns hx-ext attribute
func (h *HTMXHelpers) HXExt(ext string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-ext="%s"`, ext))
}

// HXHeaders returns hx-headers attribute
func (h *HTMXHelpers) HXHeaders(headers string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-headers='%s'`, headers))
}

// HXHistory returns hx-history attribute
func (h *HTMXHelpers) HXHistory(value bool) htmltemplate.HTMLAttr {
	if value {
		return htmltemplate.HTMLAttr(`hx-history="true"`)
	}
	return htmltemplate.HTMLAttr("")
}

// HXHistoryElt returns hx-history-elt attribute
func (h *HTMXHelpers) HXHistoryElt(selector string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-history-elt="%s"`, selector))
}

// HXInclude returns hx-include attribute
func (h *HTMXHelpers) HXInclude(selector string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-include="%s"`, selector))
}

// HXIndicator returns hx-indicator attribute
func (h *HTMXHelpers) HXIndicator(selector string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-indicator="%s"`, selector))
}

// HXParams returns hx-params attribute
func (h *HTMXHelpers) HXParams(params string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-params="%s"`, params))
}

// HXPreserve returns hx-preserve attribute
func (h *HTMXHelpers) HXPreserve(value bool) htmltemplate.HTMLAttr {
	if value {
		return htmltemplate.HTMLAttr(`hx-preserve="true"`)
	}
	return htmltemplate.HTMLAttr("")
}

// HXPrompt returns hx-prompt attribute
func (h *HTMXHelpers) HXPrompt(message string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-prompt="%s"`, message))
}

// HXPushURL returns hx-push-url attribute
func (h *HTMXHelpers) HXPushURL(value interface{}) htmltemplate.HTMLAttr {
	if value == nil || value == false {
		return htmltemplate.HTMLAttr(`hx-push-url="false"`)
	}
	if value == true {
		return htmltemplate.HTMLAttr(`hx-push-url="true"`)
	}
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-push-url="%v"`, value))
}

// HXReplaceURL returns hx-replace-url attribute
func (h *HTMXHelpers) HXReplaceURL(value interface{}) htmltemplate.HTMLAttr {
	if value == nil || value == false {
		return htmltemplate.HTMLAttr(`hx-replace-url="false"`)
	}
	if value == true {
		return htmltemplate.HTMLAttr(`hx-replace-url="true"`)
	}
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-replace-url="%v"`, value))
}

// HXRequest returns hx-request attribute
func (h *HTMXHelpers) HXRequest(options string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-request='%s'`, options))
}

// HXSelect returns hx-select attribute
func (h *HTMXHelpers) HXSelect(selector string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-select="%s"`, selector))
}

// HXSelectOOB returns hx-select-oob attribute
func (h *HTMXHelpers) HXSelectOOB(value string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-select-oob="%s"`, value))
}

// HXSSE returns hx-sse attribute
func (h *HTMXHelpers) HXSSE(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-sse="%s"`, url))
}

// HXSync returns hx-sync attribute
func (h *HTMXHelpers) HXSync(sync string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-sync="%s"`, sync))
}

// HXValidate returns hx-validate attribute
func (h *HTMXHelpers) HXValidate(value bool) htmltemplate.HTMLAttr {
	if value {
		return htmltemplate.HTMLAttr(`hx-validate="true"`)
	}
	return htmltemplate.HTMLAttr("")
}

// HXVals returns hx-vals attribute
func (h *HTMXHelpers) HXVals(vals string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-vals='%s'`, vals))
}

// HXWS returns hx-ws attribute
func (h *HTMXHelpers) HXWS(url string) htmltemplate.HTMLAttr {
	return htmltemplate.HTMLAttr(fmt.Sprintf(`hx-ws="%s"`, url))
}

// HXAttrs returns multiple HTMX attributes as a string
func (h *HTMXHelpers) HXAttrs(attrs map[string]interface{}) htmltemplate.HTMLAttr {
	var parts []string
	for key, value := range attrs {
		if value != nil && value != false && value != "" {
			parts = append(parts, fmt.Sprintf(`hx-%s="%v"`, strings.TrimPrefix(key, "hx-"), value))
		}
	}
	return htmltemplate.HTMLAttr(strings.Join(parts, " "))
}

// HXButton creates an HTMX button
func (h *HTMXHelpers) HXButton(url, method, text string, attrs map[string]interface{}) htmltemplate.HTML {
	var attrStr strings.Builder
	
	// Add method attribute
	switch strings.ToUpper(method) {
	case "GET":
		attrStr.WriteString(fmt.Sprintf(`hx-get="%s"`, url))
	case "POST":
		attrStr.WriteString(fmt.Sprintf(`hx-post="%s"`, url))
	case "PUT":
		attrStr.WriteString(fmt.Sprintf(`hx-put="%s"`, url))
	case "PATCH":
		attrStr.WriteString(fmt.Sprintf(`hx-patch="%s"`, url))
	case "DELETE":
		attrStr.WriteString(fmt.Sprintf(`hx-delete="%s"`, url))
	default:
		attrStr.WriteString(fmt.Sprintf(`hx-get="%s"`, url))
	}
	
	// Add additional attributes
	for key, value := range attrs {
		if value != nil && value != false && value != "" {
			attrStr.WriteString(fmt.Sprintf(` %s="%v"`, key, value))
		}
	}
	
	return htmltemplate.HTML(fmt.Sprintf(`<button %s>%s</button>`, attrStr.String(), text))
}

// HXLink creates an HTMX link
func (h *HTMXHelpers) HXLink(url, text string, attrs map[string]interface{}) htmltemplate.HTML {
	var attrStr strings.Builder
	attrStr.WriteString(fmt.Sprintf(`hx-get="%s"`, url))
	
	// Add additional attributes
	for key, value := range attrs {
		if value != nil && value != false && value != "" {
			attrStr.WriteString(fmt.Sprintf(` %s="%v"`, key, value))
		}
	}
	
	return htmltemplate.HTML(fmt.Sprintf(`<a href="%s" %s>%s</a>`, url, attrStr.String(), text))
}

// HXForm creates an HTMX form
func (h *HTMXHelpers) HXForm(url, method string, attrs map[string]interface{}) htmltemplate.HTML {
	var attrStr strings.Builder
	
	// Add method attribute
	switch strings.ToUpper(method) {
	case "POST":
		attrStr.WriteString(fmt.Sprintf(`hx-post="%s"`, url))
	case "PUT":
		attrStr.WriteString(fmt.Sprintf(`hx-put="%s"`, url))
	case "PATCH":
		attrStr.WriteString(fmt.Sprintf(`hx-patch="%s"`, url))
	case "DELETE":
		attrStr.WriteString(fmt.Sprintf(`hx-delete="%s"`, url))
	default:
		attrStr.WriteString(fmt.Sprintf(`hx-post="%s"`, url))
	}
	
	// Add additional attributes
	for key, value := range attrs {
		if value != nil && value != false && value != "" {
			attrStr.WriteString(fmt.Sprintf(` %s="%v"`, key, value))
		}
	}
	
	return htmltemplate.HTML(fmt.Sprintf(`<form %s></form>`, attrStr.String()))
}

