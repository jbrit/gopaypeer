package handlers

import (
	"fmt"
	"net/http"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/jbrit/gopaypeer/core"
	"github.com/jbrit/gopaypeer/models"
	"gorm.io/gorm"
)

func requireAuth(c *gin.Context, db *gorm.DB) (*models.User, error) {
	claims := &Claims{}
	if err := VerifyJwt(c.GetHeader("Authorization"), jwtKey, claims); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return nil, err
	}

	var user models.User
	if err := db.Preload("DebitCard").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return nil, fmt.Errorf("")
	}
	return &user, nil
}

func CurrentUser(c *gin.Context, db *gorm.DB) {

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type Balance struct {
	amount *uint64
	err    error
}

func getBalance(user *models.User, mint solana.PublicKey) chan Balance {
	r := make(chan Balance)
	amount, err := user.GetAssociatedTokenAccountBalance(mint)

	go func() {
		r <- Balance{
			amount: amount,
			err:    err,
		}
	}()

	return r
}

func GetBalances(c *gin.Context, db *gorm.DB) {

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}

	chNgn, chnUsd := getBalance(user, core.NgnMint), getBalance(user, core.UsdMint)
	ngnBalance, usdBalance := <-chNgn, <-chnUsd

	if ngnBalance.err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ngnBalance.err.Error()})
		return
	}
	if usdBalance.err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": usdBalance.err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ngn_kobo": ngnBalance.amount, "usd_cent": usdBalance.amount})
}

func PubkeyToUser(c *gin.Context, db *gorm.DB) {
	var users []models.User
	// Get all records
	result := db.Find(&users)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch addresses"})
	}

	pubkeyToUser := map[string]string{}
	for _, user := range users {
		pubkeyToUser[user.PublicKey] = user.Email

	}

	c.JSON(http.StatusOK, pubkeyToUser)
}
