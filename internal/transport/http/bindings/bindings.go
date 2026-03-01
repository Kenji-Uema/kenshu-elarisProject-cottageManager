package bindings

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

func BindQueryAndUri[T CottageTypeURI | CottageNameURI](c *gin.Context, uri *T, query *DateRangeQuery) map[string]any {
	if err := c.ShouldBindUri(uri); err != nil {
		return gin.H{
			"error":   "invalid path variables",
			"details": buildErrorMap(err),
		}
	}

	if err := c.ShouldBindQuery(query); err != nil {
		return gin.H{
			"error":   "invalid query params",
			"details": buildErrorMap(err),
		}
	}

	if query.From.After(query.To) {
		return gin.H{"error": "invalid period, end before start"}
	}

	return nil
}

func buildErrorMap(err error) map[string]string {
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		errorsMap := make(map[string]string)

		for _, fe := range ve {
			errorsMap[fe.Field()] = fe.Tag()
		}
		return errorsMap
	}

	return nil
}
