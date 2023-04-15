package utils

import (
	"domain0/config"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AccessTokenInfo struct {
	AccessToken      string
	TokenType        string
	ExpiresIn        int
	RefreshToken     string
	RefreshExpiresIn string
}

type AuthInfo struct {
	Sub          string
	Name         string
	Picture      string
	OpenID       string
	UnionID      string
	EnName       string
	TenantKey    string
	AvatarURL    string
	AvatarThumb  string
	AvatarMiddle string
	AvatarBig    string
	UserID       string
	EmployeeID   string
	Email        string
	Mobile       string
}

func FeishuRedirectToCodeURL() string {
	return "https://passport.feishu.cn/suite/passport/oauth/authorize?client_id=" + config.CONFIG.Feishu.AppID +
		"&redirect_uri=" + config.CONFIG.Feishu.RedirectURL +
		"&response_type=code" +
		"&state=feishu"
}

func feishuRedirectToTokenURL(code string) string {
	return "https://passport.feishu.cn/suite/passport/oauth/token?grant_type=authorization_code&client_id=" + config.CONFIG.Feishu.AppID +
		"&client_secret=" + config.CONFIG.Feishu.AppSecret +
		"&code=" + code +
		"&redirect_uri=" + config.CONFIG.Feishu.RedirectURL
}

func feishuGetUserInfoURL() string {
	return "https://passport.feishu.cn/suite/passport/oauth/userinfo"
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
	if errs != nil || hcode != 200 {
		logrus.Error(errs)
		return AuthInfo{}, errs[0]
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
	if errs != nil || hcode != 200 {
		logrus.Error(errs)
		return AuthInfo{}, errs[0]
	}

	var feishuInfo AuthInfo
	err = json.Unmarshal(body, &feishuInfo)
	if err != nil {
		logrus.Error(err)
		return AuthInfo{}, err
	}

	return feishuInfo, nil
}
