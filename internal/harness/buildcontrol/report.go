package main

import (
	"fmt"
	"sort"
	"strings"
)

const maxFindings = 50

type finding struct {
	rule    string
	subject string
	message string
	next    string
}

func newFinding(rule, subject, message, next string) finding {
	return finding{safeASCII(rule, 32), safeASCII(subject, 160), safeASCII(message, 300), safeASCII(next, 300)}
}

func appendFindings(current []finding, additional ...finding) []finding {
	for _, item := range additional {
		if len(current) < maxFindings {
			current = append(current, item)
		}
	}
	return current
}

func renderFindings(findings []finding) string {
	sort.Slice(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if left.rule != right.rule {
			return left.rule < right.rule
		}
		if left.subject != right.subject {
			return left.subject < right.subject
		}
		return left.message < right.message
	})
	var output strings.Builder
	for _, item := range findings {
		fmt.Fprintf(&output, "BLOCKED rule=%s subject=%s message=%q next=%q\n", item.rule, item.subject, item.message, item.next)
	}
	return output.String()
}

func printFindings(findings []finding) { fmt.Print(renderFindings(findings)) }

func safeASCII(value string, limit int) string {
	var output strings.Builder
	for _, character := range value {
		if character >= 0x20 && character <= 0x7e {
			output.WriteRune(character)
		} else {
			output.WriteByte('?')
		}
		if output.Len() >= limit {
			break
		}
	}
	return output.String()
}
