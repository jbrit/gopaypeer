package models

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/google/uuid"
	"github.com/jbrit/gopaypeer/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PaypeerDebitCard struct {
	ID         string    `json:"id"  gorm:"primarykey"`
	CardNumber string    `json:"card_number"`
	ExpiryDate time.Time `json:"expiry_date"`
	Cvv        string    `json:"cvv"`
	CardActive bool      `json:"card_active"`
	Balance    uint      `json:"balance"`
}

type User struct {
	ID            string           `json:"id" gorm:"primarykey"`
	Email         string           `json:"email" gorm:"unique"`
	EmailVerified bool             `json:"email_verified"`
	PasswordHash  string           `json:"-"`
	Otp           string           `json:"-"`
	OtpExpiresAt  time.Time        `json:"-"`
	PrivateKey    string           `json:"-"`
	PublicKey     string           `json:"pubkey"`
	FirstName     string           `json:"first_name"`
	LastName      string           `json:"last_name"`
	DebitCardID   string           `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DebitCard     PaypeerDebitCard `json:"debit_card"`
}

func (u *User) SendMail(subject string, message string) error {
	if os.Getenv("MAIL_MODE") == "PRODUCTION" {
		err := core.SendMail(u.Email, subject, message)
		if err != nil {
			return err
		}
	}
	fmt.Println("=====")
	fmt.Println("email to", u.Email)
	fmt.Println("email subject", subject)
	fmt.Println(message)
	fmt.Println("=====")
	return nil
}

func (user *User) ExpireOTP(OTP string, db *gorm.DB) error {
	// TODO: limit to 3 valid calls
	if !user.OtpExpiresAt.Before(time.Now()) && user.Otp == OTP {
		user.OtpExpiresAt = time.Now()
		return db.Save(user).Error
	}
	return fmt.Errorf("Invalid OTP")
}

func (user *User) GetOrCreateSolanaAccount(db *gorm.DB) (*solana.Wallet, error) {
	if user.PrivateKey != "" {
		account, err := solana.WalletFromPrivateKeyBase58(user.PrivateKey)
		core.Client.RequestAirdrop(
			context.TODO(),
			account.PublicKey(),
			solana.LAMPORTS_PER_SOL*1,
			rpc.CommitmentFinalized,
		)
		return account, err
	}

	account := solana.NewWallet()
	fmt.Println("account private key:", account.PrivateKey)
	fmt.Println("account public key:", account.PublicKey())

	user.PrivateKey = account.PrivateKey.String()
	user.PublicKey = account.PublicKey().String()

	if err := db.Save(user).Error; err != nil {
		return nil, err
	}

	// Airdrop 1 SOL to the new account:
	out, err := core.Client.RequestAirdrop(
		context.TODO(),
		account.PublicKey(),
		solana.LAMPORTS_PER_SOL*1,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		fmt.Println(err)
		return account, err
	}
	fmt.Println("airdrop transaction signature:", out)

	return account, err
}

func (user *User) MakeTransaction(instructions []solana.Instruction) (*solana.Signature, error) {
	wsClient, err := ws.Connect(context.Background(), rpc.DevNet_WS)
	if err != nil {
		return nil, err
	}
	accountFrom := solana.MustPrivateKeyFromBase58(user.PrivateKey)
	recent, err := core.Client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(accountFrom.PublicKey()),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if accountFrom.PublicKey().Equals(key) {
				return &accountFrom
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to sign transaction: %w", err)
	}

	sig, err := confirm.SendAndConfirmTransaction(
		context.TODO(),
		core.Client,
		wsClient,
		tx,
	)
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func (user *User) CreateCard(db *gorm.DB) error {
	cardNumberSuffix, err := core.GetRandomNumberString(12)
	if err != nil {
		return err
	}

	cvv, err := core.GetRandomNumberString(3)
	if err != nil {
		return err
	}

	u, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	user.DebitCard = PaypeerDebitCard{
		ID:         u.String(),
		CardNumber: fmt.Sprintf("4242%s", string(cardNumberSuffix)),
		CardActive: true,
		Cvv:        cvv,
		ExpiryDate: time.Now().Add(4 * 365 * 21 * time.Hour),
	}
	if err := db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

func (user *User) MakeSolTransfer(amount uint64, accountTo solana.PublicKey) (*solana.Signature, error) {
	instructions := []solana.Instruction{
		system.NewTransferInstruction(
			amount,
			solana.MustPublicKeyFromBase58(user.PublicKey),
			accountTo,
		).Build(),
	}
	return user.MakeTransaction(instructions)
}

func (user *User) GetAssociatedTokenAccountBalance(mint solana.PublicKey) (*uint64, *solana.PublicKey, error) {
	payer := solana.MustPrivateKeyFromBase58(user.PrivateKey)
	wallet := payer.PublicKey()
	ata, _, err := solana.FindAssociatedTokenAddress(
		wallet,
		mint,
	)
	if err != nil {
		return nil, nil, err
	}

	_, err = core.Client.GetAccountInfo(
		context.TODO(),
		ata,
	)

	if err != nil && err.Error() == "not found" {
		_, err = user.MakeTransaction([]solana.Instruction{
			associatedtokenaccount.NewCreateInstruction(
				wallet,
				wallet,
				mint,
			).Build(),
		})
	}

	if err != nil && err.Error() != "not found" {
		return nil, nil, err
	}

	// Get and Normalize balance
	balance, err := core.Client.GetTokenAccountBalance(
		context.Background(),
		ata,
		"confirmed",
	)
	if err != nil {
		return nil, nil, err
	}

	balanceString := "0"
	if balanceLength := len(balance.Value.Amount); balanceLength > 7 {
		balanceString = balance.Value.Amount[:balanceLength-7]
	}

	balanceNum, err := strconv.ParseUint(balanceString, 10, 0)
	if err != nil {
		return nil, nil, err
	}

	return &balanceNum, &ata, nil
}

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("develop.db"), &gorm.Config{})
	if err != nil {
		panic("could not connect to db")
	}

	// TODO: handle migrations as a script
	db.AutoMigrate(&User{})
	db.AutoMigrate(&PaypeerDebitCard{})

	return db
}
