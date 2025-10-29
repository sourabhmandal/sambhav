package authboss

import (
	"context"

	"sambhav/pkg/database"

	"github.com/aarondl/authboss/v3"
	"github.com/aarondl/authboss/v3/otp/twofactor/totp2fa"
	"github.com/friendsofgo/errors"
)

// MemStorer stores users in memory
type MemStorer struct {
	Users  map[string]database.User
	Tokens map[string][]string
}

var (
	assertUser   = &database.User{}
	assertStorer = &MemStorer{}

	_ authboss.User            = assertUser
	_ authboss.AuthableUser    = assertUser
	_ authboss.ConfirmableUser = assertUser
	_ authboss.LockableUser    = assertUser
	_ authboss.RecoverableUser = assertUser
	_ authboss.ArbitraryUser   = assertUser

	_ totp2fa.User = assertUser

	_ authboss.CreatingServerStorer    = assertStorer
	_ authboss.ConfirmingServerStorer  = assertStorer
	_ authboss.RecoveringServerStorer  = assertStorer
	_ authboss.RememberingServerStorer = assertStorer
)

// NewMemStorer constructor
func NewMemStorer() *MemStorer {
	return &MemStorer{
		Users: map[string]database.User{
			"rick@councilofricks.com": {
				ID:                 1,
				Name:               "Rick",
				Password:           "$2a$10$XtW/BrS5HeYIuOCXYe8DFuInetDMdaarMUJEOg/VA/JAIDgw3l4aG", // pass = 1234
				Email:              "rick@councilofricks.com",
				Confirmed:          true,
				SMSSeedPhoneNumber: "(777)-123-4567",
			},
		},
		Tokens: make(map[string][]string),
	}
}

// Save the user
func (m MemStorer) Save(_ context.Context, user authboss.User) error {
	u := user.(*database.User)
	m.Users[u.Email] = *u

	debugln("Saved user:", u.Name)
	return nil
}

// Load the user
func (m MemStorer) Load(_ context.Context, key string) (user authboss.User, err error) {
	// Check to see if our key is actually an oauth2 pid
	provider, uid, err := authboss.ParseOAuth2PID(key)
	if err == nil {
		for _, u := range m.Users {
			if u.OAuth2Provider == provider && u.OAuth2UID == uid {
				debugln("Loaded OAuth2 user:", u.Email)
				return &u, nil
			}
		}

		return nil, authboss.Errdatabase.UserNotFound
	}

	u, ok := m.database.Users[key]
	if !ok {
		return nil, authboss.Errdatabase.UserNotFound
	}

	debugln("Loaded user:", u.Name)
	return &u, nil
}

// New user creation
func (m MemStorer) New(_ context.Context) database.User {
	return &database.User{}
}

// Create the user
func (m MemStorer) Create(_ context.Context, user database.User) error {
	u := user.(*database.User)

	if _, ok := m.database.Users[u.Email]; ok {
		return authboss.Errdatabase.UserFound
	}

	debugln("Created new user:", u.Name)
	m.database.Users[u.Email] = *u
	return nil
}

// LoadByConfirmSelector looks a user up by confirmation token
func (m MemStorer) LoadByConfirmSelector(_ context.Context, selector string) (user authboss.ConfirmableUser, err error) {
	for _, v := range m.database.Users {
		if v.ConfirmSelector == selector {
			debugln("Loaded user by confirm selector:", selector, v.Name)
			return &v, nil
		}
	}

	return nil, authboss.Errdatabase.UserNotFound
}

// LoadByRecoverSelector looks a user up by confirmation selector
func (m MemStorer) LoadByRecoverSelector(_ context.Context, selector string) (user authboss.RecoverableUser, err error) {
	for _, v := range m.database.Users {
		if v.RecoverSelector == selector {
			debugln("Loaded user by recover selector:", selector, v.Name)
			return &v, nil
		}
	}

	return nil, authboss.Errdatabase.UserNotFound
}

// AddRememberToken to a user
func (m MemStorer) AddRememberToken(_ context.Context, pid, token string) error {
	m.Tokens[pid] = append(m.Tokens[pid], token)
	debugf("Adding rm token to %s: %s\n", pid, token)
	spew.Dump(m.Tokens)
	return nil
}

// DelRememberTokens removes all tokens for the given pid
func (m MemStorer) DelRememberTokens(_ context.Context, pid string) error {
	delete(m.Tokens, pid)
	debugln("Deleting rm tokens from:", pid)
	spew.Dump(m.Tokens)
	return nil
}

// UseRememberToken finds the pid-token pair and deletes it.
// If the token could not be found return ErrTokenNotFound
func (m MemStorer) UseRememberToken(_ context.Context, pid, token string) error {
	tokens, ok := m.Tokens[pid]
	if !ok {
		debugln("Failed to find rm tokens for:", pid)
		return authboss.ErrTokenNotFound
	}

	for i, tok := range tokens {
		if tok == token {
			tokens[len(tokens)-1] = tokens[i]
			m.Tokens[pid] = tokens[:len(tokens)-1]
			debugf("Used remember for %s: %s\n", pid, token)
			return nil
		}
	}

	return authboss.ErrTokenNotFound
}

// NewFromOAuth2 creates an oauth2 user (but not in the database, just a blank one to be saved later)
func (m MemStorer) NewFromOAuth2(_ context.Context, provider string, details map[string]string) (authboss.OAuth2User, error) {
	switch provider {
	case "google":
		email := details[aboauth.OAuth2Email]

		var user *database.User
		if u, ok := m.database.Users[email]; ok {
			user = &u
		} else {
			user = &database.User{}
		}

		// Google OAuth2 doesn't allow us to fetch real name without more complicated API calls
		// in order to do this properly in your own app, look at replacing the authboss oauth2.Googledatabase.UserDetails
		// method with something more thorough.
		user.Name = "Unknown"
		user.Email = details[aboauth.OAuth2Email]
		user.OAuth2UID = details[aboauth.OAuth2UID]
		user.Confirmed = true

		return user, nil
	}

	return nil, errors.Errorf("unknown provider %s", provider)
}

// SaveOAuth2 user
func (m MemStorer) SaveOAuth2(_ context.Context, user authboss.OAuth2User) error {
	u := user.(*database.User)
	m.database.Users[u.Email] = *u

	return nil
}
