package services

import (
	c "domain0/config"
	db "domain0/database"
	m "domain0/models"
	wm "domain0/models/web"
	"math/rand"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var emailReg = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)

func jwtSign(user m.User) (string, error) {
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.ID,
		"stu_id": user.GetStuId(),
		"name":   user.Name,
		"email":  user.Email,
		"role":   user.Role,
		"iat":    user.CreatedAt.Unix(),
		"exp":    user.CreatedAt.Add(time.Hour * 72).Unix(),
	})
	return rawToken.SignedString([]byte(c.CONFIG.JwtKey))
}

// @Summary login
// @description login api
// @description user can login with email or stu_id(Not implemented)
// @Param user formData string true "user email or stu_id"
// @Param pass formData string true "user password"
// @Produce json
// @Success 200 {object} wm.User{data=string}
// @Failure 400 {object} wm.User{data=int}
// @Failure 401 {object} wm.User{data=int}
// @Failure 500 {object} wm.User{data=int}
// @Router /api/v1/user/login [post]
// @tags user
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
		// todo: user stu_id login
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

// @Summary register
// @description register api
// @description user can register with email
// @Param email formData string true "user email"
// @Param pass formData string true "user password"
// @Produce json
// @Success 200 {object} wm.User{data=string}
// @Failure 400 {object} wm.User{data=int}
// @Failure 500 {object} wm.User{data=int}
// @Router /api/v1/user/register [post]
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
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
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
