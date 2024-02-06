package handlers

import (
	"net/http"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gin-gonic/gin"
	"github.com/jbrit/gopaypeer/core"
	"github.com/jbrit/gopaypeer/models"
	"gorm.io/gorm"
)

type TransferInput struct {
	Currency      string `json:"currency" form:"currency" binding:"required,oneof=ngn_kobo usd_cent"`
	Amount        uint64 `json:"amount" form:"amount" binding:"required"`
	EmailOrPubkey string `json:"email_or_pubkey" form:"email_or_pubkey" binding:"required"`
}

func MakeTransfer(c *gin.Context, db *gorm.DB) {
	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}

	var toUser models.User
	if err := db.Where("email = ?", input.EmailOrPubkey).First(&toUser).Error; err != nil {
		if err := db.Where("public_key = ?", input.EmailOrPubkey).First(&toUser).Error; err != nil {
			// TODO: possibly throw another error is user does not exist?
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Could not find user"})
			return
		}
	}

	mint := core.NgnMint
	if input.Currency == "usd_cent" {
		mint = core.UsdMint
	}

	owner := solana.MustPublicKeyFromBase58(user.PublicKey)
	ownerAta, _, err := solana.FindAssociatedTokenAddress(
		owner,
		mint,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	destinationAta, _, err := solana.FindAssociatedTokenAddress(
		solana.MustPublicKeyFromBase58(toUser.PublicKey),
		mint,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: improve gas supply flow
	if _, err = user.MakeSolTransfer(50000000, solana.MustPublicKeyFromBase58(toUser.PublicKey)); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, _, err = toUser.GetAssociatedTokenAccountBalance(mint); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err = user.MakeTransaction([]solana.Instruction{
		token.NewTransferCheckedInstruction(
			input.Amount*1e7, // * 10e7 for remaining decimals
			9,                // no of decimals
			ownerAta,
			mint,
			destinationAta,
			owner,
			[]solana.PublicKey{owner},
		).Build(),
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer Successful"})
}
