package version

import (
	"fmt"
	"strconv"
	"unicode"
)

func GenerateGuesses(ver string, evenOddPart int) []string {
	if ver == "" {
		return nil
	}

	// Split into numeric and non-numeric parts
	var parts []string
	start := 0
	isNum := unicode.IsDigit(rune(ver[0]))

	for i, r := range ver {
		if unicode.IsDigit(r) != isNum {
			parts = append(parts, ver[start:i])
			start = i
			isNum = !isNum
		}
	}
	parts = append(parts, ver[start:])

	var guesses []string
	numIndex := 0

	for i, part := range parts {
		if isNumeric(part) {
			v := atoi(part)
			if evenOddPart >= 0 && evenOddPart == numIndex {
				v += 2
			} else {
				v++
			}

			// Build guess string
			guess := ""
			for j := 0; j < i; j++ {
				guess += parts[j]
			}
			guess += fmt.Sprintf("%d", v)

			// Append trailing parts
			for j := i + 1; j < len(parts); j++ {
				if isNumeric(parts[j]) {
					guess += zeros(len(parts[j]))
				} else if isAlpha(parts[j]) {
					break
				} else {
					guess += parts[j]
				}
			}

			guesses = append(guesses, guess)
			numIndex++
		}
	}

	return guesses
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func zeros(n int) string {
	return fmt.Sprintf("%0*s", n, "")
}

func isAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}
	return true
}
