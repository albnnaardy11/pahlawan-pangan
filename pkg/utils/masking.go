package utils

import (
	"fmt"
	"strings"
)

// MaskPII masks sensitive information like emails and phone numbers for logging/display.
// Compliant with GDPR and local PDP regulations.
func MaskPII(input string) string {
	if input == "" {
		return ""
	}

	// Email masking: b***i@example.com
	if strings.Contains(input, "@") {
		parts := strings.Split(input, "@")
		if len(parts[0]) <= 2 {
			return input
		}
		return fmt.Sprintf("%c***%c@%s", parts[0][0], parts[0][len(parts[0])-1], parts[1])
	}

	// Phone masking: 0812****1234
	if len(input) > 8 {
		return fmt.Sprintf("%s****%s", input[:4], input[len(input)-4:])
	}

	return "****"
}
