package services

import (
	c "domain0/config"
	db "domain0/database"
	m "domain0/models"
	wm "domain0/models/web"
	"math/rand"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

var emailReg = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)

func jwtSign(user m.User) (string, error) {
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.ID,
		"stu_id": user.StuId,
		"name":   user.Name,
		"email":  user.Email,
		"role":   user.Role,
		"iat":    user.CreatedAt.Unix(),
		"exp":    user.CreatedAt.Add(time.Hour * 72).Unix(),
	})
	return rawToken.SignedString(c.CONFIG.JwtKey)
}

func Login(c *fiber.Ctx) error {
	randtag := rand.Intn(1919810)
	user := c.FormValue("user")
	pass := c.FormValue("pass")
	if user == "" || pass == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user or pass is empty",
		})
	}

	var userObject m.User
	if !emailReg.MatchString(user) {
		// todo: user
		return c.Status(fiber.StatusNotImplemented).JSON(wm.User{
			Status: fiber.StatusNotImplemented,
			Errors: "not implemented",
			Data:   randtag,
		})
	} else {
		// check if user exist
		if err := db.DB.Where("email = ?", user).First(&userObject).Error; err != nil {
			logrus.Warnf("%d login error : %v", randtag, err)
			return c.Status(fiber.StatusUnauthorized).JSON(wm.User{
				Status: fiber.StatusUnauthorized,
				Errors: "user not found or password error",
				Data:   randtag,
			})
		}

		// check if password is correct
		if err := bcrypt.CompareHashAndPassword([]byte(userObject.Password), []byte(pass)); err != nil {
			logrus.Warnf("%d login error : %v", randtag, err)
			return c.Status(fiber.StatusUnauthorized).JSON(wm.User{
				Status: fiber.StatusUnauthorized,
				Errors: "user not found or password error",
				Data:   randtag,
			})
		}

		// generate jwt token
		token, err := jwtSign(userObject)
		if err != nil {
			logrus.Errorf("%d login error : %v", randtag, err)
			return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   randtag,
			})
		}

		// set localstorage, not cookie
		return c.Status(fiber.StatusOK).JSON(wm.User{
			Status: fiber.StatusOK,
			Data:   token,
		})
	}
}

func Register(c *fiber.Ctx) error {
	randtag := rand.Intn(1919810)
	email := c.FormValue("email")
	pass := c.FormValue("pass")
	if pass == "" || !emailReg.MatchString(email) {
		return c.Status(fiber.StatusBadRequest).JSON(wm.User{
			Status: fiber.StatusBadRequest,
			Errors: "user or pass is not valid",
			Data:   randtag,
		})
	}

	// hash password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MaxCost)
	if err != nil {
		logrus.Errorf("%d register error : %v", randtag, err)
		return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   randtag,
		})
	}
	userObject := m.User{
		Email:    email,
		Password: string(hashedPass),
	}
	if err := db.DB.Create(&userObject).Error; err != nil {
		logrus.Errorf("%d register error : %v", randtag, err)
		return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   randtag,
		})
	}

	// generate jwt token
	token, err := jwtSign(userObject)
	if err != nil {
		logrus.Errorf("%d register error : %v", randtag, err)
		return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   randtag,
		})
	}

	// set localstorage, not cookie
	return c.Status(fiber.StatusOK).JSON(wm.User{
		Status: fiber.StatusOK,
		Data:   token,
	})
}
