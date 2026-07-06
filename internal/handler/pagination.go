// 管理面通用分页与路径参数解析。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parsePage 解析 page / page_size，非法值由 service 层 normalizePage 兜底。
func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return page, pageSize
}

// parseUint64Param 解析路径参数，0 或非法视为无效。
func parseUint64Param(c *gin.Context, name string) (uint64, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return value, true
}

// parseUint64Query 解析可选 query 参数，0 表示不过滤。
func parseUint64Query(c *gin.Context, name string) uint64 {
	value, _ := strconv.ParseUint(c.Query(name), 10, 64)
	return value
}
