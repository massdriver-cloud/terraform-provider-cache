package cache

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// GetObjectTypeFromSchema returns a tftypes.Type that can wholy represent the schema input
func GetObjectTypeFromSchema(schema *tfprotov5.Schema) tftypes.Type {
	bm := map[string]tftypes.Type{}

	for _, att := range schema.Block.Attributes {
		bm[att.Name] = att.Type
	}

	for _, b := range schema.Block.BlockTypes {
		attrs := map[string]tftypes.Type{}
		for _, att := range b.Block.Attributes {
			attrs[att.Name] = att.Type
		}
		bm[b.TypeName] = tftypes.List{
			ElementType: tftypes.Object{AttributeTypes: attrs},
		}
		// TODO handle repeated blocks
	}

	return tftypes.Object{AttributeTypes: bm}
}

// GetResourceType returns the tftypes.Type of a resource of type 'name'
func GetResourceType(name string) (tftypes.Type, error) {
	sch := GetProviderResourceSchemas()
	rsch, ok := sch[name]
	if !ok {
		return tftypes.DynamicPseudoType, fmt.Errorf("unknown resource %s - cannot find schema", name)
	}
	return GetObjectTypeFromSchema(rsch), nil
}

// GetProviderResourceSchema contains the definitions of all supported resources
func GetProviderResourceSchemas() map[string]*tfprotov5.Schema {
	return map[string]*tfprotov5.Schema{
		"cache_store": GetLatestCacheStoreSchema(),
	}
}

// GetProviderResourceSchema contains the definitions of all supported resources
func GetProviderResourceSchemasByVersion(version int64) map[string]*tfprotov5.Schema {
	return map[string]*tfprotov5.Schema{
		"cache_store": GetCacheStoreSchemaByVersion(version),
	}
}

// GetProviderResourceSchema contains the definitions of all supported resources
func GetProviderResourceUpgradeFunctions() map[string]func(resourceValue tftypes.Value, fromVersion int64) (tftypes.Value, error) {
	return map[string]func(resourceValue tftypes.Value, fromVersion int64) (tftypes.Value, error){
		"cache_store": CacheStoreUpgradeResource,
	}
}
