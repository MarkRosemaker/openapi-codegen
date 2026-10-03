// This file is written by hand, not by the generator.
//
// An error response type is passed to api.NewErrCustom, which takes an error,
// so the package author gives it an Error method. What a good message looks
// like depends on the API, which is why the generator does not guess.

package notionofficial

func errorMessage[C ~string](code C, message string) string {
	if message == "" {
		return string(code)
	}

	return string(code) + ": " + message
}

func (e *ErrorAPI) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI403) Error() string {
	switch {
	case e.ErrorAPI403OneOf0 != nil:
		return errorMessage(e.ErrorAPI403OneOf0.Code, e.ErrorAPI403OneOf0.Message)
	case e.ErrorAPI403OneOf1 != nil:
		return errorMessage(e.ErrorAPI403OneOf1.Code, e.ErrorAPI403OneOf1.Message)
	case e.ErrorAPI403OneOf2 != nil:
		return errorMessage(e.ErrorAPI403OneOf2.Code, e.ErrorAPI403OneOf2.Message)
	default:
		return "forbidden"
	}
}

func (e *ErrorAPI404) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI406) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI409) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI429) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI503) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI504) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorAPI529) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorOauth400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorOauth401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *ErrorOauth403) Error() string { return errorMessage(e.Code, e.Message) }
