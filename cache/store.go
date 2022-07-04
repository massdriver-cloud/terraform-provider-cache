package cache

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var CACHE_STORE_V2 = tfprotov5.Schema{
	Version: 2,
	Block: &tfprotov5.SchemaBlock{
		BlockTypes: []*tfprotov5.SchemaNestedBlock{},
		Attributes: []*tfprotov5.SchemaAttribute{
			{
				Name:        "timestamp",
				Type:        tftypes.String,
				Required:    false,
				Computed:    true,
				Optional:    false,
				Description: "The timestamp this cached value was created",
			},
			{
				Name:        "value",
				Type:        tftypes.DynamicPseudoType,
				Required:    true,
				Optional:    false,
				Computed:    false,
				Description: "The value to cache.",
			},
			{
				Name:        "keepers",
				Type:        tftypes.Map{ElementType: tftypes.DynamicPseudoType},
				Required:    false,
				Optional:    true,
				Computed:    false,
				Description: "Arbitrary map of values that, when changed, will trigger recreation of resource.",
			},
		},
	},
}

var CACHE_STORE_V1 = tfprotov5.Schema{
	Version: 2,
	Block: &tfprotov5.SchemaBlock{
		BlockTypes: []*tfprotov5.SchemaNestedBlock{},
		Attributes: []*tfprotov5.SchemaAttribute{
			{
				Name:        "timestamp",
				Type:        tftypes.String,
				Required:    false,
				Computed:    true,
				Optional:    false,
				Description: "The timestamp this cached value was created",
			},
			{
				Name:        "value",
				Type:        tftypes.DynamicPseudoType,
				Required:    true,
				Optional:    false,
				Computed:    false,
				Description: "The value to cache.",
			},
		},
	},
}

func GetLatestCacheStoreSchema() *tfprotov5.Schema {
	return &CACHE_STORE_V2
}

func GetCacheStoreSchemaByVersion(version int64) *tfprotov5.Schema {
	switch version {
	case 2:
		return &CACHE_STORE_V2
	case 1:
		return &CACHE_STORE_V1
	default:
		return nil
	}
}

func CacheStoreUpgradeResource(priorValue tftypes.Value, fromVersion int64) (tftypes.Value, error) {
	if fromVersion < 2 {
		return CacheStoreUpgradeResourceV1ToV2(priorValue)
	}
	return priorValue, nil
}

func CacheStoreUpgradeResourceV1ToV2(priorValue tftypes.Value) (tftypes.Value, error) {
	updatedValue := make(map[string]tftypes.Value)
	err := priorValue.As(&updatedValue)
	if err != nil {
		return priorValue, err
	}
	updatedValue["keepers"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.DynamicPseudoType}, nil)

	newValue := tftypes.NewValue(GetObjectTypeFromSchema(&CACHE_STORE_V2), updatedValue)

	return newValue, nil
}
