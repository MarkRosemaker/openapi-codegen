// This file is written by hand, not by the generator.
//
// An error response type is passed to api.NewErrCustom, which takes an error,
// so the package author gives it an Error method. What a good message looks
// like depends on the API, which is why the generator does not guess.

package notionundocumented

func errorMessage[C ~string](code C, message string) string {
	if message == "" {
		return string(code)
	}

	return string(code) + ": " + message
}

func (e *error_api_) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_403) Error() string {
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

func (e *error_api_404) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_406) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_409) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_429) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_503) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_504) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_api_529) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_oauth_400) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_oauth_401) Error() string { return errorMessage(e.Code, e.Message) }

func (e *error_oauth_403) Error() string { return errorMessage(e.Code, e.Message) }
