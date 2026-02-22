package utils

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// Response struct representing a standard payload.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	TraceID string `json:"trace_id"`
}

// SendResponse uses Fiber to return a standard envelope format.
func SendResponse(c fiber.Ctx, code int, data any) error {
	traceID := requestid.FromContext(c)

	isSuccess := code >= 200 && code < 300
	msg := "Success"
	if !isSuccess {
		msg = "Failed"
	}

	return c.Status(code).JSON(Response{
		Success: isSuccess,
		Message: msg,
		Data:    data,
		TraceID: traceID,
	})
}
