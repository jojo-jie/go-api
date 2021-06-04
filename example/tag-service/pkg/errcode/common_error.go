package errcode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	Success          = NewError(0, "成功")
	Fail             = NewError(10000000, "内部错误")
	InvalidParams    = NewError(10000001, "无效参数")
	Unauthorized     = NewError(10000002, "认证错误")
	NotFound         = NewError(10000003, "没有找到")
	Unknown          = NewError(10000004, "未知")
	DeadlineExceeded = NewError(10000005, "超出最后截止期限")
	AccessDenied     = NewError(10000006, "访问被拒绝")
	LimitExceed      = NewError(10000007, "访问限制")
	MethodNotAllowed = NewError(10000008, "不支持该方法")
)

func TogRPCError(err *Error) error {
	s := status.New(ToRPCCode(err.code), err.msg)
	return s.Err()
}

func ToRPCCode(code int) codes.Code {
	var statusCode codes.Code
	switch code {
	case Fail.Code():
		statusCode = codes.Internal
	case InvalidParams.Code():
		statusCode = codes.InvalidArgument
	case Unauthorized.Code():
		statusCode = codes.Unauthenticated
	case AccessDenied.Code():
		statusCode = codes.PermissionDenied
	case DeadlineExceeded.Code():
		statusCode = codes.DeadlineExceeded
	case NotFound.Code():
		statusCode = codes.NotFound
	case LimitExceed.Code():
		statusCode = codes.ResourceExhausted
	case MethodNotAllowed.Code():
		statusCode = codes.Unimplemented
	default:
		statusCode = codes.Unknown
	}
	return statusCode
}
