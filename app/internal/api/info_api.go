package api

import (
	"AvitoTech/internal/models"
	"AvitoTech/internal/repository"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Merch struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
}

type InfoService struct {
	products    map[string]int
	infoStorage InfoStorage
}

type InfoStorage interface {
	GetUserInfo(ctx context.Context, username string) (models.Info, error)
	SendCoins(ctx context.Context, sender, receiver string, amount int) error
	BuyItem(ctx context.Context, username string, itemName string, itemCost int) error
	TopUpBalance(ctx context.Context, username string, amount int) error
}

func NewInfoService(storage *repository.StoragePostgres) *InfoService {
	return &InfoService{
		products: map[string]int{
			"t-shirt":    80,
			"cup":        20,
			"book":       50,
			"pen":        10,
			"powerbank":  200,
			"hoody":      300,
			"umbrella":   200,
			"socks":      10,
			"wallet":     50,
			"pink-hoody": 500,
		},
		infoStorage: storage,
	}
}

func (i *InfoService) GetUserInfo(ctx *gin.Context) {
	username := ctx.Value("username").(string)
	info, err := i.infoStorage.GetUserInfo(ctx.Request.Context(), username)
	if err != nil {
		log.Println("api/GetUserInfo " + err.Error())
		ctx.JSON(http.StatusInternalServerError, "Failed to get info")
		return
	}
	ctx.JSON(http.StatusOK, info)
}

func (i *InfoService) BuyItem(ctx *gin.Context) {
	item := ctx.Param("item")
	if _, ok := i.products[item]; !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"errors": "Invalid item"})
		return
	}
	username := ctx.Value("username").(string)
	err := i.infoStorage.BuyItem(ctx.Request.Context(), username, item, i.products[item])
	if err != nil {
		var internalErr *repository.InternalError
		if !errors.As(err, &internalErr) {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}
		log.Println("api/BuyItem: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to buy item"})
		return
	}
	ctx.Status(http.StatusOK)
}

func (i *InfoService) SendCoins(ctx *gin.Context) {
	type Send struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}
	var s Send
	err := ctx.BindJSON(&s)
	if err != nil || s.Amount <= 0 {
		if err == nil {
			err = errors.New("amount must be positive")
		}
		log.Println("api/SendCoins: " + err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	username := ctx.Value("username").(string)
	err = i.infoStorage.SendCoins(ctx.Request.Context(), username, s.ToUser, s.Amount)
	if err != nil {
		var internalErr *repository.InternalError
		if !errors.As(err, &internalErr) {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}
		log.Println("api/SendCoins: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send coins"})
		return
	}
	ctx.Status(http.StatusOK)
}

func (i *InfoService) TopUpBalance(ctx *gin.Context) {
	type Amount struct {
		Amount int `json:"amount"`
	}
	var a Amount
	err := ctx.BindJSON(&a)
	if err != nil || a.Amount <= 0 {
		if err == nil {
			err = errors.New("amount must be positive")
		}
		log.Println("api/TopUpBalance: " + err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	username := ctx.Value("username").(string)
	err = i.infoStorage.TopUpBalance(ctx.Request.Context(), username, a.Amount)
	if err != nil {
		log.Println("api/SendCoins: " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to up balance"})
		return
	}
	ctx.Status(http.StatusOK)
}
