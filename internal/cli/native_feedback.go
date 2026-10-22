package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// Only the exact native artifact/validator boundary may mark author-owned work.
// Unmarked failures remain Main's responsibility; no error-text guessing.
type nativeAuthorError struct{ error }

func (e *nativeAuthorError) Unwrap() error { return e.error }

func nativeArtifactReadError(label string, err error) error {
	wrapped := fmt.Errorf("%s: %w", label, err)
	if os.IsNotExist(err) {
		return &nativeAuthorError{wrapped}
	}
	return wrapped
}

func nativeStopFailure(stage string, active bool, err error, diagnostic io.Writer) int {
	var author *nativeAuthorError
	var fields *workpacket.AnswerError
	detail := err.Error()
	if errors.As(err, &fields) && len(fields.Fields) > 0 && fields.Fields[0].Repair != "" {
		detail += " " + fields.Fields[0].Repair
	}
	if !active && (errors.As(err, &author) || fields != nil) {
		fmt.Fprintln(diagnostic, stage+": correct your actual professional artifact or return a real blocker. "+detail)
		return 2
	}
	fmt.Fprintln(diagnostic, stage+" reception incomplete; Main owns recovery; preserve the original artifacts and receipt, inspect this failure before resuming the original responsibility: "+detail)
	return 0
}
