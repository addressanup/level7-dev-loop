package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	maxFindings    = 50
	maxMessageSize = 240
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

func sortFindings(findings []finding) {
	sort.Slice(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]
		return strings.Join([]string{left.rule, left.subject, left.message, left.next}, "\x00") <
			strings.Join([]string{right.rule, right.subject, right.message, right.next}, "\x00")
	})
}

func printFindings(findings []finding) {
	sortFindings(findings)
	limit := len(findings)
	if limit > maxFindings {
		limit = maxFindings
	}
	for _, item := range findings[:limit] {
		fmt.Printf("BLOCKED rule=%s subject=%s message=%q next=%q\n", item.rule, item.subject, item.message, item.next)
	}
	if len(findings) > limit {
		fmt.Printf("BLOCKED rule=BCTL-099 subject=findings message=%q next=%q\n",
			fmt.Sprintf("%d additional findings omitted", len(findings)-limit),
			"fix the reported findings and rerun the complete gate")
	}
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
