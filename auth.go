package main

import "log"

type AuthUser struct {
	Id       int64
	Username string `xorm:"varchar(255)"`
	Password string `xorm:"varchar(255)"`
	valid    bool   // 是否合法
}

// 用户登录验证身份
func handleLogin(msg string, user *AuthUser) string {
	// 解析命令，获取参数
	cmd, args, err := ParseCommand(msg)
	if err != nil {
		return SyntaxErr
	}
	switch {
	case cmd == "USER" && args == "":
		return AnonUserDenied
	case cmd == "USER" && args != "":
		user.Username = args
		return UsrNameOkNeedPass
	case cmd == "PASS" && args == "":
		return SyntaxErr
	case cmd == "PASS" && args != "" && user.Username != "":
		user.Password = args
	}

	user.Authenticate()
	if user.valid {
		return UsrLoggedInProceed
	} else {
		user.Username = ""
		user.Password = ""
		return AuthFailureTryAgain
	}
}

// Authenticate will auth the user set valid
func (u *AuthUser) Authenticate() {
	var user AuthUser
	// 验证
	_, err := adapter.Engine.Table("auth_user").Where("username = ?", u.Username).Get(&user)
	if err != nil {
		log.Printf("get user error %s", err)
		return
	}
	if u.Password == user.Password {
		u.valid = true
		return
	}
	u.valid = false
}
