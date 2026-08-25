package fixtures

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gomplate/v3"
)

const (
	metricMean   = "mean"
	metricMedian = "median"
	metricMin    = "min"
	metricMax    = "max"
	metricP95    = "p95"

	directionLower  = "lower"
	directionHigher = "higher"
	directionNone   = "none"
)

// MetricThreshold applies absolute limits and an optional relative baseline limit.
type MetricThreshold struct {
	Min               *float64 `yaml:"min,omitempty" json:"min,omitempty"`
	Max               *float64 `yaml:"max,omitempty" json:"max,omitempty"`
	RegressionPercent *float64 `yaml:"regressionPercent,omitempty" json:"regressionPercent,omitempty"`
}

// MetricSpec describes a numeric CEL extraction and its logical-row policy.
type MetricSpec struct {
	Name      string           `yaml:"name" json:"name"`
	Extract   string           `yaml:"extract" json:"extract"`
	Unit      string           `yaml:"unit" json:"unit"`
	Aggregate string           `yaml:"aggregate,omitempty" json:"aggregate,omitempty"`
	Direction string           `yaml:"direction,omitempty" json:"direction,omitempty"`
	Baseline  string           `yaml:"baseline,omitempty" json:"baseline,omitempty"`
	Threshold *MetricThreshold `yaml:"threshold,omitempty" json:"threshold,omitempty"`
}

func (f FixtureTest) repeatCount() int {
	if f.Repeat != nil {
		return *f.Repeat
	}
	if f.FrontMatter.Repeat != nil {
		return *f.FrontMatter.Repeat
	}
	return 1
}

func (f FixtureTest) metricSpecs() []MetricSpec {
	if len(f.Metrics) > 0 {
		return f.Metrics
	}
	return f.FrontMatter.Metrics
}

func (f FixtureTest) hasSampleConfiguration() bool {
	return f.Repeat != nil || f.FrontMatter.Repeat != nil || len(f.metricSpecs()) > 0
}

func (m MetricSpec) normalizedAggregate() string {
	if m.Aggregate == "" {
		return metricMean
	}
	return strings.ToLower(m.Aggregate)
}

func (m MetricSpec) normalizedDirection() string {
	if m.Direction == "" {
		return directionNone
	}
	return strings.ToLower(m.Direction)
}

// validateFixtureConfiguration rejects metric policy errors before any command runs.
func validateFixtureConfiguration(fixtures []FixtureTest) error {
	rowsByName := make(map[string][]FixtureTest, len(fixtures))
	for _, fixture := range fixtures {
		rowsByName[fixture.Name] = append(rowsByName[fixture.Name], fixture)
		if fixture.Repeat != nil && *fixture.Repeat < 1 {
			return fmt.Errorf("fixture %q: repeat must be at least 1", fixture.Name)
		}
		if fixture.FrontMatter.Repeat != nil && *fixture.FrontMatter.Repeat < 1 {
			return fmt.Errorf("fixture %q: frontmatter repeat must be at least 1", fixture.Name)
		}
		if fixture.Expected.Timeout != nil && *fixture.Expected.Timeout <= 0 {
			return fmt.Errorf("fixture %q: timeout must be greater than zero", fixture.Name)
		}
		if fixture.Timeout != nil && *fixture.Timeout <= 0 {
			return fmt.Errorf("fixture %q: frontmatter timeout must be greater than zero", fixture.Name)
		}
		if err := validateMetricSpecs(fixture); err != nil {
			return err
		}
	}

	for _, fixture := range fixtures {
		for _, metric := range fixture.metricSpecs() {
			if metric.Baseline == "" {
				continue
			}
			baselines := rowsByName[metric.Baseline]
			switch len(baselines) {
			case 0:
				return fmt.Errorf("fixture %q metric %q: baseline row %q was not found", fixture.Name, metric.Name, metric.Baseline)
			case 1:
			default:
				return fmt.Errorf("fixture %q metric %q: baseline row name %q is ambiguous (%d rows)", fixture.Name, metric.Name, metric.Baseline, len(baselines))
			}

			baselineMetric, ok := metricSpecByName(baselines[0].metricSpecs(), metric.Name)
			if !ok {
				return fmt.Errorf("fixture %q metric %q: baseline row %q does not configure that metric", fixture.Name, metric.Name, metric.Baseline)
			}
			if metric.Unit != baselineMetric.Unit || metric.normalizedAggregate() != baselineMetric.normalizedAggregate() || metric.normalizedDirection() != baselineMetric.normalizedDirection() {
				return fmt.Errorf("fixture %q metric %q: baseline row %q must use matching unit, aggregate, and direction", fixture.Name, metric.Name, metric.Baseline)
			}
		}
	}
	return nil
}

func validateMetricSpecs(fixture FixtureTest) error {
	seen := make(map[string]struct{})
	for _, metric := range fixture.metricSpecs() {
		prefix := fmt.Sprintf("fixture %q metric %q", fixture.Name, metric.Name)
		if metric.Name == "" {
			return fmt.Errorf("fixture %q: metric name is required", fixture.Name)
		}
		if _, ok := seen[metric.Name]; ok {
			return fmt.Errorf("%s: duplicate metric name", prefix)
		}
		seen[metric.Name] = struct{}{}
		if metric.Extract == "" {
			return fmt.Errorf("%s: extract is required", prefix)
		}
		if metric.Unit == "" {
			return fmt.Errorf("%s: unit is required", prefix)
		}
		switch metric.normalizedAggregate() {
		case metricMean, metricMedian, metricMin, metricMax, metricP95:
		default:
			return fmt.Errorf("%s: unsupported aggregate %q (want mean, median, min, max, or p95)", prefix, metric.Aggregate)
		}
		switch metric.normalizedDirection() {
		case directionLower, directionHigher, directionNone:
		default:
			return fmt.Errorf("%s: unsupported direction %q (want lower, higher, or none)", prefix, metric.Direction)
		}
		if metric.Baseline != "" && metric.normalizedDirection() == directionNone {
			return fmt.Errorf("%s: baseline comparison requires direction lower or higher", prefix)
		}
		if metric.Threshold == nil {
			continue
		}
		threshold := metric.Threshold
		if threshold.Min != nil && !finite(*threshold.Min) {
			return fmt.Errorf("%s: threshold.min must be finite", prefix)
		}
		if threshold.Max != nil && !finite(*threshold.Max) {
			return fmt.Errorf("%s: threshold.max must be finite", prefix)
		}
		if threshold.Min != nil && threshold.Max != nil && *threshold.Min > *threshold.Max {
			return fmt.Errorf("%s: threshold.min cannot exceed threshold.max", prefix)
		}
		if threshold.RegressionPercent != nil {
			if !finite(*threshold.RegressionPercent) || *threshold.RegressionPercent < 0 {
				return fmt.Errorf("%s: threshold.regressionPercent must be a finite non-negative number", prefix)
			}
			if metric.Baseline == "" {
				return fmt.Errorf("%s: threshold.regressionPercent requires a baseline row", prefix)
			}
			if metric.normalizedDirection() == directionNone {
				return fmt.Errorf("%s: threshold.regressionPercent requires direction lower or higher", prefix)
			}
		}
	}
	return nil
}

func metricSpecByName(metrics []MetricSpec, name string) (MetricSpec, bool) {
	for _, metric := range metrics {
		if metric.Name == name {
			return metric, true
		}
	}
	return MetricSpec{}, false
}

func extractMetric(metric MetricSpec, variables map[string]any) (*float64, error) {
	output, err := gomplate.RunExpression(variables, gomplate.Template{
		Expression: metric.Extract,
		CelEnvs:    ANSICelFunctions(),
	})
	if err != nil {
		return nil, fmt.Errorf("evaluate extract expression %q: %w", metric.Extract, err)
	}
	value, ok := numericValue(output)
	if !ok {
		return nil, fmt.Errorf("extract expression %q returned %T; expected a number", metric.Extract, output)
	}
	if !finite(value) {
		return nil, fmt.Errorf("extract expression %q returned a non-finite number", metric.Extract)
	}
	return &value, nil
}

func numericValue(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func meanMetric(values []float64) float64 {
	var scale float64
	for _, value := range values {
		if absolute := math.Abs(value); absolute > scale {
			scale = absolute
		}
	}
	if scale == 0 {
		return 0
	}

	var total float64
	for _, value := range values {
		total += value / scale
	}
	return scale * (total / float64(len(values)))
}

func aggregateMetric(values []float64, aggregate string) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	switch aggregate {
	case metricMedian:
		middle := len(sorted) / 2
		if len(sorted)%2 == 0 {
			return meanMetric(sorted[middle-1 : middle+1])
		}
		return sorted[middle]
	case metricMin:
		return sorted[0]
	case metricMax:
		return sorted[len(sorted)-1]
	case metricP95:
		return sorted[int(math.Ceil(float64(len(sorted))*0.95))-1]
	default:
		return meanMetric(values)
	}
}

func summarizeMetrics(result *FixtureResult) {
	metrics := result.Test.metricSpecs()
	if len(metrics) == 0 {
		return
	}
	result.Metrics = make(map[string]MetricSummary, len(metrics))
	for _, metric := range metrics {
		summary := MetricSummary{
			Unit:      metric.Unit,
			Aggregate: metric.normalizedAggregate(),
			Direction: metric.normalizedDirection(),
			Samples:   make([]float64, 0, result.Test.repeatCount()),
			Threshold: metric.Threshold,
			Status:    OutcomeNotEvaluated,
		}
		var extractionErrors []string
		for _, sample := range result.Samples {
			extracted := sample.Metrics[metric.Name]
			if extracted.Value != nil && extracted.Status == OutcomePASS {
				summary.Samples = append(summary.Samples, *extracted.Value)
			}
			if extracted.Status == OutcomeERR {
				extractionErrors = append(extractionErrors, fmt.Sprintf("sample %d: %s", sample.Index, extracted.Error))
			}
		}
		if len(extractionErrors) > 0 {
			summary.Status = OutcomeERR
			summary.Error = strings.Join(extractionErrors, "; ")
			result.Metrics[metric.Name] = summary
			continue
		}
		if len(summary.Samples) != result.Test.repeatCount() {
			result.Metrics[metric.Name] = summary
			continue
		}

		value := aggregateMetric(summary.Samples, summary.Aggregate)
		if !finite(value) {
			summary.Status = OutcomeERR
			summary.Error = fmt.Sprintf("%s aggregate produced a non-finite number", summary.Aggregate)
			result.Metrics[metric.Name] = summary
			continue
		}
		summary.Value = &value
		summary.Status = OutcomePASS
		if metric.Threshold != nil {
			if metric.Threshold.Min != nil && value < *metric.Threshold.Min {
				summary.Status = OutcomeFAIL
				summary.Error = fmt.Sprintf("%s aggregate %.6g %s is below minimum %.6g %s", summary.Aggregate, value, metric.Unit, *metric.Threshold.Min, metric.Unit)
			}
			if metric.Threshold.Max != nil && value > *metric.Threshold.Max {
				summary.Status = OutcomeFAIL
				errorText := fmt.Sprintf("%s aggregate %.6g %s exceeds maximum %.6g %s", summary.Aggregate, value, metric.Unit, *metric.Threshold.Max, metric.Unit)
				summary.Error = joinErrors(summary.Error, errorText)
			}
		}
		result.Metrics[metric.Name] = summary
	}
}

// finalizeMetricComparisons runs only after every logical row has completed so
// callbacks and tree statistics see comparison-aware final verdicts.
func finalizeMetricComparisons(results []*FixtureResult) {
	byName := make(map[string][]*FixtureResult, len(results))
	for _, result := range results {
		byName[result.Name] = append(byName[result.Name], result)
	}

	for _, result := range results {
		for _, metric := range result.Test.metricSpecs() {
			if metric.Baseline == "" || metric.Baseline == result.Name {
				continue
			}
			summary := result.Metrics[metric.Name]
			if summary.Value == nil {
				continue
			}
			baselineResult := byName[metric.Baseline][0]
			baseline, ok := baselineResult.Metrics[metric.Name]
			if !ok || baseline.Value == nil {
				summary.Status = OutcomeERR
				summary.Error = joinErrors(summary.Error, fmt.Sprintf("baseline row %q did not produce a complete value", metric.Baseline))
				result.Metrics[metric.Name] = summary
				continue
			}
			comparison := &MetricComparison{
				Baseline:      metric.Baseline,
				BaselineValue: *baseline.Value,
				Status:        OutcomePASS,
			}
			if *baseline.Value == 0 {
				comparison.Status = OutcomeERR
				comparison.Error = "relative comparison to a zero baseline is undefined; use an absolute threshold"
				summary.Status = OutcomeERR
				summary.Error = joinErrors(summary.Error, comparison.Error)
				summary.Comparison = comparison
				result.Metrics[metric.Name] = summary
				continue
			}

			currentRatio := *summary.Value / math.Abs(*baseline.Value)
			baselineRatio := *baseline.Value / math.Abs(*baseline.Value)
			var regressionPercent float64
			if metric.normalizedDirection() == directionLower {
				regressionPercent = (currentRatio - baselineRatio) * 100
			} else {
				regressionPercent = (baselineRatio - currentRatio) * 100
			}
			if !finite(regressionPercent) {
				comparison.Status = OutcomeERR
				comparison.Error = "relative comparison produced a non-finite regression percentage"
				summary.Status = OutcomeERR
				summary.Error = joinErrors(summary.Error, comparison.Error)
				summary.Comparison = comparison
				result.Metrics[metric.Name] = summary
				continue
			}
			comparison.RegressionPercent = regressionPercent
			if metric.Threshold != nil && metric.Threshold.RegressionPercent != nil && comparison.RegressionPercent > *metric.Threshold.RegressionPercent {
				comparison.Status = OutcomeFAIL
				comparison.Error = fmt.Sprintf("regression %.2f%% exceeds maximum %.2f%%", comparison.RegressionPercent, *metric.Threshold.RegressionPercent)
				if summary.Status != OutcomeERR {
					summary.Status = OutcomeFAIL
				}
				summary.Error = joinErrors(summary.Error, comparison.Error)
			}
			summary.Comparison = comparison
			result.Metrics[metric.Name] = summary
		}
		finalizeLogicalResult(result)
	}
}

func metricOutcome(metrics map[string]MetricSummary) *FixtureOutcome {
	if len(metrics) == 0 {
		return nil
	}
	outcome := &FixtureOutcome{Status: OutcomePASS}
	var failures []string
	for name, metric := range metrics {
		switch metric.Status {
		case OutcomeERR:
			outcome.Status = OutcomeERR
		case OutcomeFAIL:
			if outcome.Status != OutcomeERR {
				outcome.Status = OutcomeFAIL
			}
		case OutcomeNotEvaluated:
			if outcome.Status == OutcomePASS {
				outcome.Status = OutcomeNotEvaluated
			}
		}
		if metric.Error != "" {
			failures = append(failures, fmt.Sprintf("%s: %s", name, metric.Error))
		}
	}
	sort.Strings(failures)
	outcome.Error = strings.Join(failures, "; ")
	return outcome
}

func finalizeLogicalResult(result *FixtureResult) {
	if result.Outcomes == nil {
		return
	}
	result.Outcomes.Metrics = metricOutcome(result.Metrics)
	result.Status = task.StatusPASS
	var errors []string
	for _, outcome := range []struct {
		name  string
		value *FixtureOutcome
	}{
		{name: "command", value: &result.Outcomes.Command},
		{name: "assertions", value: result.Outcomes.Assertions},
		{name: "metrics", value: result.Outcomes.Metrics},
	} {
		if outcome.value == nil {
			continue
		}
		if outcome.value.Status == OutcomeERR {
			result.Status = task.StatusERR
		} else if outcome.value.Status == OutcomeFAIL && result.Status != task.StatusERR {
			result.Status = task.StatusFAIL
		}
		if outcome.value.Error != "" {
			errors = append(errors, outcome.name+": "+outcome.value.Error)
		}
	}
	result.Error = strings.Join(errors, "; ")
}

func joinErrors(current, next string) string {
	if current == "" {
		return next
	}
	return current + "; " + next
}
