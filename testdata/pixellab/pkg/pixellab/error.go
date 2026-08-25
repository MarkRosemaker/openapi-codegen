// This file is written by hand, not by the generator.
//
// An error response type is passed to api.NewErrCustom, which takes an error,
// so the package author gives it an Error method. What a good message looks
// like depends on the API, which is why the generator does not guess.

package pixellab

import "strings"

func (e *HTTPValidationError) Error() string {
	if len(e.Detail) == 0 {
		return "validation failed"
	}

	msgs := make([]string, 0, len(e.Detail))
	for _, d := range e.Detail {
		msgs = append(msgs, d.Msg)
	}

	return "validation failed: " + strings.Join(msgs, "; ")
}
