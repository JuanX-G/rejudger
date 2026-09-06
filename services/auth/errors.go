package auth

// Possible error types that actions relating to the auth manager can return
type AuthErrorType int

const (
	AuthErrorNoSuchSession        AuthErrorType = iota // Returned if no session exists for a given token
	AuthErrorSessionAlreadyExists                      // Returned upon an attempt to create a session with the same token as an existing session
	AuthErrorExpiredSession                            // Returned if the session time out
	AuthErrorForbidden                                 // Returned if the session does not have the required permissions
	UnknownAuthError
)

type AuthError struct {
	errType       AuthErrorType
	customMessage string
}

func (a AuthError) Error() string {
	switch a.errType {
	case AuthErrorNoSuchSession:
		return "no session with the given token found"
	case AuthErrorSessionAlreadyExists:
		return "session with this token already exists"
	case AuthErrorExpiredSession:
		return "session has expired"
	case AuthErrorForbidden:
		return "session lack necessary permissions"
	case UnknownAuthError:
		fallthrough
	default:
		return "unknown auth error"
	}
}

func (a AuthError) Is(target error) bool {
	if other, ok := target.(AuthError); ok {
		return a.errType == other.errType
	}
	return false
}

var ErrorNoSuchSession = AuthError{errType: AuthErrorNoSuchSession}
