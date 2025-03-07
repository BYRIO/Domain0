package services

import (
	"database/sql"
	"domain0/config"
	c "domain0/config"
	db "domain0/database"
	m "domain0/models"
	wm "domain0/models/web"
	"domain0/utils"
	"errors"
	"gorm.io/gorm"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var emailReg = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)

const (
	localsUserName = "user_name"
)

func jwtSign(user m.User) (string, error) {
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.ID,
		"stu_id": user.GetStuId(),
		"name":   user.Name,
		"email":  user.Email,
		"role":   user.Role,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})
	return rawToken.SignedString([]byte(c.CONFIG.JwtKey))
}

func JwtToLocalsWare(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	if user.Valid {
		claims := user.Claims.(jwt.MapClaims)
		c.Locals("sub", uint(claims["sub"].(float64)))
		c.Locals("role", m.UserRole(claims["role"].(float64)))
		c.Locals(localsUserName, claims["name"].(string))
	}
	return c.Next()
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

// @Summary feishu auth enable
// description return true if feishu auth enabled in config.
// Produce json
// @Success 200
// @Router /api/v1/user/feishu/enable [get]
func FeishuAuthEnable(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(
		struct {
			Status int  `json:"status"`
			Enable bool `json:"enable"`
		}{
			Status: fiber.StatusOK,
			Enable: config.CONFIG.Feishu.Enable,
		},
	)
}

// @Summary feishu auth redirect
// @description feishu auth redirect api
// @Produce json
// @Success 302
// @Failure 400 {error}
// @Router /api/v1/user/feishu [get]
func FeishuAuthRedirect(c *fiber.Ctx) error {
	if config.CONFIG.Feishu.Enable {
		return c.Redirect(utils.FeishuRedirectToCodeURL())
	}
	return c.Redirect("/user/login")
}

// @Summary OIDC auth enable
// description return true if OIDC auth enabled in config.
// Produce json
// @Success 200
// @Router /api/v1/user/oidc/enable [get]
func OIDCAuthEnable(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(
		struct {
			Status  int    `json:"status"`
			Enable  bool   `json:"enable"`
			Name    string `json:"name"`
			LogoURL string `json:"logo_url"`
		}{
			Status:  fiber.StatusOK,
			Enable:  config.CONFIG.OIDC.Enable,
			Name:    config.CONFIG.OIDC.Name,
			LogoURL: config.CONFIG.OIDC.LogoURL,
		},
	)
}

// @Summary OIDC auth redirect
// @description OIDC auth redirect api
// @Produce json
// @Success 302
// @Failure 400 {error}
// @Router /api/v1/user/oidc [get]
func OIDCAuthRedirect(c *fiber.Ctx) error {
	if config.CONFIG.OIDC.Enable {
		return c.Redirect(utils.OIDCRedirectURL())
	}
	return c.Redirect("/user/login")
}

// @Summary oauth callback
// @description oauth callback api
// @description user can login with feishu for now
// @Param code query string true "oauth code"
// @Param state query string true "oauth state"
// @Produce json
// @Success 200 {object} wm.User{data=string}
// @Failure 400 {object} wm.User{data=int}
// @Failure 500 {object} wm.User{data=int}
// @Router /api/v1/user/callback [get]
func Callback(c *fiber.Ctx) error {
	state := c.Query("state")
	code := c.Query("code")
	if code == "" || state == "" {
		logrus.Errorf("code is empty")
		return c.Status(fiber.StatusBadRequest).JSON(wm.User{
			Status: fiber.StatusBadRequest,
			Errors: "code is empty",
			Data:   0,
		})
	}
	result := db.DB.Where("state=?", state).First(&m.SSOState{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(wm.User{
				Status: fiber.StatusBadRequest,
				Errors: "state is invalid or expired",
				Data:   0,
			})
		}
		logrus.Error(result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   0,
		})
	}
	db.DB.Where("state=?", state).Delete(&m.SSOState{})
	// get userInfo
	var userInfo utils.AuthInfo
	if strings.HasPrefix(state, "feishu") {
		var err error
		userInfo, err = utils.FeishuGetUserInfo(code)
		if err != nil || userInfo.Email == "" {
			logrus.Errorf("feishu get user info error : %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   0,
			})
		}
	} else if strings.HasPrefix(state, "oidc") {
		var err error
		userInfo, err = utils.OIDCGetUserInfo(code)
		if err != nil || userInfo.Email == "" {
			logrus.Errorf("feishu get user info error : %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   0,
			})
		}
	} else {
		logrus.Errorf("state " + state + " is invalid")
		return c.Status(fiber.StatusBadRequest).JSON(wm.User{
			Status: fiber.StatusBadRequest,
			Errors: "state is invalid",
			Data:   0,
		})
	}

	// check if user exist
	var userObject m.User
	if err := db.DB.Where("email = ?", userInfo.Email).First(&userObject).Error; err != nil {
		// not exist, create user
		userObject = m.User{
			Email:    userInfo.Email,
			Password: "",
			StuId: sql.NullString{
				String: userInfo.EmployeeID,
				Valid:  userInfo.EmployeeID != "",
			},
			Name: userInfo.Name,
		}
		if err := db.DB.Create(&userObject).Error; err != nil {
			logrus.Errorf("create user error : %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   0,
			})
		}
	}

	// generate jwt token
	token, err := jwtSign(userObject)
	if err != nil {
		logrus.Errorf("generate jwt token error : %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(wm.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   0,
		})
	}

	return c.Status(fiber.StatusOK).JSON(wm.User{
		Status: fiber.StatusOK,
		Data:   token,
	})
}
