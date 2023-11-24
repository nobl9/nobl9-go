package agent

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var agentValidation = validation.New[Agent](
	v1alpha.FieldRuleMetadataName(func(a Agent) string { return a.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(a Agent) string { return a.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(a Agent) string { return a.Metadata.Project }),
	v1alpha.FieldRuleSpecDescription(func(a Agent) string { return a.Spec.Description }),
	validation.For(func(a Agent) Spec { return a.Spec }).
		Include(specValidation),
)

var specValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(exactlyOneDataSourceTypeValidationRule),
	validation.For(func(s Spec) v1alpha.ReleaseChannel { return s.ReleaseChannel }).
		WithName("releaseChannel").
		Omitempty().
		Rules(v1alpha.ReleaseChannelValidation()),
)

const errCodeExactlyOneDataSourceType = "exactly_one_data_source_type"

var exactlyOneDataSourceTypeValidationRule = validation.NewSingleRule(func(spec Spec) error {
	var onlyType v1alpha.DataSourceType
	typesMatch := func(typ v1alpha.DataSourceType) error {
		if onlyType == 0 {
			onlyType = typ
		}
		if onlyType != typ {
			return errors.Errorf(
				"must have exactly one datas source type, detected both %s and %s",
				onlyType, typ)
		}
		return nil
	}
	if spec.Prometheus != nil {
		if err := typesMatch(v1alpha.Prometheus); err != nil {
			return err
		}
	}
	if spec.Datadog != nil {
		if err := typesMatch(v1alpha.Datadog); err != nil {
			return err
		}
	}
	if spec.NewRelic != nil {
		if err := typesMatch(v1alpha.NewRelic); err != nil {
			return err
		}
	}
	if spec.AppDynamics != nil {
		if err := typesMatch(v1alpha.AppDynamics); err != nil {
			return err
		}
	}
	if spec.Splunk != nil {
		if err := typesMatch(v1alpha.Splunk); err != nil {
			return err
		}
	}
	if spec.Lightstep != nil {
		if err := typesMatch(v1alpha.Lightstep); err != nil {
			return err
		}
	}
	if spec.SplunkObservability != nil {
		if err := typesMatch(v1alpha.SplunkObservability); err != nil {
			return err
		}
	}
	if spec.ThousandEyes != nil {
		if err := typesMatch(v1alpha.ThousandEyes); err != nil {
			return err
		}
	}
	if spec.Dynatrace != nil {
		if err := typesMatch(v1alpha.Dynatrace); err != nil {
			return err
		}
	}
	if spec.Elasticsearch != nil {
		if err := typesMatch(v1alpha.Elasticsearch); err != nil {
			return err
		}
	}
	if spec.Graphite != nil {
		if err := typesMatch(v1alpha.Graphite); err != nil {
			return err
		}
	}
	if spec.BigQuery != nil {
		if err := typesMatch(v1alpha.BigQuery); err != nil {
			return err
		}
	}
	if spec.OpenTSDB != nil {
		if err := typesMatch(v1alpha.OpenTSDB); err != nil {
			return err
		}
	}
	if spec.GrafanaLoki != nil {
		if err := typesMatch(v1alpha.GrafanaLoki); err != nil {
			return err
		}
	}
	if spec.CloudWatch != nil {
		if err := typesMatch(v1alpha.CloudWatch); err != nil {
			return err
		}
	}
	if spec.Pingdom != nil {
		if err := typesMatch(v1alpha.Pingdom); err != nil {
			return err
		}
	}
	if spec.AmazonPrometheus != nil {
		if err := typesMatch(v1alpha.AmazonPrometheus); err != nil {
			return err
		}
	}
	if spec.Redshift != nil {
		if err := typesMatch(v1alpha.Redshift); err != nil {
			return err
		}
	}
	if spec.SumoLogic != nil {
		if err := typesMatch(v1alpha.SumoLogic); err != nil {
			return err
		}
	}
	if spec.Instana != nil {
		if err := typesMatch(v1alpha.Instana); err != nil {
			return err
		}
	}
	if spec.InfluxDB != nil {
		if err := typesMatch(v1alpha.InfluxDB); err != nil {
			return err
		}
	}
	if spec.GCM != nil {
		if err := typesMatch(v1alpha.GCM); err != nil {
			return err
		}
	}
	if spec.AzureMonitor != nil {
		if err := typesMatch(v1alpha.AzureMonitor); err != nil {
			return err
		}
	}
	if spec.Generic != nil {
		if err := typesMatch(v1alpha.Generic); err != nil {
			return err
		}
	}
	if spec.Honeycomb != nil {
		if err := typesMatch(v1alpha.Honeycomb); err != nil {
			return err
		}
	}
	if onlyType == 0 {
		return errors.New("must have exactly one data source type, none were provided")
	}
	return nil
}).WithErrorCode(errCodeExactlyOneDataSourceType)

func validate(a Agent) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(agentValidation, a)
}
