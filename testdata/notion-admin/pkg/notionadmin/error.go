// This file is written by hand, not by the generator.
//
// An error response type is passed to api.NewErrCustom, which takes an error,
// so the package author gives it an Error method. What a good message looks
// like depends on the API, which is why the generator does not guess.

package notionadmin

func errorMessage[C ~string](code C, message string) string {
	if message == "" {
		return string(code)
	}

	return string(code) + ": " + message
}

func (e *AdminAPIError400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIError401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIError403) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIError404) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIError429) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIError500) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError403) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError404) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError409) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError429) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError500) Error() string { return errorMessage(e.Code, e.Message) }

func (e *AdminAPIPublicError503) Error() string { return errorMessage(e.Code, e.Message) }
