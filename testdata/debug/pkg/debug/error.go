// This file is written by hand, not by the generator.
//
// An error response type is passed to api.NewErrCustom, which takes an error, so the package author gives it an Error
// method.

package debug

func (e *Problem) Error() string { return e.Message }
