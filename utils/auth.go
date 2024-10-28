package utils

import (
	"domain0/config"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AccessTokenInfo struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

type AuthInfo struct {
	// Sub          string `json:"sub"`
	Name         string `json:"name"`
	// Picture      string `json:"picture"`
	// OpenID       string `json:"open_id"`
	// UnionID      string `json:"union_id"`
	// EnName       string `json:"en_name"`
	// TenantKey    string `json:"tenant_key"`
	// AvatarURL    string `json:"avatar_url"`
	// AvatarThumb  string `json:"avatar_thumb"`
	// AvatarMiddle string `json:"avatar_middle"`
	// AvatarBig    string `json:"avatar_big"`
	// UserID       string `json:"user_id"`
	EmployeeID   string `json:"employee_no"`
	Email string `json:"enterprise_email"`
	// Mobile       string `json:"mobile"`
}

type FeishuGenericResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    T      `json:"data"`
}

type FeishuAccessTokenInfoResponse = FeishuGenericResponse[AccessTokenInfo]

type FeishuAuthInfoResponse = FeishuGenericResponse[AuthInfo]

type FeishuAppAccessTokenInfoResponse struct {
	AppAccessToken    string `json:"app_access_token"`
	Code              int    `json:"code"`
	Expire            int    `json:"expire"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
}

func FeishuRedirectToCodeURL() string {
	return "https://open.feishu.cn/open-apis/authen/v1/authorize?app_id=" + config.CONFIG.Feishu.AppID +
		"&redirect_uri=" + url.QueryEscape(config.CONFIG.Feishu.RedirectURL) +
		"&state=feishu"
}

func feishuAppAccessTokenURL() string {
	return "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal?app_id=" + config.CONFIG.Feishu.AppID +
		"&app_secret=" + config.CONFIG.Feishu.AppSecret
}

func feishuRedirectToTokenURL(code string) string {
	return "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token?grant_type=authorization_code&code=" + code
}

func feishuGetUserInfoURL() string {
	return "https://open.feishu.cn/open-apis/authen/v1/user_info"
}

func FeishuGetUserInfo(code string) (AuthInfo, error) {
	// Query App access token
	t := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(t)

	app_req := t.Request()
	app_req.Header.SetMethod("POST")
	app_req.SetRequestURI(feishuAppAccessTokenURL())
	if err := t.Parse(); err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	hcode, body, errs := t.Bytes()
	if len(errs) != 0 || hcode != 200 {
		logrus.Error("fetch app token failed : ", string(body), errs)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	var feishuAppAccessTokenInfo FeishuAppAccessTokenInfoResponse
	err := json.Unmarshal(body, &feishuAppAccessTokenInfo)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}
	if feishuAppAccessTokenInfo.Code != 0 {
		logrus.Error("fetch app token failed : ", feishuAppAccessTokenInfo.Msg)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	// Query Access token
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	act_req := a.Request()
	act_req.Header.SetMethod("POST")
	act_req.SetRequestURI(feishuRedirectToTokenURL(code))
	act_req.Header.Set("Authorization", "Bearer "+feishuAppAccessTokenInfo.AppAccessToken)
	if err := a.Parse(); err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	hcode, body, errs = a.Bytes()
	if len(errs) != 0 || hcode != 200 {
		logrus.Error("fetch auth token failed : ", string(body), errs)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	var accessTokenInfo FeishuAccessTokenInfoResponse
	err = json.Unmarshal(body, &accessTokenInfo)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}
	if accessTokenInfo.Code != 0 {
		logrus.Error("fetch user info failed : ", accessTokenInfo.Message)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	// Query User info
	u := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(u)

	user_req := u.Request()
	user_req.Header.SetMethod("GET")
	user_req.SetRequestURI(feishuGetUserInfoURL())
	user_req.Header.Set("Authorization", "Bearer "+accessTokenInfo.Data.AccessToken)

	if err := u.Parse(); err != nil {
		logrus.Error("fetch user info failed : ", string(body), err)
		return AuthInfo{}, err
	}

	hcode, body, errs = u.Bytes()
	if len(errs) != 0 || hcode != 200 {
		logrus.Error(errs)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	var feishuInfoResponse FeishuAuthInfoResponse
	err = json.Unmarshal(body, &feishuInfoResponse)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}
	if feishuInfoResponse.Code != 0 {
		logrus.Error("fetch user info failed : ", feishuInfoResponse.Message)
		return AuthInfo{}, errors.New("feishu auth failed")
	}
	return feishuInfoResponse.Data, nil
}
