package gbs

import (
	"fmt"
	"net/http"

	"github.com/gowvp/gb28181/pkg/gbs/sip"
)

type GB28181API struct{}

func (g GB28181API) handlerRegister(ctx *sip.Context) {
	fromUser, ok := parserDevicesFromReqeust(ctx.Request)
	if !ok {
		return
	}

	if len(fromUser.DeviceID) < 18 {
		ctx.String(http.StatusBadRequest, "device id too short")
		return
	}

	// 判断是否存在授权字段
	if hdrs := ctx.Request.GetHeaders("Authorization"); len(hdrs) > 0 {
		// user := Devices{DeviceID: fromUser.DeviceID}
		// if err := db.Get(db.DBClient, &user); err == nil {
		// 	if !user.Regist {
		// 		// 如果数据库里用户未激活，替换user数据
		// 		// fromUser.ID = user.ID
		// 		fromUser.Name = user.Name
		// 		fromUser.PWD = user.PWD
		// 		user = fromUser
		// 	}
		// 	user.addr = fromUser.addr
		// 	authenticateHeader := hdrs[0].(*sip.GenericHeader)
		// 	auth := sip.AuthFromValue(authenticateHeader.Contents)
		// 	auth.SetPassword(user.PWD)
		// 	auth.SetUsername(user.DeviceID)
		// 	auth.SetMethod(string(req.Method()))
		// 	auth.SetURI(auth.Get("uri"))
		// 	if auth.CalcResponse() == auth.Get("response") {
		// 		// 验证成功
		// 		// 记录活跃设备
		// 		user.source = fromUser.source
		// 		user.addr = fromUser.addr
		// 		_activeDevices.Store(user.DeviceID, user)
		// 		if !user.Regist {
		// 			// 第一次激活，保存数据库
		// 			user.Regist = true
		// 			db.DBClient.Save(&user)
		// 			// logrus.Infoln("new user regist,id:", user.DeviceID)
		// 		}
		// 		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		// 		// 注册成功后查询设备信息，获取制作厂商等信息
		// 		go notify(notifyDevicesRegister(user))
		// 		go sipDeviceInfo(fromUser)
		// 		return
		// 	}
		// }
	}
	resp := sip.NewResponseFromRequest("", ctx.Request, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
	resp.AppendHeader(&sip.GenericHeader{HeaderName: "WWW-Authenticate", Contents: fmt.Sprintf("Digest nonce=\"%s\", algorithm=MD5, realm=\"%s\",qop=\"auth\"", sip.RandString(32), _sysinfo.Region)})
	ctx.Tx.Respond(resp)
}
