// This file is written by hand, not by the generator.
//
// An error response type is passed to api.NewErrCustom, which takes an error,
// so the package author gives it an Error method. What a good message looks
// like depends on the API, which is why the generator does not guess.

package habitica

import "fmt"

func (e *PostApiv3TaskScoreUpNotFound) Error() string {
	if e.Message == "" {
		return e.Err
	}

	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}
