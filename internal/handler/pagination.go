package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return page, pageSize
}

func parseUint64Param(c *gin.Context, name string) (uint64, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return value, true
}

func parseUint64Query(c *gin.Context, name string) uint64 {
	value, _ := strconv.ParseUint(c.Query(name), 10, 64)
	return value
}
