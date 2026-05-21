package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 统一 API 响应 envelope。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// JSON 写入标准 envelope，HTTP 状态码与 code 一致。
func JSON(c *gin.Context, httpStatus, code int, message string, data any) {
	c.JSON(httpStatus, Body{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// OK 成功响应，message 默认 success。
func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, http.StatusOK, "success", data)
}

// OKMessage 成功响应，自定义 message。
func OKMessage(c *gin.Context, message string, data any) {
	JSON(c, http.StatusOK, http.StatusOK, message, data)
}

// Fail 通用失败响应。
func Fail(c *gin.Context, httpStatus int, message string) {
	JSON(c, httpStatus, httpStatus, message, nil)
}

// BadRequest 400。
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, message)
}

// Unauthorized 401。
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, message)
}

// Forbidden 403。
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, message)
}

// NotFound 404。
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, message)
}

// InternalError 500。
func InternalError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, message)
}

// PageData 统一分页 data 结构。
type PageData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// Page 分页成功响应。
func Page(c *gin.Context, list any, total int64, page, pageSize int) {
	if list == nil {
		list = []any{}
	}
	OK(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
