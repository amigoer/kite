package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kite-plus/kite/internal/model"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrTOTPAlreadyEnabled blocks re-enrollment while 2FA is already on.
	// Callers should disable first, then re-enroll.
	ErrTOTPAlreadyEnabled = errors.New("two-factor authentication is already enabled")
	// ErrTOTPNotEnrolled signals that /2fa/enable was called before
	// /2fa/setup — i.e. there's no pending secret to confirm. Typical
	// cause: the client lost state and replayed an old enable request.
	ErrTOTPNotEnrolled = errors.New("two-factor authentication enrollment not started")
	// ErrTOTPInvalidCode covers any failed verification path.
	ErrTOTPInvalidCode = errors.New("invalid verification code")
	// ErrTOTPNotEnabled is returned when disable/verify is called on an
	// account that never finished enrollment.
	ErrTOTPNotEnabled = errors.New("two-factor authentication is not enabled")
	// ErrChallengeInvalid fires when the login challenge token cannot be
	// parsed, is expired, or targets a user whose 2FA state has since
	// changed (enable/disable bumps token_version).
	ErrChallengeInvalid = errors.New("invalid or expired two-factor challenge")
)

// totpChallengePurpose is the `purpose` claim value that distinguishes
// a 2FA challenge token from a regular access/refresh JWT. A regular
// token accidentally (or maliciously) posted to /2fa/verify must fail
// parsing; a challenge token posted to any other endpoint must likewise
// fail. Checking purpose on both sides isolates the blast radius of a
// stolen short-lived 2FA ticket.
const totpChallengePurpose = "2fa_challenge"

// totpChallengeTTL is how long a user has to enter their code after
// passing the password step. Long enough to pick up the phone, short
// enough that a leaked challenge token isn't useful.
const totpChallengeTTL = 5 * time.Minute

// totpIssuer is the label shown inside the authenticator app next to
// the code. Keeping it static avoids leaking the deployment hostname
// into the user's device.
const totpIssuer = "Kite"

// TOTPChallengeClaims is the JWT we hand back when Login succeeds on
// the password factor but the user still needs to present their TOTP
// code. It intentionally carries no Role or PasswordMustChange — those
// belong on the final token pair, after both factors pass.
type TOTPChallengeClaims struct {
	UserID       string `json:"user_id"`
	Purpose      string `json:"purpose"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

// TOTPEnrollment is the ephemeral payload the client needs to render
// the QR code: the otpauth:// URI (for the QR image) plus the raw
// Base32 secret (shown as "manual entry" fallback). The server keeps
// the same secret pending on the user row and only promotes it to
// "enabled" once the user confirms with /2fa/enable.
type TOTPEnrollment struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

// BeginTOTPEnrollment generates a fresh TOTP secret, stashes it on the
// user row (but leaves totp_enabled=false), and returns the otpauth
// URI for the client to render as a QR code. Calling this on a user
// who already has 2FA enabled is an error — clients should disable
// first to avoid silently overwriting a working secret.
func (s *AuthService) BeginTOTPEnrollment(ctx context.Context, userID string) (*TOTPEnrollment, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.TOTPEnabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: user.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp: %w", err)
	}
	secret := key.Secret()
	user.TOTPSecret = &secret
	user.TOTPEnabled = false
	user.TOTPConfirmedAt = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("persist totp secret: %w", err)
	}

	return &TOTPEnrollment{Secret: secret, URI: key.URL()}, nil
}

// ConfirmTOTPEnrollment validates the user's first code against the
// pending secret and, on success, flips totp_enabled to true and bumps
// token_version so any session predating the security change is
// forcibly rotated.
func (s *AuthService) ConfirmTOTPEnrollment(ctx context.Context, userID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user.TOTPEnabled {
		return ErrTOTPAlreadyEnabled
	}
	if user.TOTPSecret == nil || *user.TOTPSecret == "" {
		return ErrTOTPNotEnrolled
	}
	if !validateTOTPCode(*user.TOTPSecret, code) {
		return ErrTOTPInvalidCode
	}

	now := time.Now()
	user.TOTPEnabled = true
	user.TOTPConfirmedAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("enable totp: %w", err)
	}
	if err := s.userRepo.BumpTokenVersion(ctx, userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}

// DisableTOTP turns 2FA off. We require BOTH the account password and
// a current TOTP code: the password alone is not enough (an attacker
// with a stolen session + password could strip the extra factor), and
// a code alone is not enough (someone with temporary access to the
// phone shouldn't be able to permanently remove the lock). Admin
// disablement for locked-out users is a separate code path.
func (s *AuthService) DisableTOTP(ctx context.Context, userID, password, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if !user.TOTPEnabled || user.TOTPSecret == nil {
		return ErrTOTPNotEnabled
	}
	if !user.HasLocalPassword {
		return ErrLocalPasswordNotSet
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrPasswordMismatch
	}
	if !validateTOTPCode(*user.TOTPSecret, code) {
		return ErrTOTPInvalidCode
	}

	user.TOTPEnabled = false
	user.TOTPSecret = nil
	user.TOTPConfirmedAt = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("disable totp: %w", err)
	}
	if err := s.userRepo.BumpTokenVersion(ctx, userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}

// AdminResetTOTP lets an administrator strip a user's 2FA when the
// legitimate owner has lost access to their authenticator. The caller
// is responsible for having authenticated as an admin; this method
// only handles the state transition. Token version is bumped so any
// existing sessions on the victim account (including one an attacker
// might hold) are invalidated — otherwise a reset could be abused to
// silently weaken a hijacked account.
func (s *AuthService) AdminResetTOTP(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if !user.TOTPEnabled && user.TOTPSecret == nil {
		return ErrTOTPNotEnabled
	}
	user.TOTPEnabled = false
	user.TOTPSecret = nil
	user.TOTPConfirmedAt = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("reset totp: %w", err)
	}
	if err := s.userRepo.BumpTokenVersion(ctx, userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}

// LoginResult is the tagged return of LoginOrChallenge. Exactly one
// of Tokens or Challenge is non-nil. Callers switch on presence to
// decide whether to finish the session or prompt for a second factor.
type LoginResult struct {
	Tokens    *TokenPair
	Challenge *TOTPChallengeResult
}

// TOTPChallengeResult carries the challenge token and its wall-clock
// expiry so the UI can show "enter your code — expires in 4:37".
type TOTPChallengeResult struct {
	Token     string    `json:"challenge_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginOrChallenge is the 2FA-aware login entry point. It performs
// the same password-factor verification as Login, but if the user has
// TOTP enabled it returns a challenge token instead of a token pair —
// the caller must then exchange the challenge for a token pair via
// ConsumeTOTPChallenge. Login (without challenge handling) is kept
// around for tests and internal flows that never see 2FA.
func (s *AuthService) LoginOrChallenge(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		user, err = s.userRepo.GetByEmail(ctx, username)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	if !user.HasLocalPassword {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.TOTPEnabled {
		token, expiry, err := s.IssueTOTPChallenge(user)
		if err != nil {
			return nil, err
		}
		return &LoginResult{Challenge: &TOTPChallengeResult{Token: token, ExpiresAt: expiry}}, nil
	}
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Tokens: tokens}, nil
}

// IssueTOTPChallenge mints the short-lived challenge token returned by
// the password step of Login when the user has 2FA enabled. The token
// encodes the user id + the current token_version, so that any
// credential-altering event between the two login halves (password
// reset, 2FA disable, admin reset) invalidates this challenge.
func (s *AuthService) IssueTOTPChallenge(user *model.User) (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(totpChallengeTTL)
	claims := &TOTPChallengeClaims{
		UserID:       user.ID,
		Purpose:      totpChallengePurpose,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign challenge: %w", err)
	}
	return signed, expiry, nil
}

// ConsumeTOTPChallenge validates a challenge token + code and, on
// success, issues a real token pair. The challenge is not stateful —
// a given JWT can theoretically be replayed inside its 5-minute
// window, but each replay still has to present a fresh 6-digit TOTP
// code, and once the user finishes login the first time every
// subsequent code is the only barrier. We accept that tradeoff to
// avoid a "pending challenges" table.
func (s *AuthService) ConsumeTOTPChallenge(ctx context.Context, challengeToken, code string) (*TokenPair, error) {
	claims, err := s.parseChallengeToken(challengeToken)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrChallengeInvalid
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	// Challenge token was signed against a specific token_version. If
	// the user has since changed their password or toggled 2FA, reject
	// the stale challenge instead of letting it promote to a full
	// session on an account whose security state moved on.
	if claims.TokenVersion < user.TokenVersion {
		return nil, ErrChallengeInvalid
	}
	if !user.TOTPEnabled || user.TOTPSecret == nil {
		return nil, ErrTOTPNotEnabled
	}
	if !validateTOTPCode(*user.TOTPSecret, code) {
		return nil, ErrTOTPInvalidCode
	}
	return s.generateTokenPair(user)
}

func (s *AuthService) parseChallengeToken(tokenStr string) (*TOTPChallengeClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TOTPChallengeClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrChallengeInvalid
	}
	claims, ok := token.Claims.(*TOTPChallengeClaims)
	if !ok || !token.Valid {
		return nil, ErrChallengeInvalid
	}
	if claims.Purpose != totpChallengePurpose {
		return nil, ErrChallengeInvalid
	}
	return claims, nil
}

// validateTOTPCode tolerates minor clock skew by accepting the code
// from the previous + next 30-second windows in addition to the
// current one (pquerna/otp's default skew is 0, which rejects anything
// more than a second late). Digits-only normalisation lets users paste
// codes that include spaces or punctuation.
func validateTOTPCode(secret, code string) bool {
	cleaned := stripNonDigits(code)
	if len(cleaned) != 6 {
		return false
	}
	ok, err := totp.ValidateCustom(cleaned, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false
	}
	return ok
}

func stripNonDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
