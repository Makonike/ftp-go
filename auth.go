package main

type AuthUser struct {
	username string
	password string
	valid    bool // 是否合法
}

// 用户登录验证身份
func handleLogin(msg string, user *AuthUser) string {
	// 解析命令，获取参数
	cmd, args, err := parseCommand(msg)
	if err != nil {
		return SyntaxErr
	}
	switch {
	case cmd == "USER" && args == "":
		return AnonUserDenied
	case cmd == "USER" && args != "":
		user.username = args
		return UsrNameOkNeedPass
	case cmd == "PASS" && args == "":
		return SyntaxErr
	case cmd == "PASS" && args != "" && user.username != "":
		user.password = args
	}

	user.Authenticate()
	if user.valid {
		return UsrLoggedInProceed
	} else {
		user.username = ""
		user.password = ""
		return AuthFailureTryAgain
	}
}

// Authenticate will auth the user set valid
func (u *AuthUser) Authenticate() {
	u.valid = true
}
