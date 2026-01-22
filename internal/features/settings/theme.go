package settings

import (
	"context"
	"html/template"
	"net/url"
	"path/filepath"
	"strings"

	"pin/internal/config"
	"pin/internal/domain"
)

const (
	defaultThemeName             = "classic"
	defaultCustomThemeName       = "default_custom_css"
	themeDefaultKey              = "theme_default"
	themeDefaultForceKey         = "theme_default_force"
	themeDefaultCustomCSSPathKey = "theme_default_custom_css_path"
	themeUserSelectKey           = "theme_user_select"
	themeUserCustomCSSKey        = "theme_user_custom_css"
	themeUploadsSubdirectory     = "themes"
)

const (
	DefaultThemeName       = defaultThemeName
	DefaultCustomThemeName = defaultCustomThemeName
)

// ThemeOption describes a built-in theme choice for the UI.
type ThemeOption struct {
	Name        string
	Label       string
	Description string
	Page        int
}

// ThemeSettings holds the active themes and custom CSS overrides.
type ThemeSettings struct {
	ProfileTheme      string
	AdminTheme        string
	CustomCSSPath     string
	CustomCSSURL      string
	InlineCSS         string
	InlineCSSTemplate template.CSS
}

// ThemePolicy controls whether non-admin users can customize appearance.
type ThemePolicy struct {
	AllowUserTheme     bool
	AllowUserCustomCSS bool
}

type Store interface {
	GetSettings(ctx context.Context, keys ...string) (map[string]string, error)
	GetSetting(ctx context.Context, key string) (string, bool, error)
	SetSetting(ctx context.Context, key, value string) error
	SetSettings(ctx context.Context, values map[string]string) error
	DeleteSetting(ctx context.Context, key string) error
	UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error
	GetOwnerUser(ctx context.Context) (domain.User, error)
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

var builtInThemes = []ThemeOption{
	{Name: "classic", Label: "Classic paper", Description: "Warm paper, serif body, muted orange highlights. Default look."},
	{Name: "noir", Label: "Noir studio", Description: "Deep midnight panels with electric cyan accents for a dramatic dark mode."},
	{Name: "mono", Label: "Terminal mono", Description: "Developer-focused, near-black canvas with neon green details and monospace type."},
	{Name: "sunrise", Label: "Sunrise pop", Description: "Bold modern gradient vibe with coral and amber accents."},
	{Name: "forest", Label: "Evergreen modern", Description: "Fresh botanical palette with crisp sans headings and soft shadows."},
	{Name: "slate", Label: "Slate serif", Description: "Cool graphite neutrals with lavender accents and transitional typography."},
	{Name: "desert", Label: "Desert bloom", Description: "Sandy neutrals with turquoise punches and a humanist serif/sans mix."},
	{Name: "ocean", Label: "Ocean depth", Description: "Deep navy gradients with seafoam accents and rounded sans headings."},
	{Name: "pastel", Label: "Pastel cloud", Description: "Playful candy palette with soft rounded cards and friendly fonts."},
	{Name: "brutalist", Label: "Brutalist mono", Description: "High-contrast black/white with bold yellow accents and rigid grids."},
	{Name: "neonpop", Label: "Neon pop", Description: "Black canvas, hyper-saturated magenta/cyan accents, bold geometric type."},
	{Name: "velvet", Label: "Velvet luxe", Description: "Deep burgundy and gold with elegant serif headings and soft glow."},
	{Name: "glass", Label: "Glass frost", Description: "Frosted glass blur vibe with icy blues and sleek sans typography."},
	{Name: "midcentury", Label: "Mid-century", Description: "Muted teals and mustards with rounded cards and retro sans serif."},
	{Name: "tech", Label: "Tech grid", Description: "Slate/teal gradient, neon lime highlights, condensed tech-forward font."},
}

func NormalizeThemeChoice(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return defaultThemeName
	}
	if name == defaultCustomThemeName {
		return name
	}
	for _, opt := range builtInThemes {
		if opt.Name == name {
			return name
		}
	}
	return defaultThemeName
}

func ThemeDir(cfg config.Config) string {
	return filepath.Join(cfg.UploadsDir, themeUploadsSubdirectory)
}

func ThemeCustomCSSURL(path string) string {
	base := filepath.Base(strings.TrimSpace(path))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	return "/static/uploads/" + themeUploadsSubdirectory + "/" + url.PathEscape(base)
}

func ThemeOptions() []ThemeOption {
	out := make([]ThemeOption, 0, len(builtInThemes))
	for i, opt := range builtInThemes {
		opt.Page = (i / 5) + 1
		out = append(out, opt)
	}
	return out
}

func (s Service) DefaultThemeSettings(ctx context.Context) ThemeSettings {
	defaultTheme, defaultSet, _ := s.ServerDefaultTheme(ctx)
	defaultCustomCSSPath, hasDefaultCustomCSS := s.ServerDefaultCustomCSS(ctx)
	if !defaultSet {
		defaultTheme = defaultThemeName
	}
	defaultTheme = NormalizeThemeChoice(defaultTheme)
	useDefaultCustomTheme := defaultTheme == defaultCustomThemeName && hasDefaultCustomCSS
	if defaultTheme == defaultCustomThemeName && !hasDefaultCustomCSS {
		defaultTheme = defaultThemeName
		useDefaultCustomTheme = false
	}
	settings := ThemeSettings{
		ProfileTheme: defaultTheme,
		AdminTheme:   defaultTheme,
	}
	if useDefaultCustomTheme {
		settings.CustomCSSPath = defaultCustomCSSPath
	}
	settings.ProfileTheme = NormalizeThemeChoice(settings.ProfileTheme)
	settings.AdminTheme = settings.ProfileTheme
	settings.CustomCSSURL = ThemeCustomCSSURL(settings.CustomCSSPath)
	settings.InlineCSSTemplate = template.CSS(settings.InlineCSS)
	return settings
}

func (s Service) ThemeSettings(ctx context.Context, user *domain.User) ThemeSettings {
	defaultTheme, defaultSet, _ := s.ServerDefaultTheme(ctx)
	defaultCustomCSSPath, hasDefaultCustomCSS := s.ServerDefaultCustomCSS(ctx)
	policy := s.ServerThemePolicy(ctx)
	if !defaultSet {
		defaultTheme = defaultThemeName
	}
	defaultTheme = NormalizeThemeChoice(defaultTheme)
	useDefaultCustomTheme := defaultTheme == defaultCustomThemeName && hasDefaultCustomCSS
	if defaultTheme == defaultCustomThemeName && !hasDefaultCustomCSS {
		defaultTheme = defaultThemeName
		useDefaultCustomTheme = false
	}

	settings := ThemeSettings{
		ProfileTheme: defaultTheme,
		AdminTheme:   defaultTheme,
	}
	if useDefaultCustomTheme {
		settings.CustomCSSPath = defaultCustomCSSPath
	}

	if user != nil {
		userIsAdmin := isAdmin(*user)
		rawTheme := strings.TrimSpace(user.ThemeProfile)
		normalizedUserTheme := NormalizeThemeChoice(rawTheme)
		userWantsDefaultCustom := normalizedUserTheme == defaultCustomThemeName && hasDefaultCustomCSS
		allowUserCustomCSS := userIsAdmin || policy.AllowUserCustomCSS
		useDefaultTheme := (!userIsAdmin && !policy.AllowUserTheme) || rawTheme == ""

		if useDefaultTheme {
			settings.ProfileTheme = defaultTheme
			if useDefaultCustomTheme {
				settings.CustomCSSPath = defaultCustomCSSPath
				settings.InlineCSS = ""
			} else if allowUserCustomCSS {
				settings.CustomCSSPath = strings.TrimSpace(user.ThemeCustomCSSPath)
				settings.InlineCSS = user.ThemeCustomCSSInline
			}
		} else if userWantsDefaultCustom {
			settings.ProfileTheme = defaultCustomThemeName
			settings.CustomCSSPath = defaultCustomCSSPath
			settings.InlineCSS = ""
		} else {
			settings.ProfileTheme = normalizedUserTheme
			if allowUserCustomCSS {
				settings.CustomCSSPath = strings.TrimSpace(user.ThemeCustomCSSPath)
				settings.InlineCSS = user.ThemeCustomCSSInline
			}
		}

		if settings.ProfileTheme == "" {
			settings.ProfileTheme = defaultTheme
		}

		if !userIsAdmin && !policy.AllowUserCustomCSS && settings.ProfileTheme != defaultCustomThemeName {
			settings.CustomCSSPath = ""
			settings.InlineCSS = ""
		}

	} else {
		rawTheme := ""
		if owner, err := s.store.GetOwnerUser(ctx); err == nil {
			rawTheme = strings.TrimSpace(owner.ThemeProfile)
			if rawTheme == "" {
				settings.ProfileTheme = defaultTheme
			} else {
				settings.ProfileTheme = NormalizeThemeChoice(rawTheme)
			}
			settings.CustomCSSPath = strings.TrimSpace(owner.ThemeCustomCSSPath)
			settings.InlineCSS = owner.ThemeCustomCSSInline
			if settings.ProfileTheme == "" {
				settings.ProfileTheme = defaultTheme
			}
		}
	}

	settings.ProfileTheme = NormalizeThemeChoice(settings.ProfileTheme)
	settings.AdminTheme = settings.ProfileTheme
	settings.CustomCSSURL = ThemeCustomCSSURL(settings.CustomCSSPath)
	settings.InlineCSSTemplate = template.CSS(settings.InlineCSS)
	return settings
}

func (s Service) ServerDefaultTheme(ctx context.Context) (string, bool, bool) {
	values, err := s.store.GetSettings(ctx, themeDefaultKey, themeDefaultForceKey)
	if err != nil {
		return "", false, false
	}
	value := NormalizeThemeChoice(values[themeDefaultKey])
	set := strings.TrimSpace(values[themeDefaultKey]) != ""
	force := parseSettingBool(values[themeDefaultForceKey])
	return value, set, force
}

func (s Service) SaveServerDefaultTheme(ctx context.Context, theme string, force bool) error {
	theme = NormalizeThemeChoice(theme)
	forceValue := "0"
	if force {
		forceValue = "1"
	}
	return s.store.SetSettings(ctx, map[string]string{
		themeDefaultKey:      theme,
		themeDefaultForceKey: forceValue,
	})
}

func (s Service) ServerDefaultCustomCSS(ctx context.Context) (string, bool) {
	value, ok, err := s.store.GetSetting(ctx, themeDefaultCustomCSSPathKey)
	if err != nil || !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func (s Service) SaveServerDefaultCustomCSS(ctx context.Context, path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return s.store.DeleteSetting(ctx, themeDefaultCustomCSSPathKey)
	}
	return s.store.SetSetting(ctx, themeDefaultCustomCSSPathKey, trimmed)
}

func (s Service) ServerThemePolicy(ctx context.Context) ThemePolicy {
	policy := ThemePolicy{AllowUserTheme: true, AllowUserCustomCSS: true}
	values, err := s.store.GetSettings(ctx, themeUserSelectKey, themeUserCustomCSSKey)
	if err == nil {
		if val, ok := values[themeUserSelectKey]; ok {
			policy.AllowUserTheme = parseSettingBool(val)
		}
		if val, ok := values[themeUserCustomCSSKey]; ok {
			policy.AllowUserCustomCSS = parseSettingBool(val)
		}
	}
	if !policy.AllowUserTheme {
		policy.AllowUserCustomCSS = false
	}
	return policy
}

func (s Service) SaveServerThemePolicy(ctx context.Context, policy ThemePolicy) error {
	if !policy.AllowUserTheme {
		policy.AllowUserCustomCSS = false
	}
	themeValue := "0"
	if policy.AllowUserTheme {
		themeValue = "1"
	}
	customValue := "0"
	if policy.AllowUserCustomCSS {
		customValue = "1"
	}
	return s.store.SetSettings(ctx, map[string]string{
		themeUserSelectKey:    themeValue,
		themeUserCustomCSSKey: customValue,
	})
}

func (s Service) SaveThemeSettings(ctx context.Context, userID int, settings ThemeSettings) error {
	settings.ProfileTheme = NormalizeThemeChoice(settings.ProfileTheme)
	settings.AdminTheme = settings.ProfileTheme
	settings.CustomCSSPath = strings.TrimSpace(settings.CustomCSSPath)
	settings.CustomCSSURL = ThemeCustomCSSURL(settings.CustomCSSPath)
	settings.InlineCSSTemplate = template.CSS(settings.InlineCSS)
	return s.store.UpdateUserTheme(ctx, userID, settings.ProfileTheme, settings.CustomCSSPath, settings.InlineCSS)
}

func parseSettingBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func isAdmin(user domain.User) bool {
	return strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
}
