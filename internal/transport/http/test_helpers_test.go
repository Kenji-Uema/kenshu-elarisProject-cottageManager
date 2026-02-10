package http

import "github.com/gin-gonic/gin"

func setupGin() {
	gin.SetMode(gin.TestMode)
}

type assertError string

func (e assertError) Error() string { return string(e) }

func assertAnyError() error { return assertError("any error") }
