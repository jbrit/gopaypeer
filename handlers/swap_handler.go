package handlers

import (
	"encoding/binary"
	"net/http"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/jbrit/gopaypeer/core"
	"gorm.io/gorm"
)

type SwapInput struct {
	Currency     string `json:"currency" form:"currency" binding:"required,oneof=ngn_kobo usd_cent"`
	Amount       uint64 `json:"amount" form:"amount"  binding:"required"`
	MinAmountOut uint64 `json:"min_amount" form:"min_amount" `
}

func MakeSwap(c *gin.Context, db *gorm.DB) {
	var input SwapInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := requireAuth(c, db)
	if err != nil {
		return
	}

	fromMint, toMint, swapFromAta, swaptoAta := core.NgnMint, core.UsdMint, core.SwapUsdAta, core.SwapNgnAta
	if input.Currency == "usd_cent" {
		fromMint, toMint, swapFromAta, swaptoAta = core.UsdMint, core.NgnMint, core.SwapNgnAta, core.SwapUsdAta
	}

	owner := solana.MustPublicKeyFromBase58(user.PublicKey)
	ownerFromAta, _, err := solana.FindAssociatedTokenAddress(
		owner,
		fromMint,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ownerToAta, _, err := solana.FindAssociatedTokenAddress(
		owner,
		toMint,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	byteData := []byte{1}
	byteData = binary.LittleEndian.AppendUint64(byteData, input.Amount*1e7)
	byteData = binary.LittleEndian.AppendUint64(byteData, input.MinAmountOut*1e7)

	if _, err = user.MakeTransaction([]solana.Instruction{
		solana.NewInstruction(
			core.ProgramId,
			[]*solana.AccountMeta{
				solana.Meta(owner).SIGNER().WRITE(),
				solana.Meta(core.SwapPool),
				solana.Meta(solana.TokenProgramID),
				solana.Meta(owner),
				solana.Meta(ownerFromAta).WRITE(),
				solana.Meta(ownerToAta).WRITE(),
				solana.Meta(core.SwapAuthority).WRITE(),
				solana.Meta(swapFromAta).WRITE(),
				solana.Meta(swaptoAta).WRITE(),
				solana.Meta(solana.SystemProgramID),
			},
			byteData,
		),
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Swap Successful"})
}
