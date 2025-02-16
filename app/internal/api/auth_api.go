package api

import (
	"AvitoTech/internal/models"
	"AvitoTech/internal/repository"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var jwtSecretKey = []byte("sw1ft")

type AuthService struct {
	userStorage UserStorage
}

type UserStorage interface {
	CreateUser(ctx context.Context, data models.AuthRequest) (models.User, error)
}

func NewAuthService(storage *repository.StoragePostgres) *AuthService {
	return &AuthService{
		userStorage: storage,
	}
}

func (a *AuthService) Login(ctx *gin.Context) {
	var input models.AuthRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	input.Password = string(hashedPassword)

	user, err := a.userStorage.CreateUser(ctx.Request.Context(), input)
	if err != nil {
		log.Println("api/Login:" + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server is not available"})
		return
	}

	payload := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		log.Println("api/Login:" + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server is not available"})
		return
	}
	ctx.JSON(http.StatusOK, models.AuthResponse{Token: t})
}

func JwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing authorization header"})
			return
		}

		tokenString := authHeader[7:]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		username := claims["username"].(string)
		ctx.Set("username", username)
		ctx.Next()
	}
}
