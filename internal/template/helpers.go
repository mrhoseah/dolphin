package template

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// registerDefaultHelpers registers default template helpers
func (e *Engine) registerDefaultHelpers() {
	// String helpers
	e.RegisterHelper("upper", e.upperHelper)
	e.RegisterHelper("lower", e.lowerHelper)
	e.RegisterHelper("title", e.titleHelper)
	e.RegisterHelper("capitalize", e.capitalizeHelper)
	e.RegisterHelper("trim", e.trimHelper)
	e.RegisterHelper("replace", e.replaceHelper)
	e.RegisterHelper("truncate", e.truncateHelper)
	e.RegisterHelper("slug", e.slugHelper)
	e.RegisterHelper("pluralize", e.pluralizeHelper)
	e.RegisterHelper("singularize", e.singularizeHelper)
	
	// Number helpers
	e.RegisterHelper("add", e.addHelper)
	e.RegisterHelper("subtract", e.subtractHelper)
	e.RegisterHelper("multiply", e.multiplyHelper)
	e.RegisterHelper("divide", e.divideHelper)
	e.RegisterHelper("modulo", e.moduloHelper)
	e.RegisterHelper("round", e.roundHelper)
	e.RegisterHelper("ceil", e.ceilHelper)
	e.RegisterHelper("floor", e.floorHelper)
	e.RegisterHelper("abs", e.absHelper)
	e.RegisterHelper("min", e.minHelper)
	e.RegisterHelper("max", e.maxHelper)
	
	// Date/Time helpers
	e.RegisterHelper("now", e.nowHelper)
	e.RegisterHelper("formatDate", e.formatDateHelper)
	e.RegisterHelper("formatTime", e.formatTimeHelper)
	e.RegisterHelper("formatDateTime", e.formatDateTimeHelper)
	e.RegisterHelper("timeAgo", e.timeAgoHelper)
	e.RegisterHelper("timeUntil", e.timeUntilHelper)
	e.RegisterHelper("isToday", e.isTodayHelper)
	e.RegisterHelper("isYesterday", e.isYesterdayHelper)
	e.RegisterHelper("isTomorrow", e.isTomorrowHelper)
	
	// Array/Slice helpers
	e.RegisterHelper("join", e.joinHelper)
	e.RegisterHelper("split", e.splitHelper)
	e.RegisterHelper("first", e.firstHelper)
	e.RegisterHelper("last", e.lastHelper)
	e.RegisterHelper("length", e.lengthHelper)
	e.RegisterHelper("contains", e.containsHelper)
	e.RegisterHelper("index", e.indexHelper)
	e.RegisterHelper("slice", e.sliceHelper)
	e.RegisterHelper("reverse", e.reverseHelper)
	e.RegisterHelper("sort", e.sortHelper)
	e.RegisterHelper("unique", e.uniqueHelper)
	
	// Object/Map helpers
	e.RegisterHelper("keys", e.keysHelper)
	e.RegisterHelper("values", e.valuesHelper)
	e.RegisterHelper("hasKey", e.hasKeyHelper)
	e.RegisterHelper("get", e.getHelper)
	e.RegisterHelper("set", e.setHelper)
	e.RegisterHelper("merge", e.mergeHelper)
	
	// HTML helpers
	e.RegisterHelper("escape", e.escapeHelper)
	e.RegisterHelper("unescape", e.unescapeHelper)
	e.RegisterHelper("stripTags", e.stripTagsHelper)
	e.RegisterHelper("linkify", e.linkifyHelper)
	e.RegisterHelper("nl2br", e.nl2brHelper)
	e.RegisterHelper("br2nl", e.br2nlHelper)
	
	// URL helpers
	e.RegisterHelper("url", e.urlHelper)
	e.RegisterHelper("asset", e.assetHelper)
	e.RegisterHelper("route", e.routeHelper)
	e.RegisterHelper("query", e.queryHelper)
	e.RegisterHelper("fragment", e.fragmentHelper)
	
	// Security helpers
	e.RegisterHelper("csrf", e.csrfHelper)
	e.RegisterHelper("hash", e.hashHelper)
	e.RegisterHelper("random", e.randomHelper)
	e.RegisterHelper("uuid", e.uuidHelper)
	
	// Conditional helpers
	e.RegisterHelper("if", e.ifHelper)
	e.RegisterHelper("unless", e.unlessHelper)
	e.RegisterHelper("eq", e.eqHelper)
	e.RegisterHelper("ne", e.neHelper)
	e.RegisterHelper("gt", e.gtHelper)
	e.RegisterHelper("gte", e.gteHelper)
	e.RegisterHelper("lt", e.ltHelper)
	e.RegisterHelper("lte", e.lteHelper)
	e.RegisterHelper("and", e.andHelper)
	e.RegisterHelper("or", e.orHelper)
	e.RegisterHelper("not", e.notHelper)
	
	// Loop helpers
	e.RegisterHelper("range", e.rangeHelper)
	e.RegisterHelper("times", e.timesHelper)
	e.RegisterHelper("each", e.eachHelper)
	
	// Utility helpers
	e.RegisterHelper("default", e.defaultHelper)
	e.RegisterHelper("coalesce", e.coalesceHelper)
	e.RegisterHelper("empty", e.emptyHelper)
	e.RegisterHelper("present", e.presentHelper)
	e.RegisterHelper("blank", e.blankHelper)
	e.RegisterHelper("nil", e.nilHelper)
}

// String helpers
func (e *Engine) upperHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToUpper(fmt.Sprintf("%v", args[0])), nil
}

func (e *Engine) lowerHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToLower(fmt.Sprintf("%v", args[0])), nil
}

func (e *Engine) titleHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.Title(fmt.Sprintf("%v", args[0])), nil
}

func (e *Engine) capitalizeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	if len(str) == 0 {
		return "", nil
	}
	return strings.ToUpper(str[:1]) + strings.ToLower(str[1:]), nil
}

func (e *Engine) trimHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.TrimSpace(fmt.Sprintf("%v", args[0])), nil
}

func (e *Engine) replaceHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("replace helper requires 3 arguments: string, old, new")
	}
	str := fmt.Sprintf("%v", args[0])
	old := fmt.Sprintf("%v", args[1])
	new := fmt.Sprintf("%v", args[2])
	return strings.ReplaceAll(str, old, new), nil
}

func (e *Engine) truncateHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("truncate helper requires 2 arguments: string, length")
	}
	str := fmt.Sprintf("%v", args[0])
	length, err := strconv.Atoi(fmt.Sprintf("%v", args[1]))
	if err != nil {
		return "", fmt.Errorf("invalid length: %v", err)
	}
	if len(str) <= length {
		return str, nil
	}
	return str[:length] + "...", nil
}

func (e *Engine) slugHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	// Convert to lowercase
	str = strings.ToLower(str)
	// Replace spaces with hyphens
	str = strings.ReplaceAll(str, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	str = reg.ReplaceAllString(str, "")
	return str, nil
}

func (e *Engine) pluralizeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	word := fmt.Sprintf("%v", args[0])
	// Simple pluralization rules
	if strings.HasSuffix(word, "y") {
		return strings.TrimSuffix(word, "y") + "ies", nil
	}
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "sh") || strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") {
		return word + "es", nil
	}
	return word + "s", nil
}

func (e *Engine) singularizeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	word := fmt.Sprintf("%v", args[0])
	// Simple singularization rules
	if strings.HasSuffix(word, "ies") {
		return strings.TrimSuffix(word, "ies") + "y", nil
	}
	if strings.HasSuffix(word, "es") {
		return strings.TrimSuffix(word, "es"), nil
	}
	if strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s"), nil
	}
	return word, nil
}

// Number helpers
func (e *Engine) addHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("add helper requires at least 2 arguments")
	}
	result := 0.0
	for _, arg := range args {
		if num, err := strconv.ParseFloat(fmt.Sprintf("%v", arg), 64); err == nil {
			result += num
		}
	}
	return result, nil
}

func (e *Engine) subtractHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("subtract helper requires at least 2 arguments")
	}
	a, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
	if err != nil {
		return 0, err
	}
	return a - b, nil
}

func (e *Engine) multiplyHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("multiply helper requires at least 2 arguments")
	}
	result := 1.0
	for _, arg := range args {
		if num, err := strconv.ParseFloat(fmt.Sprintf("%v", arg), 64); err == nil {
			result *= num
		}
	}
	return result, nil
}

func (e *Engine) divideHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("divide helper requires at least 2 arguments")
	}
	a, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
	if err != nil {
		return 0, err
	}
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func (e *Engine) moduloHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("modulo helper requires at least 2 arguments")
	}
	a, err := strconv.ParseInt(fmt.Sprintf("%v", args[0]), 10, 64)
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseInt(fmt.Sprintf("%v", args[1]), 10, 64)
	if err != nil {
		return 0, err
	}
	if b == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}
	return a % b, nil
}

func (e *Engine) roundHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("round helper requires at least 1 argument")
	}
	num, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return 0, err
	}
	precision := 0
	if len(args) > 1 {
		precision, _ = strconv.Atoi(fmt.Sprintf("%v", args[1]))
	}
	multiplier := math.Pow(10, float64(precision))
	return math.Round(num*multiplier) / multiplier, nil
}

func (e *Engine) ceilHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("ceil helper requires at least 1 argument")
	}
	num, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return 0, err
	}
	return math.Ceil(num), nil
}

func (e *Engine) floorHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("floor helper requires at least 1 argument")
	}
	num, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return 0, err
	}
	return math.Floor(num), nil
}

func (e *Engine) absHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("abs helper requires at least 1 argument")
	}
	num, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return 0, err
	}
	return math.Abs(num), nil
}

func (e *Engine) minHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("min helper requires at least 1 argument")
	}
	min := math.Inf(1)
	for _, arg := range args {
		if num, err := strconv.ParseFloat(fmt.Sprintf("%v", arg), 64); err == nil {
			if num < min {
				min = num
			}
		}
	}
	return min, nil
}

func (e *Engine) maxHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("max helper requires at least 1 argument")
	}
	max := math.Inf(-1)
	for _, arg := range args {
		if num, err := strconv.ParseFloat(fmt.Sprintf("%v", arg), 64); err == nil {
			if num > max {
				max = num
			}
		}
	}
	return max, nil
}

// Date/Time helpers
func (e *Engine) nowHelper(args ...interface{}) (interface{}, error) {
	return time.Now(), nil
}

func (e *Engine) formatDateHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("formatDate helper requires 2 arguments: date, format")
	}
	
	var t time.Time
	switch v := args[0].(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid date type")
	}
	
	format := fmt.Sprintf("%v", args[1])
	return t.Format(format), nil
}

func (e *Engine) formatTimeHelper(args ...interface{}) (interface{}, error) {
	return e.formatDateHelper(args...)
}

func (e *Engine) formatDateTimeHelper(args ...interface{}) (interface{}, error) {
	return e.formatDateHelper(args...)
}

func (e *Engine) timeAgoHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("timeAgo helper requires at least 1 argument")
	}
	
	var t time.Time
	switch v := args[0].(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid date type")
	}
	
	now := time.Now()
	duration := now.Sub(t)
	
	if duration < time.Minute {
		return "just now", nil
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralSuffix(minutes)), nil
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralSuffix(hours)), nil
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralSuffix(days)), nil
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d month%s ago", months, pluralSuffix(months)), nil
	} else {
		years := int(duration.Hours() / (24 * 365))
		return fmt.Sprintf("%d year%s ago", years, pluralSuffix(years)), nil
	}
}

func (e *Engine) timeUntilHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("timeUntil helper requires at least 1 argument")
	}
	
	var t time.Time
	switch v := args[0].(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid date type")
	}
	
	now := time.Now()
	duration := t.Sub(now)
	
	if duration < 0 {
		return "overdue", nil
	} else if duration < time.Minute {
		return "in a moment", nil
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("in %d minute%s", minutes, pluralSuffix(minutes)), nil
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("in %d hour%s", hours, pluralSuffix(hours)), nil
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("in %d day%s", days, pluralSuffix(days)), nil
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("in %d month%s", months, pluralSuffix(months)), nil
	} else {
		years := int(duration.Hours() / (24 * 365))
		return fmt.Sprintf("in %d year%s", years, pluralSuffix(years)), nil
	}
}

func (e *Engine) isTodayHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("isToday helper requires at least 1 argument")
	}
	
	var t time.Time
	switch v := args[0].(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("invalid date type")
	}
	
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay(), nil
}

func (e *Engine) isYesterdayHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("isYesterday helper requires at least 1 argument")
	}
	
	var t time.Time
	switch v := args[0].(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("invalid date type")
	}
	
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay(), nil
}

func (e *Engine) isTomorrowHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("isTomorrow helper requires at least 1 argument")
	}
	
	var t time.Time
	switch v := args[0].(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("invalid date type")
	}
	
	tomorrow := time.Now().AddDate(0, 0, 1)
	return t.Year() == tomorrow.Year() && t.YearDay() == tomorrow.YearDay(), nil
}

// Array/Slice helpers
func (e *Engine) joinHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("join helper requires at least 2 arguments: array, separator")
	}
	
	separator := fmt.Sprintf("%v", args[1])
	var parts []string
	
	switch v := args[0].(type) {
	case []string:
		parts = v
	case []interface{}:
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
	default:
		return "", fmt.Errorf("invalid array type")
	}
	
	return strings.Join(parts, separator), nil
}

func (e *Engine) splitHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return []string{}, fmt.Errorf("split helper requires at least 2 arguments: string, separator")
	}
	
	str := fmt.Sprintf("%v", args[0])
	separator := fmt.Sprintf("%v", args[1])
	
	return strings.Split(str, separator), nil
}

func (e *Engine) firstHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("first helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		if len(v) > 0 {
			return v[0], nil
		}
		return "", nil
	case []interface{}:
		if len(v) > 0 {
			return v[0], nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid array type")
	}
}

func (e *Engine) lastHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("last helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
		return "", nil
	case []interface{}:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid array type")
	}
}

func (e *Engine) lengthHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("length helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		return len(v), nil
	case []interface{}:
		return len(v), nil
	case string:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	default:
		return 0, fmt.Errorf("invalid type for length")
	}
}

func (e *Engine) containsHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("contains helper requires at least 2 arguments: array, item")
	}
	
	item := fmt.Sprintf("%v", args[1])
	
	switch v := args[0].(type) {
	case []string:
		for _, s := range v {
			if s == item {
				return true, nil
			}
		}
		return false, nil
	case []interface{}:
		for _, i := range v {
			if fmt.Sprintf("%v", i) == item {
				return true, nil
			}
		}
		return false, nil
	case string:
		return strings.Contains(v, item), nil
	default:
		return false, fmt.Errorf("invalid type for contains")
	}
}

func (e *Engine) indexHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return -1, fmt.Errorf("index helper requires at least 2 arguments: array, item")
	}
	
	item := fmt.Sprintf("%v", args[1])
	
	switch v := args[0].(type) {
	case []string:
		for i, s := range v {
			if s == item {
				return i, nil
			}
		}
		return -1, nil
	case []interface{}:
		for i, i2 := range v {
			if fmt.Sprintf("%v", i2) == item {
				return i, nil
			}
		}
		return -1, nil
	default:
		return -1, fmt.Errorf("invalid type for index")
	}
}

func (e *Engine) sliceHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 3 {
		return []string{}, fmt.Errorf("slice helper requires at least 3 arguments: array, start, end")
	}
	
	start, err := strconv.Atoi(fmt.Sprintf("%v", args[1]))
	if err != nil {
		return []string{}, err
	}
	end, err := strconv.Atoi(fmt.Sprintf("%v", args[2]))
	if err != nil {
		return []string{}, err
	}
	
	switch v := args[0].(type) {
	case []string:
		if start < 0 || end > len(v) || start > end {
			return []string{}, fmt.Errorf("invalid slice bounds")
		}
		return v[start:end], nil
	case []interface{}:
		if start < 0 || end > len(v) || start > end {
			return []interface{}{}, fmt.Errorf("invalid slice bounds")
		}
		return v[start:end], nil
	default:
		return []string{}, fmt.Errorf("invalid type for slice")
	}
}

func (e *Engine) reverseHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []string{}, fmt.Errorf("reverse helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		result := make([]string, len(v))
		for i, j := 0, len(v)-1; i < len(v); i, j = i+1, j-1 {
			result[i] = v[j]
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, j := 0, len(v)-1; i < len(v); i, j = i+1, j-1 {
			result[i] = v[j]
		}
		return result, nil
	default:
		return []string{}, fmt.Errorf("invalid type for reverse")
	}
}

func (e *Engine) sortHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []string{}, fmt.Errorf("sort helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		result := make([]string, len(v))
		copy(result, v)
		// Simple bubble sort
		for i := 0; i < len(result); i++ {
			for j := i + 1; j < len(result); j++ {
				if result[i] > result[j] {
					result[i], result[j] = result[j], result[i]
				}
			}
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		copy(result, v)
		// Simple bubble sort
		for i := 0; i < len(result); i++ {
			for j := i + 1; j < len(result); j++ {
				if fmt.Sprintf("%v", result[i]) > fmt.Sprintf("%v", result[j]) {
					result[i], result[j] = result[j], result[i]
				}
			}
		}
		return result, nil
	default:
		return []string{}, fmt.Errorf("invalid type for sort")
	}
}

func (e *Engine) uniqueHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []string{}, fmt.Errorf("unique helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		seen := make(map[string]bool)
		result := []string{}
		for _, item := range v {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
		return result, nil
	case []interface{}:
		seen := make(map[string]bool)
		result := []interface{}{}
		for _, item := range v {
			key := fmt.Sprintf("%v", item)
			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result, nil
	default:
		return []string{}, fmt.Errorf("invalid type for unique")
	}
}

// Object/Map helpers
func (e *Engine) keysHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []string{}, fmt.Errorf("keys helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys, nil
	default:
		return []string{}, fmt.Errorf("invalid type for keys")
	}
}

func (e *Engine) valuesHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []interface{}{}, fmt.Errorf("values helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case map[string]interface{}:
		values := make([]interface{}, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values, nil
	default:
		return []interface{}{}, fmt.Errorf("invalid type for values")
	}
}

func (e *Engine) hasKeyHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("hasKey helper requires at least 2 arguments: map, key")
	}
	
	key := fmt.Sprintf("%v", args[1])
	
	switch v := args[0].(type) {
	case map[string]interface{}:
		_, exists := v[key]
		return exists, nil
	default:
		return false, fmt.Errorf("invalid type for hasKey")
	}
}

func (e *Engine) getHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("get helper requires at least 2 arguments: map, key")
	}
	
	key := fmt.Sprintf("%v", args[1])
	
	switch v := args[0].(type) {
	case map[string]interface{}:
		val, exists := v[key]
		if !exists {
			return nil, nil
		}
		return val, nil
	default:
		return nil, fmt.Errorf("invalid type for get")
	}
}

func (e *Engine) setHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("set helper requires at least 3 arguments: map, key, value")
	}
	
	key := fmt.Sprintf("%v", args[1])
	value := args[2]
	
	switch v := args[0].(type) {
	case map[string]interface{}:
		v[key] = value
		return v, nil
	default:
		return nil, fmt.Errorf("invalid type for set")
	}
}

func (e *Engine) mergeHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("merge helper requires at least 2 arguments: map1, map2")
	}
	
	result := make(map[string]interface{})
	
	for _, arg := range args {
		switch v := arg.(type) {
		case map[string]interface{}:
			for k, val := range v {
				result[k] = val
			}
		default:
			return nil, fmt.Errorf("invalid type for merge")
		}
	}
	
	return result, nil
}

// HTML helpers
func (e *Engine) escapeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	return html.EscapeString(fmt.Sprintf("%v", args[0])), nil
}

func (e *Engine) unescapeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	return html.UnescapeString(fmt.Sprintf("%v", args[0])), nil
}

func (e *Engine) stripTagsHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	// Simple tag stripping
	reg := regexp.MustCompile(`<[^>]*>`)
	return reg.ReplaceAllString(str, ""), nil
}

func (e *Engine) linkifyHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	// Simple URL detection
	urlReg := regexp.MustCompile(`(https?://[^\s]+)`)
	return urlReg.ReplaceAllString(str, `<a href="$1">$1</a>`), nil
}

func (e *Engine) nl2brHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	return strings.ReplaceAll(str, "\n", "<br>"), nil
}

func (e *Engine) br2nlHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	return strings.ReplaceAll(str, "<br>", "\n"), nil
}

// URL helpers
func (e *Engine) urlHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	path := fmt.Sprintf("%v", args[0])
	// Simple URL construction
	if strings.HasPrefix(path, "/") {
		return path, nil
	}
	return "/" + path, nil
}

func (e *Engine) assetHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	path := fmt.Sprintf("%v", args[0])
	// Simple asset URL construction
	if strings.HasPrefix(path, "/") {
		return "/assets" + path, nil
	}
	return "/assets/" + path, nil
}

func (e *Engine) routeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	route := fmt.Sprintf("%v", args[0])
	// Simple route URL construction
	return "/" + route, nil
}

func (e *Engine) queryHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", nil
	}
	url := fmt.Sprintf("%v", args[0])
	params := fmt.Sprintf("%v", args[1])
	
	if strings.Contains(url, "?") {
		return url + "&" + params, nil
	}
	return url + "?" + params, nil
}

func (e *Engine) fragmentHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", nil
	}
	url := fmt.Sprintf("%v", args[0])
	fragment := fmt.Sprintf("%v", args[1])
	
	return url + "#" + fragment, nil
}

// Security helpers
func (e *Engine) csrfHelper(args ...interface{}) (interface{}, error) {
	// This would typically generate a CSRF token
	// For now, return a placeholder
	return "csrf_token_placeholder", nil
}

func (e *Engine) hashHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	str := fmt.Sprintf("%v", args[0])
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:]), nil
}

func (e *Engine) randomHelper(args ...interface{}) (interface{}, error) {
	// Simple random string generation
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 10
	if len(args) > 0 {
		if l, err := strconv.Atoi(fmt.Sprintf("%v", args[0])); err == nil {
			length = l
		}
	}
	
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result), nil
}

func (e *Engine) uuidHelper(args ...interface{}) (interface{}, error) {
	// Simple UUID v4 generation
	// This is a simplified version
	return fmt.Sprintf("%x-%x-%x-%x-%x", 
		time.Now().UnixNano()&0xffffffff,
		time.Now().UnixNano()&0xffff,
		time.Now().UnixNano()&0xffff,
		time.Now().UnixNano()&0xffff,
		time.Now().UnixNano()&0xffffffffffff), nil
}

// Conditional helpers
func (e *Engine) ifHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("if helper requires at least 2 arguments: condition, value")
	}
	
	condition := false
	switch v := args[0].(type) {
	case bool:
		condition = v
	case string:
		condition = v != ""
	case int:
		condition = v != 0
	case float64:
		condition = v != 0
	default:
		condition = v != nil
	}
	
	if condition {
		return args[1], nil
	}
	
	if len(args) > 2 {
		return args[2], nil
	}
	
	return "", nil
}

func (e *Engine) unlessHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("unless helper requires at least 2 arguments: condition, value")
	}
	
	condition := false
	switch v := args[0].(type) {
	case bool:
		condition = v
	case string:
		condition = v != ""
	case int:
		condition = v != 0
	case float64:
		condition = v != 0
	default:
		condition = v != nil
	}
	
	if !condition {
		return args[1], nil
	}
	
	if len(args) > 2 {
		return args[2], nil
	}
	
	return "", nil
}

func (e *Engine) eqHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("eq helper requires at least 2 arguments")
	}
	return fmt.Sprintf("%v", args[0]) == fmt.Sprintf("%v", args[1]), nil
}

func (e *Engine) neHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("ne helper requires at least 2 arguments")
	}
	return fmt.Sprintf("%v", args[0]) != fmt.Sprintf("%v", args[1]), nil
}

func (e *Engine) gtHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("gt helper requires at least 2 arguments")
	}
	a, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
	if err != nil {
		return false, err
	}
	return a > b, nil
}

func (e *Engine) gteHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("gte helper requires at least 2 arguments")
	}
	a, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
	if err != nil {
		return false, err
	}
	return a >= b, nil
}

func (e *Engine) ltHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("lt helper requires at least 2 arguments")
	}
	a, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
	if err != nil {
		return false, err
	}
	return a < b, nil
}

func (e *Engine) lteHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("lte helper requires at least 2 arguments")
	}
	a, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
	if err != nil {
		return false, err
	}
	return a <= b, nil
}

func (e *Engine) andHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return false, nil
	}
	
	for _, arg := range args {
		condition := false
		switch v := arg.(type) {
		case bool:
			condition = v
		case string:
			condition = v != ""
		case int:
			condition = v != 0
		case float64:
			condition = v != 0
		default:
			condition = v != nil
		}
		if !condition {
			return false, nil
		}
	}
	
	return true, nil
}

func (e *Engine) orHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return false, nil
	}
	
	for _, arg := range args {
		condition := false
		switch v := arg.(type) {
		case bool:
			condition = v
		case string:
			condition = v != ""
		case int:
			condition = v != 0
		case float64:
			condition = v != 0
		default:
			condition = v != nil
		}
		if condition {
			return true, nil
		}
	}
	
	return false, nil
}

func (e *Engine) notHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return true, nil
	}
	
	condition := false
	switch v := args[0].(type) {
	case bool:
		condition = v
	case string:
		condition = v != ""
	case int:
		condition = v != 0
	case float64:
		condition = v != 0
	default:
		condition = v != nil
	}
	
	return !condition, nil
}

// Loop helpers
func (e *Engine) rangeHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []interface{}{}, fmt.Errorf("range helper requires at least 1 argument")
	}
	
	switch v := args[0].(type) {
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []interface{}:
		return v, nil
	default:
		return []interface{}{}, fmt.Errorf("invalid type for range")
	}
}

func (e *Engine) timesHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []interface{}{}, fmt.Errorf("times helper requires at least 1 argument")
	}
	
	count, err := strconv.Atoi(fmt.Sprintf("%v", args[0]))
	if err != nil {
		return []interface{}{}, err
	}
	
	result := make([]interface{}, count)
	for i := 0; i < count; i++ {
		result[i] = i
	}
	return result, nil
}

func (e *Engine) eachHelper(args ...interface{}) (interface{}, error) {
	return e.rangeHelper(args...)
}

// Utility helpers
func (e *Engine) defaultHelper(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("default helper requires at least 2 arguments: value, default")
	}
	
	value := args[0]
	defaultValue := args[1]
	
	if value == nil {
		return defaultValue, nil
	}
	
	switch v := value.(type) {
	case string:
		if v == "" {
			return defaultValue, nil
		}
	case int:
		if v == 0 {
			return defaultValue, nil
		}
	case float64:
		if v == 0 {
			return defaultValue, nil
		}
	case bool:
		if !v {
			return defaultValue, nil
		}
	}
	
	return value, nil
}

func (e *Engine) coalesceHelper(args ...interface{}) (interface{}, error) {
	for _, arg := range args {
		if arg != nil {
			switch v := arg.(type) {
			case string:
				if v != "" {
					return v, nil
				}
			case int:
				if v != 0 {
					return v, nil
				}
			case float64:
				if v != 0 {
					return v, nil
				}
			case bool:
				if v {
					return v, nil
				}
			default:
				return v, nil
			}
		}
	}
	return nil, nil
}

func (e *Engine) emptyHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return true, nil
	}
	
	switch v := args[0].(type) {
	case string:
		return v == "", nil
	case []string:
		return len(v) == 0, nil
	case []interface{}:
		return len(v) == 0, nil
	case map[string]interface{}:
		return len(v) == 0, nil
	case int:
		return v == 0, nil
	case float64:
		return v == 0, nil
	case bool:
		return !v, nil
	default:
		return v == nil, nil
	}
}

func (e *Engine) presentHelper(args ...interface{}) (interface{}, error) {
	result, err := e.emptyHelper(args...)
	if err != nil {
		return false, err
	}
	return !result.(bool), nil
}

func (e *Engine) blankHelper(args ...interface{}) (interface{}, error) {
	return e.emptyHelper(args...)
}

func (e *Engine) nilHelper(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return true, nil
	}
	return args[0] == nil, nil
}

// Helper function for pluralization
func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
