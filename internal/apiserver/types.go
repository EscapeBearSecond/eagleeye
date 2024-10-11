package apiserver

import (
	"github.com/EscapeBearSecond/eagleeye/pkg/types"
)

type Empty struct{}

type CreatePlanRequest types.Options

type CreatePlanReplay struct {
	PlanID string `json:"plan_id"`
}

type RestartPlanRequest struct {
	PlanID string `param:"plan_id" valdiate:"required" message:"plan_id is required"`
}

type RestartPlanReplay struct {
	PlanID string `json:"plan_id"`
}

type StopPlanRequest struct {
	PlanID string `param:"plan_id" valdiate:"required" message:"plan_id is required"`
}

type StopPlanReplay struct {
	PlanID string `json:"plan_id"`
}

type GetPlanResultsRequest struct {
	PlanID string `param:"plan_id" valdiate:"required" message:"plan_id is required"`
}

type GetPlanResultsReplay struct {
	PlanID              string             `json:"plan_id"`
	State               byte               `json:"state"` //0:成功 1:失败
	HostDiscoveryResult *types.PingResult  `json:"host_discovery_result"`
	PlanScanningResult  *types.PortResult  `json:"plan_scanning_result"`
	JobResults          []*types.JobResult `json:"job_results"`
}

type RunningPlansReplay struct {
	PlanIDs []string `json:"plan_ids"`
}

type StoppedPlansReplay struct {
	PlanIDs []string `json:"plan_ids"`
}
