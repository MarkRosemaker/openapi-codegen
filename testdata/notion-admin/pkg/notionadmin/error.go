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

func (e *adminApiError400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiError401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiError403) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiError404) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiError429) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiError500) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError403) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError404) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError409) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError429) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError500) Error() string { return errorMessage(e.Code, e.Message) }

func (e *adminApiPublicError503) Error() string { return errorMessage(e.Code, e.Message) }
