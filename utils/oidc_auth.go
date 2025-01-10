package utils

import (
	"bytes"
	"domain0/config"
	"domain0/database"
	"domain0/models"
	"encoding/json"
	"errors"
	"github.com/PaesslerAG/jsonpath"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func oidcTokenURL() string {
	return config.CONFIG.OIDC.TokenURL
}
func OIDCUserInfoURL() string {
	return config.CONFIG.OIDC.UserInfoURL
}

type OIDCTokenRes struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

func OIDCRedirectURL() string {
	var buf bytes.Buffer
	buf.WriteString(config.CONFIG.OIDC.AuthURL)
	state := "oidc" + uuid.New().String()
	ssoState := models.SSOState{
		State:       state,
		ExpiredTime: time.Now().Add(60 * time.Second),
	}
	database.DB.Create(&ssoState)
	v := url.Values{
		"client_id":     {config.CONFIG.OIDC.ClientId},
		"scope":         {config.CONFIG.OIDC.Scope},
		"response_type": {"code"},
		"state":         {state},
		"redirect_uri":  {config.CONFIG.OIDC.RedirectUrl},
	}
	buf.WriteByte('?')
	buf.WriteString(v.Encode())
	return buf.String()
}

func OIDCGetUserInfo(code string) (AuthInfo, error) {
	// Query App access token
	t := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(t)

	app_req := t.Request()
	app_req.Header.SetMethod("POST")
	app_req.SetRequestURI(oidcTokenURL())

	data := url.Values{}
	data.Set("client_id", config.CONFIG.OIDC.ClientId)
	data.Set("client_secret", config.CONFIG.OIDC.AppSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.CONFIG.OIDC.RedirectUrl)
	formData := data.Encode()

	app_req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app_req.SetBody([]byte(formData))
	if err := t.Parse(); err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	hcode, body, errs := t.Bytes()
	if len(errs) != 0 || hcode != 200 {
		logrus.Error("fetch app token failed : ", string(body), errs)
		return AuthInfo{}, errors.New("oidc auth failed")
	}

	var tokenRes OIDCTokenRes
	err := json.Unmarshal(body, &tokenRes)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}
	if tokenRes.Error != "" {
		logrus.Error("fetch app token failed : ", tokenRes.Error)
		return AuthInfo{}, errors.New("oidc auth failed")
	}

	// Query User info
	u := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(u)

	user_req := u.Request()
	user_req.Header.SetMethod("GET")
	user_req.SetRequestURI(OIDCUserInfoURL())
	user_req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)

	if err := u.Parse(); err != nil {
		logrus.Error("fetch user info failed : ", string(body), err)
		return AuthInfo{}, err
	}

	hcode, body, errs = u.Bytes()
	if len(errs) != 0 || hcode != 200 {
		logrus.Error(errs)
		return AuthInfo{}, errors.New("oidc auth failed")
	}

	res := interface{}(nil)
	err = json.Unmarshal(body, &res)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	var authInfo AuthInfo
	var aggregateErr error

	// Extract Name
	name, err := jsonpath.Get(config.CONFIG.OIDC.InfoPath.Name, res)
	if err != nil {
		logrus.Error(err)
		aggregateErr = errors.Join(aggregateErr, err)
	} else if nameStr, ok := name.(string); ok {
		authInfo.Name = nameStr
	}

	// Extract ID
	id, err := jsonpath.Get(config.CONFIG.OIDC.InfoPath.Id, res)
	if err != nil {
		logrus.Error(err)
		aggregateErr = errors.Join(aggregateErr, err)
	} else if idStr, ok := id.(string); ok {
		authInfo.EmployeeID = idStr
	}

	// Extract Email
	email, err := jsonpath.Get(config.CONFIG.OIDC.InfoPath.Email, res)
	if err != nil {
		logrus.Error(err)
		aggregateErr = errors.Join(aggregateErr, err)
	} else if emailStr, ok := email.(string); ok {
		authInfo.Email = emailStr
	}

	// Extract Error
	errorField, err := jsonpath.Get(config.CONFIG.OIDC.InfoPath.Error, res)
	if err == nil && errorField != "" { // response contains an error field
		if errorStr, ok := errorField.(string); ok {
			aggregateErr = errors.Join(aggregateErr, errors.New(errorStr))
		} else {
			logrus.Error(errorStr)
			aggregateErr = errors.Join(aggregateErr, errors.New(errorStr))
		}
	} else if err != nil && !strings.HasPrefix(err.Error(), "unknown key") {
		// error message might not exist, which means reeronse OK
		aggregateErr = errors.Join(aggregateErr, err)
	}

	// Return all errors if any occurred
	if aggregateErr != nil {
		logrus.Error(aggregateErr)
		return AuthInfo{}, aggregateErr
	}
	// Return the successfully extracted AuthInfo
	return authInfo, nil
}
