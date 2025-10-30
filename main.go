package main

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	ConcurrencyLimit = 2
	CacheMaxSize     = 50
	CacheTTL         = 10 * time.Minute
	RedisTTL         = 30 * time.Minute
)

// TTLCache implements a simple LRU cache with TTL
type TTLCache struct {
	mu         sync.Mutex
	maxSize    int
	ttl        time.Duration
	cache      map[string]*list.Element
	lruList    *list.List
	timestamps map[string]time.Time
}

type cacheEntry struct {
	key   string
	value interface{}
}

// NewTTLCache creates a new TTL cache
func NewTTLCache(maxSize int, ttl time.Duration) *TTLCache {
	return &TTLCache{
		maxSize:    maxSize,
		ttl:        ttl,
		cache:      make(map[string]*list.Element),
		lruList:    list.New(),
		timestamps: make(map[string]time.Time),
	}
}

// Get retrieves a value from the cache
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	timestamp, ok := c.timestamps[key]
	if !ok || time.Since(timestamp) > c.ttl {
		// Remove expired item
		c.lruList.Remove(element)
		delete(c.cache, key)
		delete(c.timestamps, key)
		return nil, false
	}

	// Move to end (most recently used)
	c.lruList.MoveToBack(element)
	return element.Value.(*cacheEntry).value, true
}

// Put adds a value to the cache
func (c *TTLCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.cache[key]; exists {
		// Update existing item
		c.lruList.MoveToBack(element)
		element.Value.(*cacheEntry).value = value
		c.timestamps[key] = time.Now()
		return
	}

	// Add new item
	if c.lruList.Len() >= c.maxSize {
		// Remove least recently used item
		oldest := c.lruList.Front()
		if oldest != nil {
			entry := oldest.Value.(*cacheEntry)
			c.lruList.Remove(oldest)
			delete(c.cache, entry.key)
			delete(c.timestamps, entry.key)
		}
	}

	entry := &cacheEntry{key: key, value: value}
	element := c.lruList.PushBack(entry)
	c.cache[key] = element
	c.timestamps[key] = time.Now()
}

// Size returns the current size of the cache
func (c *TTLCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lruList.Len()
}

// Clear removes all items from the cache
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*list.Element)
	c.lruList = list.New()
	c.timestamps = make(map[string]time.Time)
}

// ConcurrencyLimiter middleware to limit concurrent requests
func ConcurrencyLimiter(limit int) gin.HandlerFunc {
	semaphore := make(chan struct{}, limit)

	return func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
			c.Next()
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "server is at capacity, please try again later",
			})
			c.Abort()
		}
	}
}

// ComponentTemplate represents a React component template
type ComponentTemplate struct {
	ComponentName string   `json:"component_name"`
	ComponentType string   `json:"component_type"`
	Template      string   `json:"template"`
	RequiredKeys  []string `json:"required_keys"`
}

// ComponentMetadata represents metadata for a component
type ComponentMetadata struct {
	ComponentID  string   `json:"component_id"`
	LastUpdated  string   `json:"last_updated"`
	RequiredKeys []string `json:"required_keys"`
}

// LocalizedComponent represents the response structure
type LocalizedComponent struct {
	ComponentName string            `json:"component_name"`
	ComponentType string            `json:"component_type"`
	Language      string            `json:"language"`
	Template      string            `json:"template"`
	LocalizedData map[string]string `json:"localized_data"`
	Metadata      ComponentMetadata `json:"metadata"`
	Cached        bool              `json:"cached,omitempty"`
}

// Localization database
var localizationDB = map[string]map[string]string{
	"en": {
		"welcome_title":      "Welcome to Our App",
		"welcome_subtitle":   "Your journey starts here",
		"login_button":       "Log In",
		"signup_button":      "Sign Up",
		"navigation_home":    "Home",
		"navigation_about":   "About",
		"navigation_contact": "Contact",
		"footer_copyright":   "© 2024 Our Company. All rights reserved.",
		"user_profile_title": "User Profile",
		"user_profile_edit":  "Edit Profile",
		"settings_title":     "Settings",
		"settings_language":  "Language",
		"settings_theme":     "Theme",
		"error_404":          "Page not found",
		"error_500":          "Internal server error",
	},
	"es": {
		"welcome_title":      "Bienvenido a Nuestra App",
		"welcome_subtitle":   "Tu viaje comienza aquí",
		"login_button":       "Iniciar Sesión",
		"signup_button":      "Registrarse",
		"navigation_home":    "Inicio",
		"navigation_about":   "Acerca de",
		"navigation_contact": "Contacto",
		"footer_copyright":   "© 2024 Nuestra Empresa. Todos los derechos reservados.",
		"user_profile_title": "Perfil de Usuario",
		"user_profile_edit":  "Editar Perfil",
		"settings_title":     "Configuración",
		"settings_language":  "Idioma",
		"settings_theme":     "Tema",
		"error_404":          "Página no encontrada",
		"error_500":          "Error interno del servidor",
	},
	"fr": {
		"welcome_title":      "Bienvenue dans Notre App",
		"welcome_subtitle":   "Votre voyage commence ici",
		"login_button":       "Se Connecter",
		"signup_button":      "S'inscrire",
		"navigation_home":    "Accueil",
		"navigation_about":   "À Propos",
		"navigation_contact": "Contact",
		"footer_copyright":   "© 2024 Notre Entreprise. Tous droits réservés.",
		"user_profile_title": "Profil Utilisateur",
		"user_profile_edit":  "Modifier le Profil",
		"settings_title":     "Paramètres",
		"settings_language":  "Langue",
		"settings_theme":     "Thème",
		"error_404":          "Page non trouvée",
		"error_500":          "Erreur interne du serveur",
	},
	"de": {
		"welcome_title":      "Willkommen in Unserer App",
		"welcome_subtitle":   "Ihre Reise beginnt hier",
		"login_button":       "Anmelden",
		"signup_button":      "Registrieren",
		"navigation_home":    "Startseite",
		"navigation_about":   "Über Uns",
		"navigation_contact": "Kontakt",
		"footer_copyright":   "© 2024 Unser Unternehmen. Alle Rechte vorbehalten.",
		"user_profile_title": "Benutzerprofil",
		"user_profile_edit":  "Profil Bearbeiten",
		"settings_title":     "Einstellungen",
		"settings_language":  "Sprache",
		"settings_theme":     "Design",
		"error_404":          "Seite nicht gefunden",
		"error_500":          "Interner Serverfehler",
	},
}

// Component templates
var componentTemplates = map[string]ComponentTemplate{
	"welcome": {
		ComponentName: "WelcomeComponent",
		ComponentType: "functional",
		Template: `
import React from 'react';

const WelcomeComponent = ({ className = "welcome-container" }) => {
  return (
    <div className={className}>
      <div className="welcome-wrapper">
        <header className="welcome-header">
          <h1 className="welcome-title" data-l10n="welcome_title">
            {l10n.welcome_title}
          </h1>
          <p className="welcome-subtitle" data-l10n="welcome_subtitle">
            {l10n.welcome_subtitle}
          </p>
        </header>
        <div className="welcome-actions">
          <button 
            className="btn btn-primary" 
            onClick={handleLogin}
            data-l10n="login_button"
          >
            {l10n.login_button}
          </button>
          <button 
            className="btn btn-secondary" 
            onClick={handleSignup}
            data-l10n="signup_button"
          >
            {l10n.signup_button}
          </button>
        </div>
      </div>
    </div>
  );
};

export default WelcomeComponent;
`,
		RequiredKeys: []string{"welcome_title", "welcome_subtitle", "login_button", "signup_button"},
	},
	"navigation": {
		ComponentName: "NavigationComponent",
		ComponentType: "functional",
		Template: `
import React from 'react';

const NavigationComponent = ({ className = "navigation-container" }) => {
  return (
    <nav className={className}>
      <ul className="nav-list">
        <li className="nav-item">
          <a href="/" className="nav-link" data-l10n="navigation_home">
            {l10n.navigation_home}
          </a>
        </li>
        <li className="nav-item">
          <a href="/about" className="nav-link" data-l10n="navigation_about">
            {l10n.navigation_about}
          </a>
        </li>
        <li className="nav-item">
          <a href="/contact" className="nav-link" data-l10n="navigation_contact">
            {l10n.navigation_contact}
          </a>
        </li>
      </ul>
    </nav>
  );
};

export default NavigationComponent;
`,
		RequiredKeys: []string{"navigation_home", "navigation_about", "navigation_contact"},
	},
	"user_profile": {
		ComponentName: "UserProfileComponent",
		ComponentType: "functional",
		Template: `
import React from 'react';

const UserProfileComponent = ({ className = "user-profile-container" }) => {
  return (
    <div className={className}>
      <div className="profile-wrapper">
        <h2 className="profile-title" data-l10n="user_profile_title">
          {l10n.user_profile_title}
        </h2>
        <div className="profile-actions">
          <button 
            className="btn btn-outline" 
            onClick={handleEditProfile}
            data-l10n="user_profile_edit"
          >
            {l10n.user_profile_edit}
          </button>
        </div>
      </div>
    </div>
  );
};

export default UserProfileComponent;
`,
		RequiredKeys: []string{"user_profile_title", "user_profile_edit"},
	},
	"footer": {
		ComponentName: "FooterComponent",
		ComponentType: "functional",
		Template: `
import React from 'react';

const FooterComponent = ({ className = "footer-container" }) => {
  return (
    <footer className={className}>
      <div className="footer-content">
        <p className="footer-copyright" data-l10n="footer_copyright">
          {l10n.footer_copyright}
        </p>
      </div>
    </footer>
  );
};

export default FooterComponent;
`,
		RequiredKeys: []string{"footer_copyright"},
	},
}

// Global cache instances
var componentCache = NewTTLCache(CacheMaxSize, CacheTTL)
var redisClient *redis.Client
var ctx = context.Background()

// initRedis initializes the Redis client
func initRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return client
}

// getFromRedis retrieves a component from Redis
func getFromRedis(key string) (*LocalizedComponent, error) {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var component LocalizedComponent
	if err := json.Unmarshal([]byte(val), &component); err != nil {
		return nil, err
	}

	return &component, nil
}

// setInRedis stores a component in Redis with TTL
func setInRedis(key string, component *LocalizedComponent) error {
	data, err := json.Marshal(component)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, key, data, RedisTTL).Err()
}

// interpolateTemplate replaces {l10n.key} patterns with actual localized values
func interpolateTemplate(template string, localizedData map[string]string) string {
	result := template
	for key, value := range localizedData {
		pattern := regexp.MustCompile(`\{l10n\.` + regexp.QuoteMeta(key) + `\}`)
		result = pattern.ReplaceAllString(result, fmt.Sprintf(`"%s"`, value))
	}
	return result
}

// getLocalizedComponent generates a localized React component
func getLocalizedComponent(componentType, lang string) (*LocalizedComponent, error) {
	template, exists := componentTemplates[componentType]
	if !exists {
		return nil, fmt.Errorf("component type '%s' not found", componentType)
	}

	// Get localized strings, fallback to English
	strings, exists := localizationDB[lang]
	if !exists {
		strings = localizationDB["en"]
	}

	// Get only the required keys for this component
	componentStrings := make(map[string]string)
	for _, key := range template.RequiredKeys {
		if value, ok := strings[key]; ok {
			componentStrings[key] = value
		} else {
			componentStrings[key] = fmt.Sprintf("[%s]", key)
		}
	}

	// Interpolate template
	localizedTemplate := interpolateTemplate(template.Template, componentStrings)

	// Generate component ID with timestamp
	componentID := fmt.Sprintf("%s_%s_%d", componentType, lang, time.Now().UnixMilli()%10000)

	return &LocalizedComponent{
		ComponentName: template.ComponentName,
		ComponentType: template.ComponentType,
		Language:      lang,
		Template:      localizedTemplate,
		LocalizedData: componentStrings,
		Metadata: ComponentMetadata{
			ComponentID:  componentID,
			LastUpdated:  "2024-01-15T10:30:00Z",
			RequiredKeys: template.RequiredKeys,
		},
	}, nil
}

// Health check handler
func healthCheck(c *gin.Context) {
	redisStatus := "disconnected"
	if redisClient != nil {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			redisStatus = "connected"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "healthy",
		"service":           "localization-manager-backend",
		"version":           "0.1.0",
		"cache_size":        componentCache.Size(),
		"concurrency_limit": ConcurrencyLimit,
		"redis_status":      redisStatus,
	})
}

// Get localized component handler
func getLocalizedComponentEndpoint(c *gin.Context) {
	componentType := c.Param("component_type")
	lang := c.DefaultQuery("lang", "en")

	cacheKey := fmt.Sprintf("component:%s:%s", componentType, lang)

	// Check TTL cache first
	if cached, found := componentCache.Get(cacheKey); found {
		component := cached.(*LocalizedComponent)
		// refresh the cache
		componentCache.Put(cacheKey, component)
		// refresh the redis cache
		setInRedis(cacheKey, component)
		response := *component
		response.Cached = true
		c.JSON(http.StatusOK, response)
		return
	}

	// TTL cache miss, check Redis
	if redisClient != nil {
		component, err := getFromRedis(cacheKey)
		if err == nil && component != nil {
			// Found in Redis, store in TTL cache
			componentCache.Put(cacheKey, component)

			// Refresh Redis TTL
			setInRedis(cacheKey, component)

			response := *component
			response.Cached = true
			c.JSON(http.StatusOK, response)
			return
		}
	}

	// Both caches missed, generate component
	component, err := getLocalizedComponent(componentType, lang)
	if err != nil {
		availableComponents := make([]string, 0, len(componentTemplates))
		for key := range componentTemplates {
			availableComponents = append(availableComponents, key)
		}
		c.JSON(http.StatusNotFound, gin.H{
			"error":                err.Error(),
			"available_components": availableComponents,
		})
		return
	}

	// Store in both caches
	componentCache.Put(cacheKey, component)
	if redisClient != nil {
		setInRedis(cacheKey, component)
	}

	component.Cached = false
	c.JSON(http.StatusOK, component)
}

func main() {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	// Initialize Redis
	redisClient = initRedis()
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		fmt.Printf("⚠️  Redis connection failed: %v (continuing without Redis)\n", err)
		redisClient = nil
	} else {
		fmt.Println("✅ Redis connected successfully")
	}

	router := gin.Default()

	// Apply concurrency limiter middleware
	router.Use(ConcurrencyLimiter(ConcurrencyLimit))

	// Routes
	router.GET("/health", healthCheck)
	router.GET("/api/component/:component_type", getLocalizedComponentEndpoint)

	// Start server
	fmt.Println("🚀 Localization Manager Backend starting on :8000")
	fmt.Println("📚 Available components:", strings.Join(getComponentKeys(), ", "))
	fmt.Println("🌍 Supported languages:", strings.Join(getLanguageKeys(), ", "))

	if err := router.Run(":8000"); err != nil {
		panic(err)
	}
}

// Helper function to get component keys
func getComponentKeys() []string {
	keys := make([]string, 0, len(componentTemplates))
	for key := range componentTemplates {
		keys = append(keys, key)
	}
	return keys
}

// Helper function to get language keys
func getLanguageKeys() []string {
	keys := make([]string, 0, len(localizationDB))
	for key := range localizationDB {
		keys = append(keys, key)
	}
	return keys
}
