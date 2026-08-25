package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	maxFindings          = 50
	maxCollectedFindings = maxFindings + 1
	maxMessageSize       = 240
	maxOutputBytes       = 48 << 10
)

type finding struct {
	rule    string
	subject string
	message string
	next    string
}

func newFinding(rule, subject, message, next string) finding {
	return finding{
		rule:    safeASCII(rule, 32),
		subject: safeASCII(subject, 120),
		message: safeASCII(message, maxMessageSize),
		next:    safeASCII(next, maxMessageSize),
	}
}

func appendFindings(current []finding, additional ...finding) []finding {
	for _, candidate := range additional {
		if len(current) < maxCollectedFindings {
			current = append(current, candidate)
			continue
		}

		largest := 0
		for index := 1; index < len(current); index++ {
			if findingLess(current[largest], current[index]) {
				largest = index
			}
		}
		if findingLess(candidate, current[largest]) {
			current[largest] = candidate
		}
	}
	return current
}

func sortFindings(findings []finding) {
	sort.Slice(findings, func(i, j int) bool {
		return findingLess(findings[i], findings[j])
	})
}

func findingLess(left, right finding) bool {
	if left.rule != right.rule {
		return left.rule < right.rule
	}
	if left.subject != right.subject {
		return left.subject < right.subject
	}
	if left.message != right.message {
		return left.message < right.message
	}
	return left.next < right.next
}

func printFindings(findings []finding) {
	fmt.Print(renderFindings(findings))
}

func renderFindings(findings []finding) string {
	sortFindings(findings)
	limit := len(findings)
	if limit > maxFindings {
		limit = maxFindings
	}
	var output strings.Builder
	for _, item := range findings[:limit] {
		line := fmt.Sprintf("BLOCKED rule=%s subject=%s message=%q next=%q\n", item.rule, item.subject, item.message, item.next)
		if output.Len()+len(line) > maxOutputBytes {
			break
		}
		output.WriteString(line)
	}
	if len(findings) > limit {
		line := fmt.Sprintf("BLOCKED rule=BCTL-099 subject=findings message=%q next=%q\n",
			"finding collection cap reached; additional findings omitted",
			"fix the reported findings and rerun the complete gate")
		if output.Len()+len(line) <= maxOutputBytes {
			output.WriteString(line)
		}
	}
	return output.String()
}

func safeASCII(value string, limit int) string {
	var builder strings.Builder
	for _, character := range value {
		if character >= 0x20 && character <= 0x7e {
			builder.WriteRune(character)
		} else {
			builder.WriteByte('?')
		}
		if builder.Len() >= limit {
			break
		}
	}
	return builder.String()
}
