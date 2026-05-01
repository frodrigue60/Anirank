package middleware

import (
	"log"
	"strings"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/auth"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT context over routes and ensures user still exists in DB
func AuthMiddleware(jwtService *auth.JWTService, userRepo domain.UserRepository, cache domain.Cache) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return domain.NewAppError(401, "Missing authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return domain.NewAppError(401, "Invalid authorization header format", nil)
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Auth Error: %v", err)
			return domain.NewAppError(401, "Invalid or expired token", nil)
		}

		if claims.UserUUID == "" {
			return domain.NewAppError(401, "Invalid token payload: missing user identifier. Please re-login.", nil)
		}

		// 1. Try Cache First
		cacheKey := "session:" + claims.UserUUID
		var cachedUser struct {
			ID   uint64
			UUID string
		}
		
		if cache != nil {
			if err := cache.Get(c.Context(), cacheKey, &cachedUser); err == nil {
				c.Locals("user_id", cachedUser.ID)
				c.Locals("user_uuid", cachedUser.UUID)
				c.Locals("user_roles", claims.Roles)
				return c.Next()
			}
		}

		// 2. Fallback to DB
		user, err := userRepo.GetByUUID(c.Context(), claims.UserUUID)
		if err != nil {
			if err == domain.ErrNotFound {
				return domain.NewAppError(401, "Session belongs to a non-existent user. Please re-login.", nil)
			}
			return domain.NewAppError(500, "Database validation error", err)
		}

		// 3. Store in Cache (Short TTL)
		if cache != nil {
			cachedUser.ID = user.ID
			cachedUser.UUID = user.UUID
			_ = cache.Set(c.Context(), cacheKey, cachedUser, 2*time.Minute)
		}

		c.Locals("user_id", user.ID)
		c.Locals("user_uuid", user.UUID)
		c.Locals("user_roles", claims.Roles)

		return c.Next()
	}
}

// OptionalAuthMiddleware parses JWT if present, but allows request to continue if not
func OptionalAuthMiddleware(jwtService *auth.JWTService, userRepo domain.UserRepository, cache domain.Cache) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			if claims, err := jwtService.ValidateToken(tokenString); err == nil && claims.UserUUID != "" {
				
				// 1. Try Cache
				cacheKey := "session:" + claims.UserUUID
				var cachedUser struct {
					ID   uint64
					UUID string
				}
				if cache != nil {
					if err := cache.Get(c.Context(), cacheKey, &cachedUser); err == nil {
						c.Locals("user_id", cachedUser.ID)
						c.Locals("user_uuid", cachedUser.UUID)
						c.Locals("user_roles", claims.Roles)
						return c.Next()
					}
				}

				// 2. DB Fallback
				if user, uErr := userRepo.GetByUUID(c.Context(), claims.UserUUID); uErr == nil {
					if cache != nil {
						cachedUser.ID = user.ID
						cachedUser.UUID = user.UUID
						_ = cache.Set(c.Context(), cacheKey, cachedUser, 2*time.Minute)
					}
					c.Locals("user_id", user.ID)
					c.Locals("user_uuid", user.UUID)
					c.Locals("user_roles", claims.Roles)
				}
			}
		}

		return c.Next()
	}
}

// StaffMiddleware only allows users with core administrative type roles to proceed.
func StaffMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles, ok := c.Locals("user_roles").([]string)
		if !ok || len(roles) == 0 {
			return domain.NewAppError(403, "Forbidden. Action requires staff privileges", nil)
		}

		isStaff := false
		for _, r := range roles {
			if r == "owner" || r == "admin" || r == "editor" || r == "creator" {
				isStaff = true
				break
			}
		}

		if !isStaff {
			return domain.NewAppError(403, "Forbidden. Action requires staff privileges", nil)
		}
		return c.Next()
	}
}

// AdminMiddleware strictly only allows top-level admin
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles, ok := c.Locals("user_roles").([]string)
		if !ok || len(roles) == 0 {
			return domain.NewAppError(403, "Forbidden. Feature restricted to Administrators", nil)
		}

		isAdmin := false
		for _, r := range roles {
			if r == "owner" || r == "admin" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			return domain.NewAppError(403, "Forbidden. Feature restricted to Administrators", nil)
		}
		return c.Next()
	}
}
