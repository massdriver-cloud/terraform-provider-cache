package cache

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ValidateResourceTypeConfig function
func (s *RawProviderServer) ValidateResourceTypeConfig(ctx context.Context, req *tfprotov5.ValidateResourceTypeConfigRequest) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	resp := &tfprotov5.ValidateResourceTypeConfigResponse{}
	// requiredKeys := []string{"apiVersion", "kind", "metadata"}
	// forbiddenKeys := []string{"status"}

	rt, err := GetResourceType(req.TypeName)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to determine resource type",
			Detail:   err.Error(),
		})
		return resp, nil
	}
	log.Println("--------------------TEST----------------------")
	log.Printf("ResourceType: %v\n", rt)

	// Decode proposed resource state
	config, err := req.Config.Unmarshal(rt)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to unmarshal resource state",
			Detail:   err.Error(),
		})
		return resp, nil
	}

	att := tftypes.NewAttributePath()
	att = att.WithAttributeName("manifest")

	configVal := make(map[string]tftypes.Value)
	err = config.As(&configVal)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to extract resource state from SDK value",
			Detail:   err.Error(),
		})
		return resp, nil
	}

	_, ok := configVal["value"]
	if !ok {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
			Severity:  tfprotov5.DiagnosticSeverityError,
			Summary:   "Value missing from resource configuration",
			Detail:    "A value attribute containing a valid terraform value is required.",
			Attribute: att,
		})
		return resp, nil
	}

	// rawManifest := make(map[string]tftypes.Value)
	// err = manifest.As(&rawManifest)
	// if err != nil {
	// 	if err.Error() == "unmarshaling unknown values is not supported" {
	// 		// Likely this validation call came too early and the manifest still contains unknown values.
	// 		// Bailing out without error to allow the resource to be completed at a later stage.
	// 		return resp, nil
	// 	}
	// 	resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
	// 		Severity:  tfprotov5.DiagnosticSeverityError,
	// 		Summary:   `Failed to extract "manifest" attribute value from resource configuration`,
	// 		Detail:    err.Error(),
	// 		Attribute: att,
	// 	})
	// 	return resp, nil
	// }

	// for _, key := range requiredKeys {
	// 	if _, present := rawManifest[key]; !present {
	// 		kp := att.WithAttributeName(key)
	// 		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
	// 			Severity:  tfprotov5.DiagnosticSeverityError,
	// 			Summary:   `Attribute key missing from "manifest" value`,
	// 			Detail:    fmt.Sprintf("'%s' attribute key is missing from manifest configuration", key),
	// 			Attribute: kp,
	// 		})
	// 	}
	// }

	// for _, key := range forbiddenKeys {
	// 	if _, present := rawManifest[key]; present {
	// 		kp := att.WithAttributeName(key)
	// 		resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
	// 			Severity:  tfprotov5.DiagnosticSeverityError,
	// 			Summary:   `Forbidden attribute key in "manifest" value`,
	// 			Detail:    fmt.Sprintf("'%s' attribute key is not allowed in manifest configuration", key),
	// 			Attribute: kp,
	// 		})
	// 	}
	// }

	return resp, nil
}
