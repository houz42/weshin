package useroauth

// wechat user oauth api
// https://mp.weixin.qq.com/wiki/ 微信网页开发/微信网页授权

import (
	"bytes"
	"log"
	"net/url"

	"github.com/MenInBack/weshin/wx"
)

const (
	oAuthPath        = "https://open.weixin.qq.com/connect/oauth2/authorize"
	accessTokenPath  = "https://api.weixin.qq.com/sns/oauth2/access_token"
	refreshTokenPath = "https://api.weixin.qq.com/sns/oauth2/refresh_token"
	verifyTokenPath  = "https://api.weixin.qq.com/sns/auth"
	userinfoPath     = "https://api.weixin.qq.com/sns/userinfo"
)

const (
	OAuthScopeBase     = "snsapi_base"
	OAUthScopeUserInfo = "snsapi_userinfo"
)

// JumpToAuth compose jump uri for user authorization.
// callback to redirectURI should be handled by caller of this package
// https://open.weixin.qq.com/connect/oauth2/authorize?appid=APPID&redirect_uri=REDIRECT_URI&response_type=code&scope=SCOPE&state=STATE#wechat_redirect
func (o *OAuth) JumpToAuth(scope, redirectURI, state string) (jumpURL string, err error) {
	u := bytes.NewBufferString(oAuthPath)
	u.WriteString("?appid=")
	u.WriteString(o.AppID)
	u.WriteString("&redirect_uri=")
	u.WriteString(url.QueryEscape(redirectURI))
	u.WriteString("&response_type=code")
	u.WriteString("&scope=")
	u.WriteString(scope)
	u.WriteString("&state=")
	u.WriteString(state)
	if o.Mode == wx.ModeComponent {
		u.WriteString("&component_appid=")
		u.WriteString(o.ComponentID)
	}
	u.WriteString("#wechat_redirect")

	log.Print("jump uri for authorization: ", u.String())
	return u.String(), nil
}

// GrantAuthorizeToken grant access token for user authorization
// code is in callback request url after user agreed for oauth
// https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
func (o *OAuth) GrantAuthorizeToken(code string, timeout int) (token *UserAccessToken, err error) {
	log.Print("authorizing code: ", code)
	var parameters []wx.QueryParameter
	switch o.Mode {
	case wx.ModeComponent:
		parameters = []wx.QueryParameter{
			{"appid", o.AppID},
			{"code", code},
			{"grant_type", wx.GrantTypeAuthorize},
			{"component_appid", o.ComponentID},
			{"component_access_token", o.server.GetAccessToken()},
		}
	case wx.ModeMP:
		parameters = []wx.QueryParameter{
			{"appid", o.AppID},
			{"secret", o.secret},
			{"code", code},
			{"grant_type", wx.GrantTypeAuthorize},
		}
	}
	req := wx.HttpClient{
		Path:       accessTokenPath,
		Timeout:    timeout,
		Parameters: parameters,
	}

	token = new(UserAccessToken)
	err = req.Get(token)
	if err != nil {
		log.Print("authorize code failed: ", err)
		return nil, err
	}

	return token, nil
}

// RefreshAuthorizeToken refresh access token for user authorization
// https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
func (o *OAuth) RefreshAuthorizeToken(refreshToken string, timeout int) (token *UserAccessToken, err error) {
	var parameters []wx.QueryParameter
	switch o.Mode {
	case wx.ModeComponent:
		parameters = []wx.QueryParameter{
			{"appid", o.AppID},
			{"grant_type", wx.GrantTypeRefresh},
			{"refresh_token", refreshToken},
			{"component_appid", o.ComponentID},
			{"component_access_token", o.server.GetAccessToken()},
		}
	case wx.ModeMP:
		parameters = []wx.QueryParameter{
			{"appid", o.AppID},
			{"grant_type", wx.GrantTypeRefresh},
			{"refresh_token", refreshToken},
		}
	}

	req := wx.HttpClient{
		Path:       refreshTokenPath,
		Timeout:    timeout,
		Parameters: parameters,
	}

	token = new(UserAccessToken)
	err = req.Get(token)
	if err != nil {
		log.Print("refresh token failed: ", err)
		return nil, err
	}

	return token, err
}

// GetUserInfo get authorized user info
// token is user access token granted earlier, not access token of mp account or component
// https://api.weixin.qq.com/sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID&lang=zh_CN
func GetUserInfo(openID, token, lang string, timeout int) (info *UserInfo, err error) {
	if lang == "" {
		lang = wx.LangCN
	} else if lang != wx.LangCN && lang != wx.LangTW && lang != wx.LangEN {
		return nil, wx.ParameterError{InvalidParameter: "lang"}
	}

	req := wx.HttpClient{
		Path:    userinfoPath,
		Timeout: timeout,
		Parameters: []wx.QueryParameter{
			{"access_token", token},
			{"openid", openID},
			{"lang", lang},
		},
	}

	info = new(UserInfo)
	err = req.Get(info)
	if err != nil {
		log.Print("query user info failed: ", err)
		return nil, err
	}

	return info, err
}

// VerifyAuthorizeToken validates user access token
// https://api.weixin.qq.com/sns/auth?access_token=ACCESS_TOKEN&openid=OPENID
func (o *OAuth) VerifyAuthorizeToken(openID, token string, timeout int) (valid bool, err error) {
	req := wx.HttpClient{
		Path:    verifyTokenPath,
		Timeout: timeout,
		Parameters: []wx.QueryParameter{
			{"access_token", token},
			{"openid", openID},
		},
	}

	err = req.Get(nil)
	if err != nil {
		log.Print("verify user access token failed: ", err)
		return false, err
	}

	return true, nil
}
