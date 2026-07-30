package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func currentUserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	id, _ := v.(uint)
	return id
}

func paramID(c *gin.Context) (uint, bool) {
	return paramUint(c, "id")
}

func paramUint(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

func queryUintPtr(c *gin.Context, key string) *uint {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	val := uint(v)
	return &val
}
