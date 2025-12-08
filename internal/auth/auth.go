package auth

import (
	"local-monitor/internal/db"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

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
		// Set cookie
		c.Cookie(&fiber.Cookie{
			Name:     "auth_session",
			Value:    "valid", // In real world use a signed JWT or session ID
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
		})
		return c.Redirect("/")
	}

	return c.Render("login", fiber.Map{"error": "Invalid credentials", "username": req.Username})
}

func LogoutHandler(c *fiber.Ctx) error {
	c.ClearCookie("auth_session")
	return c.Redirect("/login")
}

func RequireLogin(c *fiber.Ctx) error {
	if c.Path() == "/login" || c.Path() == "/public/css/style.css" || c.Path() == "/public/js/main.js" {
		return c.Next()
	}

	cookie := c.Cookies("auth_session")
	if cookie != "valid" {
		return c.Redirect("/login")
	}
	return c.Next()
}
