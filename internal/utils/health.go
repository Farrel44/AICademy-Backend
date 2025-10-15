package utils

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  DBHealth  `json:"database"`
	Server    string    `json:"server"`
}

type DBHealth struct {
	Status             string  `json:"status"`
	MaxOpenConnections int     `json:"max_open_connections"`
	OpenConnections    int     `json:"open_connections"`
	InUse              int     `json:"in_use"`
	Idle               int     `json:"idle"`
	WaitCount          int64   `json:"wait_count"`
	WaitDuration       string  `json:"wait_duration"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

// HealthCheck returns health status of the application and database
func HealthCheck(sqlDB *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dbHealth := checkDBHealth(sqlDB)

		status := "healthy"
		if dbHealth.Status != "healthy" {
			status = "unhealthy"
		}

		health := HealthStatus{
			Status:    status,
			Timestamp: time.Now(),
			Database:  dbHealth,
			Server:    "AICademy Backend",
		}

		statusCode := fiber.StatusOK
		if status != "healthy" {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(health)
	}
}

// DetailedDBStats returns detailed database connection pool statistics
func DetailedDBStats(sqlDB *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		stats := sqlDB.Stats()

		utilizationPercent := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100

		detailedStats := map[string]interface{}{
			"timestamp":            time.Now(),
			"max_open_connections": stats.MaxOpenConnections,
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"wait_count":           stats.WaitCount,
			"wait_duration":        stats.WaitDuration.String(),
			"max_idle_closed":      stats.MaxIdleClosed,
			"max_lifetime_closed":  stats.MaxLifetimeClosed,
			"utilization_percent":  utilizationPercent,
			"pool_health": map[string]interface{}{
				"status":          getPoolHealthStatus(utilizationPercent, stats.WaitCount),
				"recommendations": getPoolRecommendations(utilizationPercent, stats.WaitCount),
			},
		}

		return c.JSON(detailedStats)
	}
}

func checkDBHealth(sqlDB *sql.DB) DBHealth {
	stats := sqlDB.Stats()
	utilizationPercent := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100

	// Check if database is responsive
	if err := sqlDB.Ping(); err != nil {
		return DBHealth{
			Status: "unhealthy",
		}
	}

	status := "healthy"
	if utilizationPercent > 90 || stats.WaitCount > 100 {
		status = "degraded"
	}

	return DBHealth{
		Status:             status,
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration.String(),
		UtilizationPercent: utilizationPercent,
	}
}

func getPoolHealthStatus(utilization float64, waitCount int64) string {
	if utilization > 90 || waitCount > 100 {
		return "critical"
	}
	if utilization > 70 || waitCount > 10 {
		return "warning"
	}
	return "healthy"
}

func getPoolRecommendations(utilization float64, waitCount int64) []string {
	var recommendations []string

	if utilization > 80 {
		recommendations = append(recommendations, "Consider increasing MaxOpenConns")
	}
	if waitCount > 0 {
		recommendations = append(recommendations, "Connection waits detected - review query performance")
	}
	if utilization < 30 {
		recommendations = append(recommendations, "Pool utilization is low - current config is adequate")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Pool performance is optimal")
	}

	return recommendations
}
