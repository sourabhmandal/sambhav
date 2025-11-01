package authboss

import (
	"encoding/base64"
	"regexp"
	"sambhav/pkg/env"
	"time"

	"github.com/aarondl/authboss/v3"
	_ "github.com/aarondl/authboss/v3/auth"
	"github.com/aarondl/authboss/v3/defaults"
	_ "github.com/aarondl/authboss/v3/logout"
	aboauth "github.com/aarondl/authboss/v3/oauth2"
	"github.com/aarondl/authboss/v3/otp/twofactor"
	"github.com/aarondl/authboss/v3/otp/twofactor/totp2fa"
	_ "github.com/aarondl/authboss/v3/recover"
	_ "github.com/aarondl/authboss/v3/register"
	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"

	abclientstate "github.com/aarondl/authboss-clientstate"

	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ab        = authboss.New()
	abstore   = NewMemStorer()
	schemaDec = schema.NewDecoder()

	sessionStore abclientstate.SessionStorer
	cookieStore  abclientstate.CookieStorer
)

const (
	sessionCookieName = "ab_blog"
)

func setupAuthboss() {
	ab.Config.Paths.RootURL = "http://localhost:3000"
	ab.Config.Storage.Server = abstore
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore
	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}

	// We render mail with the authboss-renderer but we use a LogMailer
	// which simply sends the e-mail to stdout.
	// ab.Config.Core.MailRenderer = abrenderer.NewEmail("/auth", "ab_views")

	// The preserve fields are things we don't want to
	// lose when we're doing user registration (prevents having
	// to type them again)
	ab.Config.Modules.RegisterPreserveFields = []string{"email", "name"}

	// TOTP2FAIssuer is the name of the issuer we use for totp 2fa
	ab.Config.Modules.TOTP2FAIssuer = "sambhav-app"
	ab.Config.Modules.ResponseOnUnauthed = authboss.RespondRedirect

	// Turn on e-mail authentication required
	ab.Config.Modules.TwoFactorEmailAuthRequired = true

	// This instantiates and uses every default implementation
	// in the Config.Core area that exist in the defaults package.
	// Just a convenient helper if you don't want to do anything fancy.
	defaults.SetCore(&ab.Config, true, false)

	// Here we initialize the bodyreader as something customized in order to accept a name
	// parameter for our user as well as the standard e-mail and password.
	//
	// We also change the validation for these fields
	// to be something less secure so that we can use test data easier.
	emailRule := defaults.Rules{
		FieldName: "email", Required: true,
		MatchError: "Must be a valid e-mail address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]+`),
		MaxLength:  100,
	}
	passwordRule := defaults.Rules{
		FieldName: "password", Required: true,
		MinLength: 4,
		MaxLength: 100,
	}
	nameRule := defaults.Rules{
		FieldName: "name", Required: true,
		MinLength: 2,
		MaxLength: 100,
	}

	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: true,
		Rulesets: map[string][]defaults.Rules{
			"register":    {emailRule, passwordRule, nameRule},
			"recover_end": {passwordRule},
		},
		Confirms: map[string][]string{
			"register":    {"password", authboss.ConfirmPrefix + "password"},
			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		},
		Whitelist: map[string][]string{
			"register": {"email", "name", "password"},
		},
	}

	// Set up 2fa
	twofaRecovery := &twofactor.Recovery{Authboss: ab}
	if err := twofaRecovery.Setup(); err != nil {
		panic(err)
	}

	totp := &totp2fa.TOTP{Authboss: ab}
	if err := totp.Setup(); err != nil {
		panic(err)
	}

	// Set up Google OAuth2 if we have credentials in the
	// file oauth2.toml for it.
	env, _ := env.EnvVars()
	ab.Config.Modules.OAuth2Providers = map[string]authboss.OAuth2Provider{
		"google": {
			OAuth2Config: &oauth2.Config{
				ClientID:     env.GoogleClientID,
				ClientSecret: env.GoogleClientSecret,
				Scopes:       []string{`profile`, `email`},
				Endpoint:     google.Endpoint,
			},
			FindUserDetails: aboauth.GoogleUserDetails,
		},
	}

	// Initialize authboss (instantiate modules etc.)
	if err := ab.Init(); err != nil {
		panic(err)
	}
}

func setupAuth() {
	cookieStoreKey, _ := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)))
	sessionStoreKey, _ := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)))
	cookieStore = abclientstate.NewCookieStorer(cookieStoreKey, nil)
	cookieStore.HTTPOnly = false
	cookieStore.Secure = false
	sessionStore = abclientstate.NewSessionStorer(sessionCookieName, sessionStoreKey, nil)
	cstore := sessionStore.Store.(*sessions.CookieStore)
	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false
	cstore.MaxAge(int((30 * 24 * time.Hour) / time.Second))

	setupAuthboss()
	schemaDec.IgnoreUnknownKeys(true)

}

// Router exposes the internal authboss router so other packages can proxy
// requests into authboss (useful for JSON/api-based flows).
func Router() http.Handler {
	return ab.Config.Core.Router
}

func init() {
	// Initialize authboss on package load so Router() is ready to use.
	setupAuth()
}
