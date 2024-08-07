package controller

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gofiber/fiber/v2"
)

type Session struct {
	Nonce string
}

func MetamaskLogin(c *fiber.Ctx) error {
	// Generate a nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	// Save the nonce and associate it with the user's session
	session := c.Locals("session").(*Session)
	session.Nonce = hex.EncodeToString(nonce)

	// Send the nonce to the user
	return c.JSON(fiber.Map{
		"nonce": session.Nonce,
	})
}

// TODO: metamask 로그인
//func MetamaskCallback(c *fiber.Ctx) error {
//	// Get the signed nonce and public address from the request
//	signedNonce := c.FormValue("signedNonce")
//	publicAddress := c.FormValue("publicAddress")
//
//	// Verify the signed nonce
//	sig := strings.TrimPrefix(signedNonce, "0x")
//	if len(sig) != 130 {
//		return c.Status(fiber.StatusBadRequest).SendString("Invalid signature")
//	}
//
//	sigBytes, err := hex.DecodeString(sig)
//	if err != nil {
//		return c.Status(fiber.StatusBadRequest).SendString("Invalid signature")
//	}
//
//	r, s := sigBytes[:32], sigBytes[32:64]
//	v := sigBytes[64]
//	if v < 27 {
//		v += 27
//	}
//
//	session := c.Locals("session").(*Session)
//	nonce, err := hex.DecodeString(session.Nonce)
//	if err != nil {
//		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
//	}
//
//	sig := make([]byte, 65)
//	copy(sig[32-len(r):32], r)
//	copy(sig[64-len(s):64], s)
//	sig[64] = v
//	if sig[64] < 27 {
//		sig[64] += 27
//	}
//
//	pubKey, err := crypto.SigToPub(crypto.Keccak256Hash(nonce).Bytes(), sig)
//	if err != nil {
//		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
//	}
//	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
//	if !strings.EqualFold(recoveredAddr.Hex(), publicAddress) {
//		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
//	}
//	//
//
//	// The user is authenticated, create or update the user in the database
//	var user models.User
//	database.DB.Where("public_address = ?", publicAddress).First(&user)
//	if user.ID == 0 {
//		user = models.User{
//			PublicAddress: publicAddress,
//		}
//		database.DB.Create(&user)
//	} else {
//		database.DB.Save(&user)
//	}
//
//	return c.JSON(fiber.Map{
//		"user":    user,
//		"message": "User logged in successfully",
//	})
//}
