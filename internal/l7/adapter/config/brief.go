package config

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const MaxBrief = 256 << 10

var (
	changeIDPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)
	gitIDPattern    = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
)

func BriefPath(changeID string) string {
	return "docs/artifacts/changes/" + changeID + ".md"
}

func CreateBrief(root string, brief domain.ChangeBrief) error {
	if !filepath.IsAbs(root) {
		return errors.New("repository root must be absolute")
	}
	if err := ValidateBrief(brief); err != nil {
		return err
	}
	data, err := RenderBrief(brief)
	if err != nil {
		return err
	}
	destination := filepath.Join(root, filepath.FromSlash(brief.Path))
	if err := localfile.EnsureDirectory(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	return localfile.AtomicCreate(destination, data, 0o644)
}

func LoadBrief(root, relative string) (domain.ChangeBrief, error) {
	if !filepath.IsAbs(root) {
		return domain.ChangeBrief{}, errors.New("repository root must be absolute")
	}
	if err := ValidateRepositoryPath(relative); err != nil {
		return domain.ChangeBrief{}, err
	}
	data, err := localfile.Read(filepath.Join(root, filepath.FromSlash(relative)), MaxBrief)
	if err != nil {
		return domain.ChangeBrief{}, err
	}
	brief, err := ParseBrief(data)
	if err != nil {
		return domain.ChangeBrief{}, err
	}
	if brief.Path != relative {
		return domain.ChangeBrief{}, errors.New("change brief path does not match its change ID")
	}
	return brief, nil
}

func RenderBrief(brief domain.ChangeBrief) ([]byte, error) {
	if err := ValidateBrief(brief); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	fmt.Fprintf(&output, "# Level 7 change — %s\n\n", brief.ID)
	output.WriteString("| Field | Value |\n|---|---|\n")
	fmt.Fprintf(&output, "| Change ID | `%s` |\n", brief.ID)
	fmt.Fprintf(&output, "| Risk tier | `%d` |\n", brief.Tier)
	fmt.Fprintf(&output, "| Base commit | `%s` |\n", brief.Base)
	fmt.Fprintf(&output, "\n## Problem\n\n%s\n", brief.Problem)
	writeCodeList(&output, "Scope", brief.Scope)
	writeTextList(&output, "Acceptance criteria", brief.AcceptanceCriteria)
	writeTextList(&output, "Risks and mitigations", brief.Risks)
	writeTextList(&output, "Rollback", brief.Rollback)
	if output.Len() > MaxBrief {
		return nil, errors.New("rendered change brief exceeds size limit")
	}
	return output.Bytes(), nil
}

func ParseBrief(data []byte) (domain.ChangeBrief, error) {
	if len(data) == 0 || len(data) > MaxBrief || !utf8.Valid(data) || bytes.ContainsRune(data, '\r') || data[len(data)-1] != '\n' {
		return domain.ChangeBrief{}, errors.New("change brief encoding or size is invalid")
	}
	lines := strings.Split(string(data[:len(data)-1]), "\n")
	headings := []string{"## Problem", "## Scope", "## Acceptance criteria", "## Risks and mitigations", "## Rollback"}
	positions := make([]int, len(headings))
	previous := -1
	for index, heading := range headings {
		position := uniqueLine(lines, heading)
		if position < 0 || position <= previous {
			return domain.ChangeBrief{}, fmt.Errorf("missing, duplicate, or misplaced heading %q", heading)
		}
		positions[index] = position
		previous = position
	}
	if positions[0] != 8 || len(lines) < 18 {
		return domain.ChangeBrief{}, errors.New("change brief header layout is invalid")
	}
	if lines[1] != "" || lines[2] != "| Field | Value |" || lines[3] != "|---|---|" || lines[7] != "" {
		return domain.ChangeBrief{}, errors.New("change brief metadata table is invalid")
	}
	fields := make(map[string]string)
	for _, line := range lines[4:7] {
		parts := strings.Split(line, "|")
		if len(parts) != 4 || strings.TrimSpace(parts[0]) != "" || strings.TrimSpace(parts[3]) != "" {
			return domain.ChangeBrief{}, errors.New("change brief metadata row is invalid")
		}
		key := strings.TrimSpace(parts[1])
		rawValue := strings.TrimSpace(parts[2])
		if len(rawValue) < 3 || rawValue[0] != '`' || rawValue[len(rawValue)-1] != '`' || strings.Count(rawValue, "`") != 2 {
			return domain.ChangeBrief{}, errors.New("change brief metadata value must use one code span")
		}
		value := rawValue[1 : len(rawValue)-1]
		if fields[key] != "" {
			return domain.ChangeBrief{}, fmt.Errorf("duplicate change brief field %q", key)
		}
		fields[key] = value
	}
	if len(fields) != 3 || fields["Change ID"] == "" || fields["Risk tier"] == "" || fields["Base commit"] == "" {
		return domain.ChangeBrief{}, errors.New("change brief metadata fields are invalid")
	}
	tierValue, err := strconv.Atoi(fields["Risk tier"])
	if err != nil {
		return domain.ChangeBrief{}, errors.New("change brief risk tier is invalid")
	}
	brief := domain.ChangeBrief{
		ID:   fields["Change ID"],
		Tier: domain.RiskTier(tierValue),
		Base: fields["Base commit"],
	}
	brief.Path = BriefPath(brief.ID)
	if lines[0] != "# Level 7 change — "+brief.ID {
		return domain.ChangeBrief{}, errors.New("change brief title does not match its change ID")
	}
	sections := make([][]string, len(headings))
	for index, position := range positions {
		end := len(lines)
		if index+1 < len(positions) {
			end = positions[index+1]
		}
		section, err := trimmedSection(lines[position+1 : end])
		if err != nil {
			return domain.ChangeBrief{}, fmt.Errorf("section %q: %w", headings[index], err)
		}
		sections[index] = section
	}
	if len(sections[0]) != 1 {
		return domain.ChangeBrief{}, errors.New("problem must contain exactly one line")
	}
	brief.Problem = sections[0][0]
	brief.Scope, err = parseCodeList(sections[1])
	if err != nil {
		return domain.ChangeBrief{}, err
	}
	brief.AcceptanceCriteria, err = parseTextList(sections[2])
	if err != nil {
		return domain.ChangeBrief{}, err
	}
	brief.Risks, err = parseTextList(sections[3])
	if err != nil {
		return domain.ChangeBrief{}, err
	}
	brief.Rollback, err = parseTextList(sections[4])
	if err != nil {
		return domain.ChangeBrief{}, err
	}
	if err := ValidateBrief(brief); err != nil {
		return domain.ChangeBrief{}, err
	}
	return brief, nil
}

func ValidateBrief(brief domain.ChangeBrief) error {
	if !changeIDPattern.MatchString(brief.ID) || len(brief.ID) > 64 {
		return errors.New("change ID must be 1..64 lowercase letters, digits, or interior hyphens")
	}
	if brief.Tier != domain.TierProduct && brief.Tier != domain.TierHighRisk {
		return errors.New("tracked change brief requires risk tier 2 or 3")
	}
	if !gitIDPattern.MatchString(brief.Base) {
		return errors.New("base commit must be a full lowercase Git object ID")
	}
	if brief.Path != BriefPath(brief.ID) {
		return errors.New("change brief path does not match its change ID")
	}
	if !safeBriefLine(brief.Problem) {
		return errors.New("problem must be one bounded non-empty line")
	}
	if len(brief.Scope) < 1 || len(brief.Scope) > 256 {
		return errors.New("scope must contain 1..256 paths")
	}
	seen := make(map[string]bool)
	for _, scoped := range brief.Scope {
		if err := ValidateRepositoryPath(scoped); err != nil {
			return fmt.Errorf("invalid scope path %q: %w", scoped, err)
		}
		if seen[scoped] {
			return fmt.Errorf("duplicate scope path %q", scoped)
		}
		seen[scoped] = true
	}
	for label, values := range map[string][]string{
		"acceptance criteria": brief.AcceptanceCriteria,
		"risks":               brief.Risks,
		"rollback":            brief.Rollback,
	} {
		if len(values) < 1 || len(values) > 64 {
			return fmt.Errorf("%s must contain 1..64 values", label)
		}
		for _, value := range values {
			if !safeBriefLine(value) {
				return fmt.Errorf("%s contains an invalid line", label)
			}
		}
	}
	return nil
}

func writeCodeList(output *bytes.Buffer, heading string, values []string) {
	fmt.Fprintf(output, "\n## %s\n\n", heading)
	for _, value := range values {
		fmt.Fprintf(output, "- `%s`\n", value)
	}
}

func writeTextList(output *bytes.Buffer, heading string, values []string) {
	fmt.Fprintf(output, "\n## %s\n\n", heading)
	for _, value := range values {
		fmt.Fprintf(output, "- %s\n", value)
	}
}

func uniqueLine(lines []string, target string) int {
	position := -1
	for index, line := range lines {
		if line != target {
			continue
		}
		if position >= 0 {
			return -1
		}
		position = index
	}
	return position
}

func trimmedSection(lines []string) ([]string, error) {
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return nil, errors.New("section is empty")
	}
	for _, line := range lines {
		if line == "" {
			return nil, errors.New("section contains an unexpected blank line")
		}
	}
	return lines, nil
}

func parseCodeList(lines []string) ([]string, error) {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "- `") || !strings.HasSuffix(line, "`") || len(line) <= 4 {
			return nil, errors.New("scope must use one backtick-delimited path per bullet")
		}
		result = append(result, line[3:len(line)-1])
	}
	return result, nil
}

func parseTextList(lines []string) ([]string, error) {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "- ") || len(line) <= 2 {
			return nil, errors.New("section must use one non-empty value per bullet")
		}
		result = append(result, line[2:])
	}
	return result, nil
}

func safeBriefLine(value string) bool {
	if !safeText(value, 2048, false) || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character == '\n' || character == '\r' || character == 0x7f || (character < 0x20 && character != '\t') {
			return false
		}
	}
	return true
}
