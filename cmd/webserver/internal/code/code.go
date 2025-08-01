package code

import "github.com/mangohow/cloud-ide/pkg/serialize"

const (
	QuerySuccess = iota + 10
	QueryFailed

	LoginSuccess
	LoginFailed
	LoginUserDeleted
	LoginUserNotExist
	LoginPasswordIncorrect

	SpaceCreateSuccess
	SpaceCreateFailed
	SpaceCreateNameDuplicate
	SpaceCreateReachMaxCount

	SpaceStartSuccess
	SpaceStartFailed

	SpaceDeleteSuccess
	SpaceDeleteFailed
	SpaceDeleteIsRunning

	SpaceStopSuccess
	SpaceStopFailed
	SpaceStopIsNotRunning

	UserNameAvailable
	UserNameUnavailable
	UserSendValidateCodeSuccess
	UserSendValidateCodeFailed
	UserEmailCodeInvalid
	UserEmailInvalid
	UserUsernameExist
	UserRegisterSuccess
	UserRegisterFailed
	UserEmailCodeIncorrect
	UserEmailAlreadyInUse
	UserNameLengthInvalid

	UserEmailNotExists
	UserPasswordLengthInvalid
	UserResetPasswordFailed

	SpaceStartNotExist
	SpaceOtherSpaceIsRunning

	SpaceNameModifySuccess
	SpaceNameModifyFailed
	SpaceAlreadyExist
	SpaceNotFound
	ResourceExhausted
	
	// 支付相关错误码
	PaymentSuccess
	PaymentFailed
	PaymentCallbackSuccess
	PaymentCallbackFailed
)

type UserStatus uint32

const (
	StatusNormal UserStatus = iota
	StatusDeleted
)

var messageForCode = map[int]string{
	QuerySuccess:                "查询成功",
	QueryFailed:                 "查询失败",
	LoginSuccess:                "登录成功",
	LoginFailed:                 "登录失败",
	LoginUserDeleted:            "用户已注销",
	LoginUserNotExist:           "该用户不存在",
	LoginPasswordIncorrect:      "密码错误",
	SpaceCreateSuccess:          "创建成功",
	SpaceCreateFailed:           "创建失败",
	SpaceCreateNameDuplicate:    "不能和已有工作空间名称重复",
	SpaceCreateReachMaxCount:    "达到最大工作空间创建上限,请删除其它工作空间后重试",
	SpaceStartSuccess:           "工作空间启动成功",
	SpaceStartFailed:            "工作空间启动失败,请重试",
	SpaceDeleteSuccess:          "删除工作空间成功",
	SpaceDeleteFailed:           "删除工作空间失败",
	SpaceDeleteIsRunning:        "无法删除正在运行的工作空间,请先停止运行",
	SpaceStopSuccess:            "停止工作空间成功",
	SpaceStopFailed:             "停止工作空间失败",
	SpaceStopIsNotRunning:       "工作空间未运行",
	UserNameAvailable:           "用户名可用",
	UserNameUnavailable:         "用户名不可用",
	UserSendValidateCodeFailed:  "验证码发送失败,请重试",
	UserSendValidateCodeSuccess: "验证码发送成功,五分钟内有效",
	UserEmailCodeInvalid:        "邮箱验证码不合法",
	UserEmailInvalid:            "邮箱格式不正确",
	UserUsernameExist:           "用户名已存在",
	UserRegisterSuccess:         "注册成功",
	UserRegisterFailed:          "注册失败",
	UserEmailCodeIncorrect:      "邮箱验证码不正确",
	UserEmailAlreadyInUse:       "该邮箱已经被注册",
	UserNameLengthInvalid:       "用户名长度必须在3-10个字符之间",
	UserEmailNotExists:          "邮箱不存在",
	UserPasswordLengthInvalid:   "密码长度必须在6-20个字符之间",
	UserResetPasswordFailed:     "重置密码失败",
	SpaceStartNotExist:          "工作空间不存在",
	SpaceOtherSpaceIsRunning:    "检测到有其它工作空间正在运行,请先停止正在运行的工作空间",
	SpaceNameModifySuccess:      "名称修改成功",
	SpaceNameModifyFailed:       "名称修改失败",
	SpaceAlreadyExist:           "工作空间已存在",
	SpaceNotFound:               "未找到该工作空间",
	ResourceExhausted:           "资源不足,无法启动工作空间",
	PaymentSuccess:              "支付成功",
	PaymentFailed:               "支付失败",
	PaymentCallbackSuccess:      "支付回调处理成功",
	PaymentCallbackFailed:       "支付回调处理失败",
}

func GetMessage(code int) string {
	return messageForCode[code]
}

func init() {
	serialize.SetCodeMessager(GetMessage)
}