package apiserver

import (
	"context"

	eagleeye "github.com/EscapeBearSecond/falcon/pkg/sdk"
	"github.com/EscapeBearSecond/falcon/pkg/types"
)

type PlanService struct{}

// @Summary 创建计划
// @Description 创建计划
// @Tags plans
// @Accept json
// @Produce json
// @Param plan body CreatePlanRequest true "计划"
// @Success 200 {object} CreatePlanReplay
// @Failure 400 {object} status
// @Router /plan [post]
func (s *PlanService) Create(ctx context.Context, request *CreatePlanRequest) (*CreatePlanReplay, error) {
	results := &GetPlanResultsReplay{
		JobResults: []*types.JobResult{},
	}

	s.setCallback(request, results)

	entry, err := Eagleeye.NewEntry((*types.Options)(request))
	if err != nil {
		return nil, WithCaller(err)
	}

	results.PlanID = entry.EntryID

	err = DB.StorePlan(entry.EntryID, request)
	if err != nil {
		return nil, WithCaller(err)
	}

	go s.runAsync(entry, results)

	return &CreatePlanReplay{
		PlanID: entry.EntryID,
	}, nil
}

// @Summary 重启计划
// @Description 重启计划
// @Tags plans
// @Accept json
// @Produce json
// @Param plan_id path string true "计划ID"
// @Success 200 {object} RestartPlanReplay
// @Failure 400 {object} status
// @Failure 404 {object} status
// @Router /plan/{plan_id} [post]
func (s *PlanService) Restart(ctx context.Context, request *RestartPlanRequest) (*RestartPlanReplay, error) {
	newEntry := Eagleeye.Entry(request.PlanID)
	if newEntry != nil {
		newEntry.Stop()
	}

	plan, err := DB.GetPlan(request.PlanID)
	if err != nil {
		return nil, WithCaller(err)
	}

	results := &GetPlanResultsReplay{
		JobResults: []*types.JobResult{},
	}

	s.setCallback(plan, results)

	newEntry, err = Eagleeye.NewEntry((*types.Options)(plan))
	if err != nil {
		return nil, WithCaller(err)
	}

	err = DB.RestorePlan(request.PlanID, newEntry.EntryID, plan)
	if err != nil {
		return nil, WithCaller(err)
	}

	go s.runAsync(newEntry, results)

	return &RestartPlanReplay{PlanID: newEntry.EntryID}, nil
}

// @Summary 停止计划
// @Description 停止计划
// @Tags plans
// @Accept json
// @Produce json
// @Param plan_id path string true "计划ID"
// @Success 200 {object} StopPlanReplay
// @Failure 400 {object} status
// @Failure 404 {object} status
// @Router /plan/{plan_id} [delete]
func (s *PlanService) Stop(ctx context.Context, request *StopPlanRequest) (*StopPlanReplay, error) {
	newEntry := Eagleeye.Entry(request.PlanID)
	if newEntry != nil {
		newEntry.Stop()
	}

	_, err := DB.GetPlan(request.PlanID)
	if err != nil {
		return nil, WithCaller(err)
	}

	err = DB.DeletePlan(request.PlanID)
	if err != nil {
		return nil, WithCaller(err)
	}

	return &StopPlanReplay{PlanID: request.PlanID}, nil
}

// @Summary 获取计划结果
// @Description 获取计划结果
// @Tags plans
// @Accept json
// @Produce json
// @Param plan_id path string true "计划ID"
// @Success 200 {object} GetPlanResultsReplay
// @Failure 400 {object} status
// @Failure 404 {object} status
// @Router /plan/{plan_id}/results [get]
func (s *PlanService) GetResults(ctx context.Context, request *GetPlanResultsRequest) (*GetPlanResultsReplay, error) {
	results, err := DB.GetResults(request.PlanID)
	if err != nil {
		return nil, WithCaller(err)
	}
	return results, nil
}

// @Summary 获取运行中的计划
// @Description 获取运行中的计划
// @Tags plans
// @Accept json
// @Produce json
// @Success 200 {object} RunningPlansReplay
// @Router /plan/running [get]
func (s *PlanService) RunningPlans(ctx context.Context, _ *Empty) (*RunningPlansReplay, error) {
	plans, err := DB.RunningPlans()
	if err != nil {
		return nil, WithCaller(err)
	}
	return &RunningPlansReplay{PlanIDs: plans}, nil
}

// @Summary 获取已停止的计划
// @Description 获取已停止的计划
// @Tags plans
// @Accept json
// @Produce json
// @Success 200 {object} StoppedPlansReplay
// @Router /plan/stopped [get]
func (s *PlanService) StoppedPlans(ctx context.Context, _ *Empty) (*StoppedPlansReplay, error) {
	plans, err := DB.StoppedPlans()
	if err != nil {
		return nil, WithCaller(err)
	}
	return &StoppedPlansReplay{PlanIDs: plans}, nil
}

func (s *PlanService) runAsync(entry *eagleeye.EagleeyeEntry, results *GetPlanResultsReplay) {
	err := entry.Run(context.Background())
	if err != nil {
		Logger.Error("Entry.Run failed", "plan_id", entry.EntryID, "error", err)
		results.State = 1
	}

	err = DB.StoreResults(entry.EntryID, results)
	if err != nil {
		Logger.Error("DB.AddResults failed", "plan_id", entry.EntryID, "error", err)
		return
	}

	Logger.Info("Entry.Run success", "plan_id", entry.EntryID)
}

func (s *PlanService) setCallback(plan *CreatePlanRequest, results *GetPlanResultsReplay) {
	if plan.HostDiscovery.Use {
		plan.HostDiscovery.ResultCallback = func(ctx context.Context, pr *types.PingResult) error {
			results.HostDiscoveryResult = pr
			return nil
		}
	}
	if plan.PortScanning.Use {
		plan.PortScanning.ResultCallback = func(ctx context.Context, pr *types.PortResult) error {
			results.PlanScanningResult = pr
			return nil
		}
	}
	for i := range plan.Jobs {
		plan.Jobs[i].ResultCallback = func(ctx context.Context, jr *types.JobResult) error {
			results.JobResults = append(results.JobResults, jr)
			return nil
		}
	}
}
