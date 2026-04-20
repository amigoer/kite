package router

import (
	"html/template"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// publicUser is the subset of user fields the landing-page templates need to
// render a login-aware header. It is populated from the access_token cookie
// when present and is always safe to render as nil.
type publicUser struct {
	ID          string
	Username    string
	DisplayName string
	AvatarURL   string
	Initial     string
	Role        string
	IsAdmin     bool
}

// registerLanding wires the server-rendered landing pages (/, /explore,
// /upload) and parses the embedded template set. These pages recognise the
// access_token cookie so the header reflects login state, but they never
// require authentication.
func registerLanding(r *gin.Engine, cfg Config, userRepo *repo.UserRepo) {
	if cfg.TemplateFS != nil {
		if tmpl, err := template.ParseFS(cfg.TemplateFS, "layouts/*.html", "pages/*.html"); err == nil {
			r.SetHTMLTemplate(tmpl)
		}
	}

	settingRepo := repo.NewSettingRepo(cfg.DB)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"CurrentUser": getOptionalUser(c, cfg.AuthSvc, userRepo),
			"ActiveNav":   "",
		})
	})

	r.GET("/explore", func(c *gin.Context) {
		val, _ := settingRepo.Get(c.Request.Context(), "allow_public_gallery")
		c.HTML(http.StatusOK, "explore.html", gin.H{
			"GalleryEnabled": val == "true",
			"CurrentUser":    getOptionalUser(c, cfg.AuthSvc, userRepo),
			"ActiveNav":      "explore",
		})
	})

	r.GET("/upload", func(c *gin.Context) {
		val, _ := settingRepo.Get(c.Request.Context(), "allow_guest_upload")
		c.HTML(http.StatusOK, "upload.html", gin.H{
			"GuestUploadEnabled": val == "true",
			"CurrentUser":        getOptionalUser(c, cfg.AuthSvc, userRepo),
			"ActiveNav":          "upload",
		})
	})
}

// getOptionalUser decodes the access_token cookie (if any) into a publicUser
// view used by landing templates. Returns nil for anonymous visitors or when
// the token is invalid or expired. Never returns an error since every failure
// mode degrades to "not logged in".
func getOptionalUser(c *gin.Context, authSvc *service.AuthService, userRepo *repo.UserRepo) *publicUser {
	cookie, err := c.Cookie("access_token")
	if err != nil || cookie == "" {
		return nil
	}
	claims, err := authSvc.ValidateToken(cookie)
	if err != nil {
		return nil
	}

	view := &publicUser{
		ID:          claims.UserID,
		Username:    claims.Username,
		DisplayName: claims.Username,
		Role:        claims.Role,
		IsAdmin:     claims.Role == "admin",
	}

	if user, getErr := userRepo.GetByID(c.Request.Context(), claims.UserID); getErr == nil {
		view.Username = user.Username
		if user.Nickname != nil {
			if nickname := strings.TrimSpace(*user.Nickname); nickname != "" {
				view.DisplayName = nickname
			}
		}
		if user.AvatarURL != nil {
			view.AvatarURL = strings.TrimSpace(*user.AvatarURL)
		}
	}

	view.Initial = nameInitial(view.Username)
	return view
}

// nameInitial returns the uppercase first rune of name, or "U" when name is
// empty or decoding fails. Used as the avatar fallback glyph.
func nameInitial(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "U"
	}
	r, _ := utf8.DecodeRuneInString(trimmed)
	if r == utf8.RuneError {
		return "U"
	}
	return strings.ToUpper(string(r))
}
