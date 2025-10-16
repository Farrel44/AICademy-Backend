package utils

import (
	"errors"
	"strings"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CaptureError(c *fiber.Ctx, err error) {
	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
	}
}

func CaptureMessage(c *fiber.Ctx, message string) {
	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.CaptureMessage(message)
	}
}

func AddBreadcrumb(c *fiber.Ctx, message, category string) {
	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Message:  message,
			Category: category,
			Level:    sentry.LevelInfo,
		}, nil)
	}
}

func SetUserContext(c *fiber.Ctx, userID, email, role string) {
	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:    userID,
				Email: email,
			})
			scope.SetTag("user_role", role)
		})
	}
}

func SetTag(c *fiber.Ctx, key, value string) {
	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.Scope().SetTag(key, value)
	}
}

func SetExtra(c *fiber.Ctx, key string, value interface{}) {
	if hub := sentryfiber.GetHubFromContext(c); hub != nil {
		hub.Scope().SetExtra(key, value)
	}
}

func CaptureDBError(err error, operation string) {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("error_type", "database")
		scope.SetTag("operation", operation)

		if strings.Contains(err.Error(), "connection") {
			scope.SetTag("db_issue", "connection")
		} else if strings.Contains(err.Error(), "timeout") {
			scope.SetTag("db_issue", "timeout")
		} else if strings.Contains(err.Error(), "constraint") {
			scope.SetTag("db_issue", "constraint")
		}

		scope.SetLevel(sentry.LevelError)
		sentry.CaptureException(err)
	})
}

func CaptureSlowQuery(duration string, query string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("performance_issue", "slow_query")
		scope.SetExtra("duration", duration)
		scope.SetExtra("query", query)
		scope.SetLevel(sentry.LevelWarning)
		sentry.CaptureMessage("Slow database query detected")
	})
}
