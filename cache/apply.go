package cache

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ApplyResourceChange function
func (s *RawProviderServer) ApplyResourceChange(ctx context.Context, req *tfprotov5.ApplyResourceChangeRequest) (*tfprotov5.ApplyResourceChangeResponse, error) {
	resp := &tfprotov5.ApplyResourceChangeResponse{}

	execDiag := s.canExecute()
	if len(execDiag) > 0 {
		resp.Diagnostics = append(resp.Diagnostics, execDiag...)
		return resp, nil
	}

	rt, err := GetResourceType(req.TypeName)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to determine planned resource type",
			Detail:   err.Error(),
		})
		return resp, nil
	}

	applyPlannedState, err := req.PlannedState.Unmarshal(rt)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to unmarshal planned resource state",
			Detail:   err.Error(),
		})
		return resp, nil
	}
	s.logger.Trace("[ApplyResourceChange][PlannedState] %#v", applyPlannedState)

	applyPriorState, err := req.PriorState.Unmarshal(rt)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to unmarshal prior resource state",
			Detail:   err.Error(),
		})
		return resp, nil
	}
	s.logger.Trace("[ApplyResourceChange]", "[PriorState]", dump(applyPriorState))

	s.logger.Trace("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	s.logger.Trace("[ApplyResourceChange]", "[PriorState]", dump(applyPriorState))
	s.logger.Trace("[ApplyResourceChange]", "[req.PlannedState]", dump(req.PlannedState))
	s.logger.Trace("[ApplyResourceChange]", "[PlannedState]", dump(applyPlannedState))

	newVal := make(map[string]tftypes.Value)
	applyPlannedState.As(&newVal)

	newVal["timestamp"] = tftypes.NewValue(tftypes.String, "testing")

	newStateVal := tftypes.NewValue(applyPlannedState.Type(), newVal)

	s.logger.Trace("[ApplyResourceChange]", "[PropStateVal]", dump(newStateVal))

	plannedState, err := tfprotov5.NewDynamicValue(rt, newStateVal)
	//plannedState, err := tfprotov5.NewDynamicValue(applyPlannedState.Type(), newStateVal)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to assemble proposed state during apply",
			Detail:   err.Error(),
		})
		return resp, nil
	}
	s.logger.Trace("[ApplyResourceChange]", "[PlannedState]", dump(plannedState))

	resp.NewState = &plannedState
	//resp.NewState = req.PlannedState
	return resp, nil
}
