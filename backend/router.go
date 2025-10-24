package main

import "github.com/gin-gonic/gin"

func router(r *gin.Engine, hdl *handler) {
	r.POST("/login", hdl.login)
} 