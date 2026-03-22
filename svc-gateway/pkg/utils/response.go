package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func translateFieldName(field string) string {
	switch field {
	case "Username":
		return "用户名"
	case "Password":
		return "密码"
	case "Email":
		return "邮箱"
	case "EmailCode":
		return "验证码"
	case "ConfirmedPassword":
		return "确认密码"
	case "Avatar":
		return "头像"
	case "Nickname":
		return "昵称"
	case "Bio":
		return "个人简介"
	case "Website":
		return "个人网站"
	case "Location":
		return "所在地"
	case "UserID":
		return "用户ID"
	default:
		return field
	}
}

func formatValidationError(f validator.FieldError) string {
	field := translateFieldName(f.Field())
	param := f.Param()

	switch f.Tag() {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "required_without":
		return fmt.Sprintf("当%s缺失时，%s不能为空", translateFieldName(param), field)
	case "min":
		return fmt.Sprintf("%s长度不能少于 %s 个字符", field, param)
	case "max":
		return fmt.Sprintf("%s长度不能超过 %s 个字符", field, param)
	case "email":
		return fmt.Sprintf("请填写真实有效的%s", field)
	case "eqfield":
		return fmt.Sprintf("%s与%s不一致", field, translateFieldName(param))
	case "len":
		return fmt.Sprintf("%s长度必须是 %s 位", field, param)
	case "numeric":
		return fmt.Sprintf("%s必须是纯数字", field)
	}
	return fmt.Sprintf("%s填写有误", field)
}

func ErrorResponse(err error) map[string]any {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		var errMsgs []string
		for _, f := range validationErrs {
			errMsgs = append(errMsgs, formatValidationError(f))
		}
		return map[string]any{"error": strings.Join(errMsgs, "; ")}
	}
	return map[string]any{"error": err.Error()}
}

func MessageResponse(message string) map[string]any {
	return map[string]any{"message": message}
}
