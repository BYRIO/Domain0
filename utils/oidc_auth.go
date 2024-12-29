package utils

import (
	"bytes"
	"domain0/config"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func oidcTokenURL() string {
	return config.CONFIG.OIDC.BaseUrl + "/idp/oidc/token"
}
func OIDCUserInfoURL() string {
	return config.CONFIG.OIDC.BaseUrl + "/idp/oidc/me"
}

type OIDCTokenRes struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
	Message     string `json:"message"`
}
type OIDCInfoRes struct {
	Email            string      `json:"email"`
	Identities       *Identities `json:"identities"`
	Error            string      `json:"error"`
	ErrorDescription string      `json:"error_description"`
}

type Identities struct {
	YourCAS *YourCAS `json:"yourcas"`
}

type YourCAS struct {
	UserID  string   `json:"userId"` // studentID
	Details *Details `json:"details"`
}

type Details struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func OIDCRedirectURL() string {
	var buf bytes.Buffer
	buf.WriteString(config.CONFIG.OIDC.BaseUrl + "/idp/oidc/auth")
	v := url.Values{
		"client_id":     {config.CONFIG.OIDC.ClientId},
		"scope":         {"openid email identities"},
		"response_type": {"code"},
		"state":         {"oidc"},
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
		logrus.Error("fetch app token failed : ", tokenRes.Message)
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

	var infoRes OIDCInfoRes
	err = json.Unmarshal(body, &infoRes)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}
	if infoRes.Error != "" {
		logrus.Error("fetch user info failed : ", infoRes.ErrorDescription)
		return AuthInfo{}, errors.New("oidc auth failed")
	}
	return AuthInfo{
		Name:       infoRes.Identities.YourCAS.Details.Name,
		EmployeeID: infoRes.Identities.YourCAS.UserID,
		Email:      infoRes.Email,
	}, nil
}
