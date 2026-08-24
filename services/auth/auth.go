package auth

import (
	"revit/internal/customSync"
	"revit/internal/permissions"
	"sync"
	"time"
)

// Interface for auth manager.
type AuthManager interface {
	DeleteToken(token string) error
	NewSession(expiryDur int, userId int, userName string, perms permissions.PermissionSet) (string, error)
	UpdateSessionPerms(token string, perms permissions.PermissionSet) error
	HasPermission(token string, context string, required permissions.ActionPermission) error
	IsValid(token string) error
	HasAllPermissions(token string, context string, required ...permissions.ActionPermission) error
	GetUserId(token string) (int64, error)
}

// The AuthManager provides an API for managing authentication to the app. It is a light weight alternative to other forms of auth that run as a separate service.
// This could be used in production, but is also meant to be used for debugging and testing purposes.
// Because of frequest insertions it uses Mutex + Map instead of sync.Map.
// It is safe for any goroutines to access its fields and perform actions using the provided methods.
type AuthManager struct {
	mu       sync.RWMutex
	sessions map[string]*customSync.LockedValue[Session] //map a token (string) to session info
}

func NewAuthManager() *AuthManager {
	return &AuthManager{sessions: make(map[string]*customSync.LockedValue[Session])}
}

// getSessionLock is an internal function enabling the methods to lock the sessions map for shorter periods of time.
// It resolves the session for a specified token, returns a AuthError with type AuthErrorNoSuchSession
// if it fails to find a session for that token. It is safe for concurent calls.
func (a *AuthManager) getSessionLock(token string) (*customSync.LockedValue[Session], error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s, ok := a.sessions[token]
	if !ok {
		return nil, AuthError{errType: AuthErrorNoSuchSession}
	}

	return s, nil
}

func (a *AuthManager) DeleteToken(token string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, found := a.sessions[token]
	if !found {
		return AuthError{errType: AuthErrorNoSuchSession}
	}
	delete(a.sessions, token)
	return nil
}

// NewSession creates a new session behind a mutex, and adds it to the AuthManagers registry.
func (a *AuthManager) NewSession(expiryDur int, userId int, userName string, perms permissions.PermissionSet) (string, error) {
	var token string
	var err error
	a.mu.Lock()
	defer a.mu.Unlock()
	for {
		token, err = generateToken()
		if err != nil {
			return "", err
		}
		if _, found := a.sessions[token]; !found {
			break
		}
	}
	session := newSession(expiryDur, userId, userName, perms)
	sessionPtr := customSync.MakeLockedValue(session)

	a.sessions[token] = sessionPtr
	return token, nil
}

func (a *AuthManager) rotateToken(originalToken string) (string, error) {
	var newToken string
	var err error
	a.mu.Lock()
	defer a.mu.Unlock()
	session, found := a.sessions[originalToken]
	if !found {
		return "", AuthError{errType: AuthErrorNoSuchSession}
	}
	for {
		newToken, err = generateToken()
		if err != nil {
			return "", err
		}
		if _, found := a.sessions[newToken]; !found {
			break
		}
	}
	a.sessions[newToken] = session
	delete(a.sessions, originalToken)
	return newToken, nil
}

func (a *AuthManager) UpdateSessionPerms(token string, perms permissions.PermissionSet) error {
	sessionPtr, err := a.getSessionLock(token)
	if err != nil {
		return err
	}
	updateSessionPermsCallback := func(s *Session) {
		UpdateSessionPerms(s, perms)
	}
	sessionPtr.WithWritePointer(updateSessionPermsCallback)
	return nil
}

func (a *AuthManager) HasPermission(token string, context string, required permissions.ActionPermission) error {
	session, err := a.getSessionLock(token)
	if err != nil {
		return err
	}

	var verifyErr error
	session.WithReadPointer(func(s *Session) {
		if time.Now().After(s.Expiry) {
			verifyErr = AuthError{errType: AuthErrorExpiredSession}
			return
		}
		perms, ok := s.Permissions[context]
		if !ok {
			verifyErr = AuthError{errType: AuthErrorForbidden}
			return
		}
		if !permissions.HasPermission(perms, required) {
			verifyErr = AuthError{errType: AuthErrorForbidden}
			return
		}
		verifyErr = nil
	})

	return verifyErr
}

func (a *AuthManager) IsValid(token string) error {
	session, err := a.getSessionLock(token)
	if err != nil {
		return err
	}

	var verifyErr error
	session.WithReadPointer(func(s *Session) {
		if time.Now().After(s.Expiry) {
			verifyErr = AuthError{errType: AuthErrorExpiredSession}
			go a.DeleteToken(token)
			return
		}
	})

	return verifyErr
}

func (a *AuthManager) HasAllPermissions(token string, context string, required ...permissions.ActionPermission) error {
	session, err := a.getSessionLock(token)
	if err != nil {
		return err
	}

	var verifyErr error
	session.WithReadPointer(func(s *Session) {
		if time.Now().After(s.Expiry) {
			verifyErr = AuthError{errType: AuthErrorExpiredSession}
			return
		}
		perms, ok := s.Permissions[context]
		if !ok {
			verifyErr = AuthError{errType: AuthErrorForbidden}
			return
		}

		for _, req := range required {
			if !permissions.HasPermission(perms, req) {
				verifyErr = AuthError{errType: AuthErrorForbidden}
				return
			}
		}

		verifyErr = nil
	})

	return verifyErr
}

func (a *AuthManager) GetUserId(token string) (int64, error) {
	session, err := a.getSessionLock(token)
	if err != nil {
		return 0, err
	}
	var userId int64
	session.WithReadPointer(func(s *Session) {
		userId = int64(s.UserId)
	})
	return userId, nil
}
