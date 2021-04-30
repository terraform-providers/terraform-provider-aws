package aws

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/service/elasticache"
	multierror "github.com/hashicorp/go-multierror"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfelasticache "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache"
)

const (
	redisVersionPreV6RegexpRaw  = `[1-5](\.[[:digit:]]+){2}`
	redisVersionPostV6RegexpRaw = `([6-9]|[[:digit:]]{2})\.x`

	redisVersionRegexpRaw = redisVersionPreV6RegexpRaw + "|" + redisVersionPostV6RegexpRaw
)

const (
	redisVersionRegexpPattern       = "^" + redisVersionRegexpRaw + "$"
	redisVersionPostV6RegexpPattern = "^" + redisVersionPostV6RegexpRaw + "$"
)

var (
	redisVersionRegexp       = regexp.MustCompile(redisVersionRegexpPattern)
	redisVersionPostV6Regexp = regexp.MustCompile(redisVersionPostV6RegexpPattern)
)

func ValidateElastiCacheRedisVersionString(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !redisVersionRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: Redis versions must match <major>.x when using version 6 or higher, or <major>.<minor>.<bug-fix>", k))
	}

	return
}

// NormalizeElastiCacheEngineVersion returns a github.com/hashicorp/go-version Version
// that can handle a regular 1.2.3 version number or a 6.x version number used for
// ElastiCache Redis version 6 and higher
func NormalizeElastiCacheEngineVersion(version string) (*gversion.Version, error) {
	if matches := redisVersionPostV6Regexp.FindStringSubmatch(version); matches != nil {
		version = matches[1]
	}
	return gversion.NewVersion(version)
}

// CustomizeDiffElastiCacheEngineVersion causes re-creation of the resource if the version is being downgraded
func CustomizeDiffElastiCacheEngineVersion(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" || !diff.HasChange("engine_version") {
		return nil
	}

	o, n := diff.GetChange("engine_version")
	oVersion, err := NormalizeElastiCacheEngineVersion(o.(string))
	if err != nil {
		return fmt.Errorf("error parsing old engine_version: %w", err)
	}
	nVersion, err := NormalizeElastiCacheEngineVersion(n.(string))
	if err != nil {
		return fmt.Errorf("error parsing new engine_version: %w", err)
	}

	if nVersion.GreaterThan(oVersion) {
		return nil
	}

	return diff.ForceNew("engine_version")
}

// CustomizeDiffValidateClusterAZMode validates that `num_cache_nodes` is greater than 1 when `az_mode` is "cross-az"
func CustomizeDiffValidateClusterAZMode(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("az_mode"); !ok || v.(string) != elasticache.AZModeCrossAz {
		return nil
	}
	if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) != 1 {
		return nil
	}

	return errors.New(`az_mode "cross-az" is not supported with num_cache_nodes = 1`)
}

// CustomizeDiffValidateClusterEngineVersion validates the correct format for `engine_version`, based on `engine`
func CustomizeDiffValidateClusterEngineVersion(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// Memcached: Versions in format <major>.<minor>.<bug fix>
	// Redis: Starting with version 6, must match <major>.x, prior to version 6, <major>.<minor>.<bug fix>
	engineVersion, ok := diff.GetOk("engine_version")
	if !ok {
		return nil
	}

	var validator schema.SchemaValidateFunc
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == tfelasticache.EngineMemcached {
		validator = validateVersionString
	} else {
		validator = ValidateElastiCacheRedisVersionString
	}

	_, errs := validator(engineVersion, "engine_version")

	var err *multierror.Error
	err = multierror.Append(err, errs...)
	return err.ErrorOrNil()
}

// CustomizeDiffValidateClusterNumCacheNodes validates that `num_cache_nodes` is 1 when `engine` is "redis"
func CustomizeDiffValidateClusterNumCacheNodes(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == tfelasticache.EngineMemcached {
		return nil
	}

	if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) == 1 {
		return nil
	}
	return errors.New(`engine "redis" does not support num_cache_nodes > 1`)
}

// CustomizeDiffClusterMemcachedNodeType causes re-creation when `node_type` is changed and `engine` is "memcached"
func CustomizeDiffClusterMemcachedNodeType(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// Engine memcached does not currently support vertical scaling
	// https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/Scaling.html#Scaling.Memcached.Vertically
	if diff.Id() == "" || !diff.HasChange("node_type") {
		return nil
	}
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == tfelasticache.EngineRedis {
		return nil
	}
	return diff.ForceNew("node_type")
}

// CustomizeDiffValidateClusterMemcachedSnapshotIdentifier validates that `final_snapshot_identifier` is not set when `engine` is "memcached"
func CustomizeDiffValidateClusterMemcachedSnapshotIdentifier(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == tfelasticache.EngineRedis {
		return nil
	}
	if _, ok := diff.GetOk("final_snapshot_identifier"); !ok {
		return nil
	}
	return errors.New(`engine "memcached" does not support final_snapshot_identifier`)
}

// CustomizeDiffValidateReplicationGroupAutomaticFailover validates that `automatic_failover_enabled` is set when `multi_az_enabled` is true
func CustomizeDiffValidateReplicationGroupAutomaticFailover(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v := diff.Get("multi_az_enabled").(bool); !v {
		return nil
	}
	if v := diff.Get("automatic_failover_enabled").(bool); !v {
		return errors.New(`automatic_failover_enabled must be true if multi_az_enabled is true`)
	}
	return nil
}
