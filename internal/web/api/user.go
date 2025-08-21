package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gowvp/gb28181/internal/conf"
	"github.com/ixugo/goddd/pkg/reason"
	"github.com/ixugo/goddd/pkg/web"
)

type UserAPI struct {
	conf *conf.Bootstrap
}

func NewUserAPI(conf *conf.Bootstrap) UserAPI {
	return UserAPI{
		conf: conf,
	}
}

func RegisterUser(r gin.IRouter, api UserAPI, mid ...gin.HandlerFunc) {
	group := r.Group("/user")
	group.POST("/login", web.WarpH(api.login))
	group.PUT("/user", web.WarpHs(api.updateCredentials, mid...)...)
}

// 登录请求结构体
type loginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录响应结构体
type loginOutput struct {
	Token string `json:"token"`
	User  string `json:"user"`
}

// 登录接口
func (api UserAPI) login(_ *gin.Context, in *loginInput) (*loginOutput, error) {
	// 验证用户名和密码
	if api.conf.Server.Username != "" && api.conf.Server.Password != "" {
		api.conf.Server.Username = "admin"
		api.conf.Server.Password = "admin"
		if in.Username != api.conf.Server.Username || in.Password != api.conf.Server.Password {
			return nil, reason.ErrNameOrPasswd
		}
	}

	data := web.NewClaimsData().SetUsername(in.Username)

	token, err := web.NewToken(data, api.conf.Server.HTTP.JwtSecret, web.WithExpiresAt(time.Now().Add(3*24*time.Hour)))
	if err != nil {
		return nil, reason.ErrServer.SetMsg("生成token失败: " + err.Error())
	}

	return &loginOutput{
		Token: token,
		User:  in.Username,
	}, nil
}

// 修改凭据请求结构体
type updateCredentialsInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 修改凭据接口
func (api UserAPI) updateCredentials(c *gin.Context, in *updateCredentialsInput) (gin.H, error) {
	// 更新配置中的用户名和密码
	api.conf.Server.Username = in.Username
	api.conf.Server.Password = in.Password

	// 写入配置文件
	if err := conf.WriteConfig(api.conf, api.conf.ConfigPath); err != nil {
		return nil, reason.ErrServer.SetMsg("保存配置失败: " + err.Error())
	}

	return gin.H{"msg": "凭据更新成功"}, nil
}
