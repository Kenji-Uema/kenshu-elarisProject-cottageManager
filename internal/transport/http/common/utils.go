package common

import (
	"cottageManager/internal/domain"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ToPeriod(start time.Time, end time.Time) domain.Period {
	return domain.Period{Start: start, End: end}
}

func BindQueryAndUri[T CottageTypeURI | CottageNameURI](c *gin.Context, uri *T, query *DateRangeQuery) error {
	if err := c.ShouldBindUri(uri); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Warn("failed to bind request URI", "error", err, "method", c.Request.Method, "path", c.Request.URL.Path)
		return fmt.Errorf("error binding uri: %w", err)
	}

	if err := c.ShouldBindQuery(query); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Warn("failed to bind query parameters", "error", err, "method", c.Request.Method, "path", c.Request.URL.Path)
		return fmt.Errorf("error binding query: %w", err)
	}

	if query.From.After(query.To) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "From date must be before to date"})
		slog.Warn("invalid date range provided", "from", query.From, "to", query.To, "method", c.Request.Method, "path", c.Request.URL.Path)
		return fmt.Errorf("error binding uri: from date must be before to date")
	}
	return nil
}
