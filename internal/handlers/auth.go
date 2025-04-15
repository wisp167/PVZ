package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/wisp167/pvz/api"
)

const (
	RoleKey = "role"
)

type JWTConfig struct {
	Skipper    func(c echo.Context) bool
	SigningKey []byte
}

func AuthWithConfig(config JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return config.SigningKey, nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if customClaims, ok := token.Claims.(*Claims); ok {
				if customClaims.Role != "employee" && customClaims.Role != "moderator" {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid role")
				}
				c.Set(RoleKey, customClaims.Role) // Now you can access the Role field
			} else {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			return next(c)
		}
	}
}

func RoleRequired(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			role, ok := c.Get(RoleKey).(string)

			if !ok || role != requiredRole {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}
			c.Set(RoleKey, nil)
			return next(c)
		}
	}
}

type Claims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

type AuthResponse struct {
	Token string `json:"token"`
}

func DummyLogin(role string, jwtkey []byte) (api.Token, error) {
	expirationTime := time.Now().Add(time.Hour)
	claims := &Claims{
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtkey)
	if err != nil {
		return api.Token(""), err
	}

	return api.Token(tokenString), nil
}
