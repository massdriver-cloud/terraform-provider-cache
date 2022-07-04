package cache

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RawProviderServer implements the ProviderServer interface as exported from ProtoBuf.
type RawProviderServer struct {
	// Since the provider is essentially a gRPC server, the execution flow is dictated by the order of the client (Terraform) request calls.
	// Thus it needs a way to persist state between the gRPC calls. These attributes store values that need to be persisted between gRPC calls,
	// such as instances of the Kubernetes clients, configuration options needed at runtime.
	logger hclog.Logger

	//providerEnabled bool
	hostTFVersion string
}

func dump(v interface{}) hclog.Format {
	return hclog.Fmt("%v", v)
}

// PrepareProviderConfig function
func (s *RawProviderServer) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	s.logger.Trace("[PrepareProviderConfig][Request]\n%s\n", dump(*req))
	resp := &tfprotov5.PrepareProviderConfigResponse{PreparedConfig: req.Config}
	return resp, nil
}

// ValidateDataSourceConfig function
func (s *RawProviderServer) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	s.logger.Trace("[ValidateDataSourceConfig][Request]\n%s\n", dump(*req))
	resp := &tfprotov5.ValidateDataSourceConfigResponse{}
	return resp, nil
}

func (s *RawProviderServer) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	resp := &tfprotov5.UpgradeResourceStateResponse{}
	resp.Diagnostics = []*tfprotov5.Diagnostic{}

	var resourceValue tftypes.Value
	var err error

	stateSchemaVersion := req.Version
	currentSchemaVersion := GetProviderResourceSchemas()[req.TypeName].Version
	resourceType := GetObjectTypeFromSchema(GetProviderResourceSchemas()[req.TypeName])

	if stateSchemaVersion != currentSchemaVersion {
		stateSchema := GetProviderResourceSchemasByVersion(req.Version)[req.TypeName]
		stateType := GetObjectTypeFromSchema(stateSchema)
		resourceValue, err = req.RawState.Unmarshal(stateType)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Failed to decode old state during upgrade",
				Detail:   err.Error(),
			})
			return resp, nil
		}

		upgradeFunc := GetProviderResourceUpgradeFunctions()[req.TypeName]
		resourceValue, err = upgradeFunc(resourceValue, stateSchemaVersion)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Failed to upgrade the schema of the resource",
				Detail:   err.Error(),
			})
			return resp, nil
		}
	} else {
		resourceValue, err = req.RawState.Unmarshal(resourceType)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Failed to decode old state during upgrade",
				Detail:   err.Error(),
			})
			return resp, nil
		}
	}

	us, err := tfprotov5.NewDynamicValue(resourceType, resourceValue)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to encode new state during upgrade",
			Detail:   err.Error(),
		})
	}
	resp.UpgradedState = &us

	return resp, nil
}

// ReadDataSource function
func (s *RawProviderServer) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	s.logger.Trace("[ReadDataSource][Request]\n%s\n", dump(*req))

	return nil, status.Errorf(codes.Unimplemented, "method ReadDataSource not implemented")
}

// StopProvider function
func (s *RawProviderServer) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	s.logger.Trace("[StopProvider][Request]\n%s\n", dump(*req))

	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
