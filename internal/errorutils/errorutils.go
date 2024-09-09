package errorutils

import "strings"

// JoinErrors joins multiple errors into a single pretty-formatted string.
func JoinErrors[T error](b *strings.Builder, errs []T, indent string) {
	for i, err := range errs {
		buildErrorMessage(b, err.Error(), indent)
		if i < len(errs)-1 {
			b.WriteString("\n")
		}
	}
}

const listPoint = "- "

func buildErrorMessage(b *strings.Builder, errMsg, indent string) {
	b.WriteString(indent)
	if !strings.HasPrefix(errMsg, listPoint) {
		b.WriteString(listPoint)
	}
	// Indent the whole error message.
	errMsg = strings.ReplaceAll(errMsg, "\n", "\n"+indent)
	b.WriteString(errMsg)
}
