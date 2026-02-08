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

	// Convert to runes for UTF-8 safety (important for names/emails with emojis/non-ASCII)
	runes := []rune(input)

	// Email masking: b***i@example.com
	if strings.Contains(input, "@") {
		parts := strings.Split(input, "@")
		localRunes := []rune(parts[0])
		if len(localRunes) <= 2 {
			return input
		}
		return fmt.Sprintf("%c***%c@%s", localRunes[0], localRunes[len(localRunes)-1], parts[1])
	}

	// Phone masking: 0812****1234
	if len(runes) > 8 {
		return fmt.Sprintf("%s****%s", string(runes[:4]), string(runes[len(runes)-4:]))
	}

	return "****"
}
