package main

const (
	DataCnxAlreadyOpenStartXfr = "125 Data connection already open, starting transfer\r\n"                            // 打开数据连接，开始传输
	TypeSetOk                  = "200 Type set ok\r\n"                                                                // 连接协议设置成功
	PortOk                     = "200 PORT ok\r\n"                                                                    // 端口设置成功
	CmdOk                      = "200 Command ok\r\n"                                                                 // 命令成功
	FeatResponse               = "211-Features:\r\n  FEAT\r\n  MDTM\r\n  PASV\r\n  SIZE\r\n  TYPE A;I\r\n211 End\r\n" // 系统状态回复
	SysType                    = "215 UNIX Type: L8\r\n"                                                              // 系统类型回复
	FtpServerReady             = "220 FTP Server Ready\r\n"                                                           // 服务就绪
	GoodbyeMsg                 = "221 Goodbye!"                                                                       // 退出FTP
	TxfrCompleteOk             = "226 Data transfer complete\r\n"                                                     // 结束数据连接，数据传输完成
	EnteringPasvMode           = "227 Entering Passive Mode (%s)\r\n"                                                 // 进入被动模式
	PwdResponse                = "257 \"/\"\r\n"                                                                      // 路径名建立
	UsrLoggedInProceed         = "230 User Logged In Proceed\r\n"                                                     // 进入登录过程
	UsrNameOkNeedPass          = "331 Username OK Need Pass\r\n"                                                      // 要求输入密码
	SyntaxErr                  = "500 Syntax Error\r\n"                                                               // 无效命令
	CmdNotImplementd           = "502 Command not implemented\r\n"                                                    // 命令没有被执行/实现
	NotLoggedIn                = "530 Not Logged In\r\n"                                                              // 登录失败
	AuthFailure                = "530 Auth Failure\r\n"                                                               // 认证错误
	AuthFailureTryAgain        = "530 Please login with USER and PASS."                                               // 请以账号和密码登录
	AnonUserDenied             = "550 Anon User Denied\r\n"                                                           // 没有权限或者是文件不存在
)
