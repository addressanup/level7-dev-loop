package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	requirementIDPattern = regexp.MustCompile(`^L7-[A-Z]+-[0-9]{3}$`)
	fullRangePattern     = regexp.MustCompile(`^(L7-[A-Z]+-)([0-9]{3})(?:–([0-9]{3}))?$`)
	shortRangePattern    = regexp.MustCompile(`^([0-9]{3})(?:–([0-9]{3}))?$`)
	backlogOwnerPattern  = regexp.MustCompile(`^L7-BL-[0-9]{3}$`)
)

func firstMarkdownCell(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return "", false
	}
	remainder := trimmed[1:]
	end := strings.IndexByte(remainder, '|')
	if end < 0 {
		return "", false
	}
	return strings.TrimSpace(remainder[:end]), true
}

func splitMarkdownRow(line string) ([]string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
		return nil, false
	}
	parts := strings.Split(trimmed[1:len(trimmed)-1], "|")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts, true
}

func unquoteCodeCell(cell string) (string, bool) {
	if len(cell) < 2 || cell[0] != '`' || cell[len(cell)-1] != '`' {
		return "", false
	}
	value := cell[1 : len(cell)-1]
	if strings.ContainsRune(value, '`') {
		return "", false
	}
	return value, true
}

func expandRequirementExpression(expression string) ([]string, []finding) {
	normalized := strings.ReplaceAll(expression, "`", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	if normalized == "" {
		return nil, []finding{newFinding("TRACE-110", expression, "requirement expression is empty", "restore a valid requirement expression")}
	}

	var ids []string
	var currentPrefix string
	for _, token := range strings.Split(normalized, ",") {
		if token == "" {
			return nil, []finding{newFinding("TRACE-110", expression, "requirement expression contains an empty token", "remove the empty token")}
		}
		var startText, endText string
		if match := fullRangePattern.FindStringSubmatch(token); match != nil {
			currentPrefix = match[1]
			startText, endText = match[2], match[3]
		} else if match := shortRangePattern.FindStringSubmatch(token); match != nil && currentPrefix != "" {
			startText, endText = match[1], match[2]
		} else {
			return nil, []finding{newFinding("TRACE-110", token, "malformed or prefixless requirement token", "use an exact ID or zero-padded range")}
		}
		start, _ := strconv.Atoi(startText)
		end := start
		if endText != "" {
			end, _ = strconv.Atoi(endText)
		}
		if end < start {
			return nil, []finding{newFinding("TRACE-111", token, "requirement range is reversed", "put the lower ID first")}
		}
		if end-start > 999 {
			return nil, []finding{newFinding("TRACE-112", token, "requirement range exceeds the expansion limit", "narrow the range")}
		}
		for value := start; value <= end; value++ {
			id := fmt.Sprintf("%s%03d", currentPrefix, value)
			if !requirementIDPattern.MatchString(id) {
				return nil, []finding{newFinding("TRACE-110", id, "expanded requirement ID is malformed", "correct the range")}
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}
