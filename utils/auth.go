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
}

type AuthInfo struct {
	// Sub          string `json:"sub"`
	// Name         string `json:"name"`
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
	// EmployeeID   string `json:"employee_id"`
	Email string `json:"enterprise_email"`
	// Mobile       string `json:"mobile"`
}

type FeishuAuthInfoResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"msg"`
	Data    AuthInfo `json:"data"`
}

func FeishuRedirectToCodeURL() string {
	return "https://passport.feishu.cn/suite/passport/oauth/authorize?client_id=" + config.CONFIG.Feishu.AppID +
		"&redirect_uri=" + url.QueryEscape(config.CONFIG.Feishu.RedirectURL) +
		"&response_type=code" +
		"&state=feishu"
}

func feishuRedirectToTokenURL(code string) string {
	return "https://passport.feishu.cn/suite/passport/oauth/token?grant_type=authorization_code&client_id=" + config.CONFIG.Feishu.AppID +
		"&client_secret=" + config.CONFIG.Feishu.AppSecret +
		"&code=" + code +
		"&redirect_uri=" + url.QueryEscape(config.CONFIG.Feishu.RedirectURL)
}

func feishuGetUserInfoURL() string {
	return "https://open.feishu.cn/open-apis/authen/v1/user_info"
}

func FeishuGetUserInfo(code string) (AuthInfo, error) {
	// Query Access token
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	act_req := a.Request()
	act_req.Header.SetMethod("POST")
	act_req.SetRequestURI(feishuRedirectToTokenURL(code))

	if err := a.Parse(); err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	hcode, body, errs := a.Bytes()
	if len(errs) != 0 || hcode != 200 {
		logrus.Error(errs)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	var accessTokenInfo AccessTokenInfo
	err := json.Unmarshal(body, &accessTokenInfo)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	// Query User info
	u := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(u)

	user_req := u.Request()
	user_req.Header.SetMethod("GET")
	user_req.SetRequestURI(feishuGetUserInfoURL())
	user_req.Header.Set("Authorization", "Bearer "+accessTokenInfo.AccessToken)

	if err := u.Parse(); err != nil {
		logrus.Error(err)
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
		logrus.Error(feishuInfoResponse.Message)
		return AuthInfo{}, errors.New("feishu auth failed")
	}

	return feishuInfoResponse.Data, nil
}
