package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/pkg/serialize"
)

type Handler func(ctx *gin.Context) *serialize.Response

func HandlerAdapter(h Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := h(ctx)
		if r != nil {
			ctx.Header("Content-Type", "application/json; charset=utf-8")
			ctx.JSON(r.HttpStatus, &r.R)
		}

		serialize.PutResponse(r)
	}
}
