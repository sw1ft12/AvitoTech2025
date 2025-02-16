package server

import (
	"AvitoTech/internal/api"
	"AvitoTech/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *gin.Engine
	auth   *api.AuthService
	info   *api.InfoService
}

func (s *Server) Run(address string) error {
	g := s.router.Group("/api")
	{
		g.POST("/auth", s.auth.Login)
		g.GET("/info", api.JwtMiddleware(), s.info.GetUserInfo)
		g.POST("/sendCoin", api.JwtMiddleware(), s.info.SendCoins)
		g.GET("/buy/:item", api.JwtMiddleware(), s.info.BuyItem)
		g.POST("/topUpBalance", api.JwtMiddleware(), s.info.TopUpBalance)
	}
	return s.router.Run(address)
}

func NewServer(pool *pgxpool.Pool) *Server {
	repo := repository.NewRepo(pool)
	return &Server{
		router: gin.Default(),
		auth:   api.NewAuthService(repo),
		info:   api.NewInfoService(repo),
	}
}
