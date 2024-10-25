package apiserver

import (
	_ "github.com/EscapeBearSecond/falcon/docs"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func RegisterRoutes(e *echo.Echo, release ...bool) {
	var releaseMode bool
	if len(release) > 0 && release[0] {
		releaseMode = release[0]
	}

	v1Group := e.Group("/api/v1")

	if !releaseMode {
		v1Group.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	planService := &PlanService{}

	v1Group.POST("/plan", Handle(planService.Create))
	v1Group.PUT("/plan/:plan_id", Handle(planService.Restart))
	v1Group.DELETE("/plan/:plan_id", Handle(planService.Stop))
	v1Group.GET("/plan/:plan_id/results", Handle(planService.GetResults))
	v1Group.GET("/plan/running", Handle(planService.RunningPlans))
	v1Group.GET("/plan/stopped", Handle(planService.StoppedPlans))
}
