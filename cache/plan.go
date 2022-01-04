package cache

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PlanResourceChange function
func (s *RawProviderServer) PlanResourceChange(ctx context.Context, req *tfprotov5.PlanResourceChangeRequest) (*tfprotov5.PlanResourceChangeResponse, error) {
	resp := &tfprotov5.PlanResourceChangeResponse{}

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

	// Decode proposed resource state
	proposedState, err := req.ProposedNewState.Unmarshal(rt)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to unmarshal planned resource state",
			Detail:   err.Error(),
		})
		return resp, nil
	}

	proposedVal := make(map[string]tftypes.Value)
	err = proposedState.As(&proposedVal)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to extract planned resource state from tftypes.Value",
			Detail:   err.Error(),
		})
		return resp, nil
	}

	// Decode prior resource state
	priorState, err := req.PriorState.Unmarshal(rt)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to unmarshal prior resource state",
			Detail:   err.Error(),
		})
		return resp, nil
	}
	s.logger.Trace("[PlanResourceChange]", "[PriorState]", dump(priorState))

	priorVal := make(map[string]tftypes.Value)
	err = priorState.As(&priorVal)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to extract prior resource state from tftypes.Value",
			Detail:   err.Error(),
		})
		return resp, nil
	}

	if proposedState.IsNull() {
		// we plan to delete the resource
		if _, ok := priorVal["timestamp"]; ok {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Invalid prior state while planning for destroy",
				Detail:   fmt.Sprintf("'timestamp' attribute missing from state: %s", err),
			})
			return resp, nil
		}
		resp.PlannedState = req.ProposedNewState
		return resp, nil
	}

	if proposedVal["timestamp"].IsNull() {
		// plan for Create
		proposedVal["timestamp"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		propStateVal := tftypes.NewValue(proposedState.Type(), proposedVal)
		s.logger.Trace("[PlanResourceChange]", "new planned state", dump(propStateVal))

		plannedState, err := tfprotov5.NewDynamicValue(rt, propStateVal)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Failed to assemble proposed state during plan",
				Detail:   err.Error(),
			})
			return resp, nil
		}

		resp.PlannedState = &plannedState
	} else {
		// plan for Update
		// NO-OP
		resp.PlannedState = req.PriorState
	}

	return resp, nil
}
