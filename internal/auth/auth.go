package auth

import (
	"local-monitor/internal/db"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getJWTSecret())

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func getJWTSecret() string {
	if secret := os.Getenv("MONITOR_SECURITY_JWT_SECRET"); secret != "" {
		return secret
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return "change-this-secret-in-production-use-env-JWT_SECRET"
}

func Login(username, password string) bool {
	hash, err := db.GetUser(username)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateDefaultUser() {
	if !db.UserExists() {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		db.CreateUser("admin", string(hash))
	}
}

func UpdatePassword(username, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.UpdateUser(username, string(hash))
}

func NewAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// potential for session check here if using cookies
		// for now we stick to simple check or token
		return c.Next()
	}
}

func LoginHandler(c *fiber.Ctx) error {
	type LoginReq struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	var req LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Render("login", fiber.Map{"error": "Invalid request"})
	}

	if Login(req.Username, req.Password) {
		token, err := generateJWT(req.Username)
		if err != nil {
			return c.Render("login", fiber.Map{"error": "Failed to generate token"})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: "Lax",
		})
		return c.Redirect("/")
	}

	return c.Render("login", fiber.Map{"error": "Invalid credentials", "username": req.Username})
}

func generateJWT(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func LogoutHandler(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
	return c.Redirect("/login")
}

func RequireLogin(c *fiber.Ctx) error {
	if c.Path() == "/login" || strings.HasPrefix(c.Path(), "/public/") {
		return c.Next()
	}

	token := c.Cookies("auth_token")
	if token == "" {
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		return c.Redirect("/login")
	}

	claims, err := validateJWT(token)
	if err != nil || claims == nil {
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}
		return c.Redirect("/login")
	}

	c.Locals("username", claims.Username)
	return c.Next()
}
