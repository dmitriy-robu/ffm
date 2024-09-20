package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
	"go-fitness/external/config"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"go-fitness/internal/api/service"
	"go-fitness/internal/api/types"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type AdminAuthMiddleware struct {
	log        *slog.Logger
	ch         *cache.Cache
	validation *validator.Validate

	cfg *config.Config

	userService *service.UserService
}

func NewAdminAuthMiddleware(
	log *slog.Logger,
	cache *cache.Cache,
	userService *service.UserService,
	cfg *config.Config,
) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		log:         log,
		ch:          cache,
		userService: userService,
		validation:  validator.New(),
		cfg:         cfg,
	}
}

// New TODO: check if payed
func (m *AdminAuthMiddleware) New() func(next http.Handler) http.Handler {
	SecretKey := []byte(m.cfg.JWT)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "http.middleware.AuthMiddleware.New"

			log := m.log.With(
				sl.String("op", op),
				sl.String("request_id", middleware.GetReqID(r.Context())),
			)

			unauthorizedResponse := func(logMessage string, err error) {
				if err != nil {
					log.Warn(logMessage, sl.Err(err))
				} else {
					log.Warn(logMessage)
				}
				response.Respond(w, response.Response{
					Status:  http.StatusUnauthorized,
					Message: "Unauthorized",
				})
			}

			tokenHeader := r.Header.Get("Authorization")
			if tokenHeader == "" {
				unauthorizedResponse("token wasn't provided", nil)
				return
			}

			tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")

			user, err := m.validateTokenAndGetUser(tokenString, SecretKey, log)
			if err != nil {
				unauthorizedResponse("failed to validate token or get user", err)
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *AdminAuthMiddleware) validateTokenAndGetUser(
	tokenString string,
	secretKey []byte,
	log *slog.Logger,
) (types.User, error) {
	log = log.With(
		sl.String("token", tokenString),
		sl.String("secret_key", string(secretKey)),
	)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("unexpected signing method", sl.String("alg", fmt.Sprintf("%v", token.Header["alg"])))
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})
	if err != nil {
		log.Error("failed to parse token", sl.Err(err))
		return types.User{}, errors.New("failed to parse token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Error("invalid token", sl.String("token", tokenString))
		return types.User{}, errors.New("invalid token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		log.Error("token does not contain exp", sl.String("token", tokenString))
		return types.User{}, errors.New("token does not contain exp")
	}

	if time.Now().Unix() > int64(exp) {
		log.Error("token has expired", sl.String("token", tokenString))
		return types.User{}, errors.New("token has expired")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userUUID, ok := claims["user_uuid"].(string)
	if !ok {
		log.Error("token does not contain user UUID", sl.String("token", tokenString))
		return types.User{}, errors.New("token does not contain user UUID")
	}

	user, err := m.userService.GetUserByUUID(ctx, userUUID)
	if err != nil {
		log.Error("failed to get user by UUID", sl.Err(err))
		return types.User{}, errors.New("failed to get user by UUID")
	}

	if user.UUID == "" {
		log.Error("user not found", sl.String("user_uuid", userUUID))
		return types.User{}, errors.New("user not found")
	}

	role, err := m.userService.GetRoleByUserID(ctx, user.ID)
	if err != nil {
		log.Error("failed to get role by user ID", sl.Err(err))
		return types.User{}, errors.New("failed to get role by user ID")
	}

	if role.Name != "admin" {
		log.Error("user is not admin", sl.String("user_uuid", userUUID))
		return types.User{}, errors.New("user is not admin")
	}

	return user, nil
}
