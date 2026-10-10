package pluginrelease

import (
	"errors"
	"strconv"
	"strings"
)

// Version is a strict semantic version: MAJOR.MINOR.PATCH with an optional
// dot-separated prerelease suffix. Build metadata and a leading "v" are
// rejected so that one release maps to exactly one spelling.
type Version struct {
	Major, Minor, Patch int
	Pre                 []string
}

func ParseVersion(value string) (Version, error) {
	if value == "" || len(value) > 64 || strings.ContainsAny(value, " \t\r\n\x00+") {
		return Version{}, errors.New("version is empty, oversized, or contains unsupported characters")
	}
	core, pre, hasPre := strings.Cut(value, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, errors.New("version must be MAJOR.MINOR.PATCH")
	}
	numbers := make([]int, 3)
	for index, part := range parts {
		number, err := versionNumber(part)
		if err != nil {
			return Version{}, err
		}
		numbers[index] = number
	}
	version := Version{Major: numbers[0], Minor: numbers[1], Patch: numbers[2]}
	if hasPre {
		if pre == "" {
			return Version{}, errors.New("prerelease suffix is empty")
		}
		for _, identifier := range strings.Split(pre, ".") {
			if identifier == "" || !versionIdentifier(identifier) {
				return Version{}, errors.New("prerelease identifier is invalid")
			}
			if numeric(identifier) && len(identifier) > 1 && identifier[0] == '0' {
				return Version{}, errors.New("numeric prerelease identifier has a leading zero")
			}
			version.Pre = append(version.Pre, identifier)
		}
	}
	return version, nil
}

func (version Version) String() string {
	value := strconv.Itoa(version.Major) + "." + strconv.Itoa(version.Minor) + "." + strconv.Itoa(version.Patch)
	if len(version.Pre) > 0 {
		value += "-" + strings.Join(version.Pre, ".")
	}
	return value
}

func (version Version) Prerelease() bool { return len(version.Pre) > 0 }

// Compare orders versions per semantic versioning: a prerelease sorts before
// its release, numeric identifiers compare numerically, and a shorter
// identifier list sorts first when all shared identifiers are equal.
func (version Version) Compare(other Version) int {
	for _, pair := range [][2]int{{version.Major, other.Major}, {version.Minor, other.Minor}, {version.Patch, other.Patch}} {
		if pair[0] != pair[1] {
			return sign(pair[0] - pair[1])
		}
	}
	switch {
	case len(version.Pre) == 0 && len(other.Pre) == 0:
		return 0
	case len(version.Pre) == 0:
		return 1
	case len(other.Pre) == 0:
		return -1
	}
	for index := 0; index < len(version.Pre) && index < len(other.Pre); index++ {
		left, right := version.Pre[index], other.Pre[index]
		if left == right {
			continue
		}
		leftNumeric, rightNumeric := numeric(left), numeric(right)
		switch {
		case leftNumeric && rightNumeric:
			leftValue, _ := strconv.Atoi(left)
			rightValue, _ := strconv.Atoi(right)
			return sign(leftValue - rightValue)
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		case left < right:
			return -1
		default:
			return 1
		}
	}
	return sign(len(version.Pre) - len(other.Pre))
}

func versionNumber(part string) (int, error) {
	if part == "" || !numeric(part) || (len(part) > 1 && part[0] == '0') || len(part) > 9 {
		return 0, errors.New("version number is invalid")
	}
	return strconv.Atoi(part)
}

func versionIdentifier(value string) bool {
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && character != '-' {
			return false
		}
	}
	return true
}

func numeric(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}
