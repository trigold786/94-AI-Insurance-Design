package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
)

func main() {
	cfg := config.Load()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		userID := c.GetHeader("x-user-id")
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})

	api := r.Group("/v1")
	{
		api.GET("/profile", handleGetProfile)
		api.PUT("/profile", handleUpdateProfile)
		api.GET("/policies", handleQueryPolicies)
		api.POST("/plans/generate", handleGeneratePlan)
		api.GET("/plans/:id", handleGetPlan)
		api.GET("/compliance/:policy_id", handleGetCompliance)
	}

	log.Printf("api-server starting on :%d", cfg.ServerPort)
	r.Run(fmt.Sprintf(":%d", cfg.ServerPort))
}

// Handler stubs — implemented in subsequent sprints
func handleGetProfile(c *gin.Context)       { c.JSON(200, gin.H{"code": 0, "data": gin.H{}}) }
func handleUpdateProfile(c *gin.Context)    { c.JSON(200, gin.H{"code": 0, "data": gin.H{}}) }
func handleQueryPolicies(c *gin.Context)    { c.JSON(200, gin.H{"code": 0, "data": []gin.H{}}) }
func handleGeneratePlan(c *gin.Context)     { c.JSON(200, gin.H{"code": 0, "data": gin.H{}}) }
func handleGetPlan(c *gin.Context)          { c.JSON(200, gin.H{"code": 0, "data": gin.H{}}) }
func handleGetCompliance(c *gin.Context)    { c.JSON(200, gin.H{"code": 0, "data": gin.H{}}) }
