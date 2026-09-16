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
	measurementMean   = "mean"
	measurementMedian = "median"
	measurementMin    = "min"
	measurementMax    = "max"
	measurementP95    = "p95"

	directionLower  = "lower"
	directionHigher = "higher"
	directionNone   = "none"
)

// MeasurementThreshold applies absolute limits and an optional relative baseline limit.
type MeasurementThreshold struct {
	Min               *float64 `yaml:"min,omitempty" json:"min,omitempty"`
	Max               *float64 `yaml:"max,omitempty" json:"max,omitempty"`
	RegressionPercent *float64 `yaml:"regressionPercent,omitempty" json:"regressionPercent,omitempty"`
}

// MeasurementSpec describes a numeric CEL extraction and its logical-row policy.
type MeasurementSpec struct {
	Name      string                `yaml:"name" json:"name"`
	Extract   string                `yaml:"extract" json:"extract"`
	Unit      string                `yaml:"unit" json:"unit"`
	Aggregate string                `yaml:"aggregate,omitempty" json:"aggregate,omitempty"`
	Direction string                `yaml:"direction,omitempty" json:"direction,omitempty"`
	Baseline  string                `yaml:"baseline,omitempty" json:"baseline,omitempty"`
	Threshold *MeasurementThreshold `yaml:"threshold,omitempty" json:"threshold,omitempty"`
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

func (f FixtureTest) measurementSpecs() []MeasurementSpec {
	if f.Measurements != nil {
		return f.Measurements
	}
	return f.FrontMatter.Measurements
}

func (f FixtureTest) hasSampleConfiguration() bool {
	return f.Repeat != nil || f.FrontMatter.Repeat != nil || len(f.measurementSpecs()) > 0
}

func (m MeasurementSpec) normalizedAggregate() string {
	if m.Aggregate == "" {
		return measurementMean
	}
	return strings.ToLower(m.Aggregate)
}

func (m MeasurementSpec) normalizedDirection() string {
	if m.Direction == "" {
		return directionNone
	}
	return strings.ToLower(m.Direction)
}

// validateFixtureConfiguration rejects measurement policy errors before any command runs.
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
		if err := validateMeasurementSpecs(fixture); err != nil {
			return err
		}
	}

	for _, fixture := range fixtures {
		for _, measurement := range fixture.measurementSpecs() {
			if measurement.Baseline == "" {
				continue
			}
			if measurement.Baseline == fixture.Name {
				if len(fixture.Measurements) > 0 {
					return fmt.Errorf("fixture %q measurement %q: baseline must name another row", fixture.Name, measurement.Name)
				}
				continue
			}
			baselines := rowsByName[measurement.Baseline]
			switch len(baselines) {
			case 0:
				return fmt.Errorf("fixture %q measurement %q: baseline row %q was not found", fixture.Name, measurement.Name, measurement.Baseline)
			case 1:
			default:
				return fmt.Errorf("fixture %q measurement %q: baseline row name %q is ambiguous (%d rows)", fixture.Name, measurement.Name, measurement.Baseline, len(baselines))
			}

			baselineMeasurement, ok := measurementSpecByName(baselines[0].measurementSpecs(), measurement.Name)
			if !ok {
				return fmt.Errorf("fixture %q measurement %q: baseline row %q does not configure that measurement", fixture.Name, measurement.Name, measurement.Baseline)
			}
			if measurement.Unit != baselineMeasurement.Unit || measurement.normalizedAggregate() != baselineMeasurement.normalizedAggregate() || measurement.normalizedDirection() != baselineMeasurement.normalizedDirection() {
				return fmt.Errorf("fixture %q measurement %q: baseline row %q must use matching unit, aggregate, and direction", fixture.Name, measurement.Name, measurement.Baseline)
			}
		}
	}
	return nil
}

func validateMeasurementSpecs(fixture FixtureTest) error {
	seen := make(map[string]struct{})
	for _, measurement := range fixture.measurementSpecs() {
		prefix := fmt.Sprintf("fixture %q measurement %q", fixture.Name, measurement.Name)
		if measurement.Name == "" {
			return fmt.Errorf("fixture %q: measurement name is required", fixture.Name)
		}
		if _, ok := seen[measurement.Name]; ok {
			return fmt.Errorf("%s: duplicate measurement name", prefix)
		}
		seen[measurement.Name] = struct{}{}
		if measurement.Extract == "" {
			return fmt.Errorf("%s: extract is required", prefix)
		}
		if measurement.Unit == "" {
			return fmt.Errorf("%s: unit is required", prefix)
		}
		switch measurement.normalizedAggregate() {
		case measurementMean, measurementMedian, measurementMin, measurementMax, measurementP95:
		default:
			return fmt.Errorf("%s: unsupported aggregate %q (want mean, median, min, max, or p95)", prefix, measurement.Aggregate)
		}
		switch measurement.normalizedDirection() {
		case directionLower, directionHigher, directionNone:
		default:
			return fmt.Errorf("%s: unsupported direction %q (want lower, higher, or none)", prefix, measurement.Direction)
		}
		if measurement.Baseline != "" && measurement.normalizedDirection() == directionNone {
			return fmt.Errorf("%s: baseline comparison requires direction lower or higher", prefix)
		}
		if measurement.Threshold == nil {
			continue
		}
		threshold := measurement.Threshold
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
			if measurement.Baseline == "" {
				return fmt.Errorf("%s: threshold.regressionPercent requires a baseline row", prefix)
			}
			if measurement.normalizedDirection() == directionNone {
				return fmt.Errorf("%s: threshold.regressionPercent requires direction lower or higher", prefix)
			}
		}
	}
	return nil
}

func measurementSpecByName(measurements []MeasurementSpec, name string) (MeasurementSpec, bool) {
	for _, measurement := range measurements {
		if measurement.Name == name {
			return measurement, true
		}
	}
	return MeasurementSpec{}, false
}

func extractMeasurement(measurement MeasurementSpec, variables map[string]any) (*float64, error) {
	output, err := gomplate.RunExpression(variables, gomplate.Template{
		Expression: measurement.Extract,
		CelEnvs:    ANSICelFunctions(),
	})
	if err != nil {
		return nil, fmt.Errorf("evaluate extract expression %q: %w", measurement.Extract, err)
	}
	value, ok := numericValue(output)
	if !ok {
		return nil, fmt.Errorf("extract expression %q returned %T; expected a number", measurement.Extract, output)
	}
	if !finite(value) {
		return nil, fmt.Errorf("extract expression %q returned a non-finite number", measurement.Extract)
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

func meanMeasurement(values []float64) float64 {
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

func aggregateMeasurement(values []float64, aggregate string) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	switch aggregate {
	case measurementMedian:
		middle := len(sorted) / 2
		if len(sorted)%2 == 0 {
			return meanMeasurement(sorted[middle-1 : middle+1])
		}
		return sorted[middle]
	case measurementMin:
		return sorted[0]
	case measurementMax:
		return sorted[len(sorted)-1]
	case measurementP95:
		return sorted[int(math.Ceil(float64(len(sorted))*0.95))-1]
	default:
		return meanMeasurement(values)
	}
}

func summarizeMeasurements(result *FixtureResult) {
	measurements := result.Test.measurementSpecs()
	if len(measurements) == 0 {
		return
	}
	result.Measurements = make(map[string]MeasurementSummary, len(measurements))
	for _, measurement := range measurements {
		summary := MeasurementSummary{
			Unit:      measurement.Unit,
			Aggregate: measurement.normalizedAggregate(),
			Direction: measurement.normalizedDirection(),
			Samples:   make([]float64, 0, result.Test.repeatCount()),
			Threshold: measurement.Threshold,
			Status:    OutcomeNotEvaluated,
		}
		var extractionErrors []string
		for _, sample := range result.Samples {
			extracted := sample.Measurements[measurement.Name]
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
			result.Measurements[measurement.Name] = summary
			continue
		}
		if len(summary.Samples) != result.Test.repeatCount() {
			result.Measurements[measurement.Name] = summary
			continue
		}

		value := aggregateMeasurement(summary.Samples, summary.Aggregate)
		if !finite(value) {
			summary.Status = OutcomeERR
			summary.Error = fmt.Sprintf("%s aggregate produced a non-finite number", summary.Aggregate)
			result.Measurements[measurement.Name] = summary
			continue
		}
		summary.Value = &value
		summary.Status = OutcomePASS
		if measurement.Threshold != nil {
			if measurement.Threshold.Min != nil && value < *measurement.Threshold.Min {
				summary.Status = OutcomeFAIL
				summary.Error = fmt.Sprintf("%s aggregate %.6g %s is below minimum %.6g %s", summary.Aggregate, value, measurement.Unit, *measurement.Threshold.Min, measurement.Unit)
			}
			if measurement.Threshold.Max != nil && value > *measurement.Threshold.Max {
				summary.Status = OutcomeFAIL
				errorText := fmt.Sprintf("%s aggregate %.6g %s exceeds maximum %.6g %s", summary.Aggregate, value, measurement.Unit, *measurement.Threshold.Max, measurement.Unit)
				summary.Error = joinErrors(summary.Error, errorText)
			}
		}
		result.Measurements[measurement.Name] = summary
	}
}

// finalizeMeasurementComparisons runs only after every logical row has completed so
// callbacks and tree statistics see comparison-aware final verdicts.
func finalizeMeasurementComparisons(results []*FixtureResult) {
	byName := make(map[string][]*FixtureResult, len(results))
	for _, result := range results {
		byName[result.Name] = append(byName[result.Name], result)
	}

	for _, result := range results {
		for _, measurement := range result.Test.measurementSpecs() {
			if measurement.Baseline == "" || measurement.Baseline == result.Name {
				continue
			}
			summary := result.Measurements[measurement.Name]
			if summary.Value == nil {
				continue
			}
			baselineResult := byName[measurement.Baseline][0]
			baseline, ok := baselineResult.Measurements[measurement.Name]
			if !ok || baseline.Value == nil {
				summary.Status = OutcomeERR
				summary.Error = joinErrors(summary.Error, fmt.Sprintf("baseline row %q did not produce a complete value", measurement.Baseline))
				result.Measurements[measurement.Name] = summary
				continue
			}
			comparison := &MeasurementComparison{
				Baseline:      measurement.Baseline,
				BaselineValue: *baseline.Value,
				Status:        OutcomePASS,
			}
			if *baseline.Value == 0 {
				comparison.Status = OutcomeERR
				comparison.Error = "relative comparison to a zero baseline is undefined; use an absolute threshold"
				summary.Status = OutcomeERR
				summary.Error = joinErrors(summary.Error, comparison.Error)
				summary.Comparison = comparison
				result.Measurements[measurement.Name] = summary
				continue
			}

			currentRatio := *summary.Value / math.Abs(*baseline.Value)
			baselineRatio := *baseline.Value / math.Abs(*baseline.Value)
			var regressionPercent float64
			if measurement.normalizedDirection() == directionLower {
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
				result.Measurements[measurement.Name] = summary
				continue
			}
			comparison.RegressionPercent = regressionPercent
			if measurement.Threshold != nil && measurement.Threshold.RegressionPercent != nil {
				threshold := *measurement.Threshold.RegressionPercent
				// Permit arithmetic roundoff from the ratio and percentage operations
				// when the mathematical result is exactly on the threshold.
				scale := math.Max(1, math.Max(math.Abs(comparison.RegressionPercent), math.Abs(threshold)))
				tolerance := 16 * (math.Nextafter(1, 2) - 1) * scale
				if comparison.RegressionPercent-threshold > tolerance {
					comparison.Status = OutcomeFAIL
					comparison.Error = fmt.Sprintf("regression %.2f%% exceeds maximum %.2f%%", comparison.RegressionPercent, threshold)
					if summary.Status != OutcomeERR {
						summary.Status = OutcomeFAIL
					}
					summary.Error = joinErrors(summary.Error, comparison.Error)
				}
			}
			summary.Comparison = comparison
			result.Measurements[measurement.Name] = summary
		}
		finalizeLogicalResult(result)
	}
}

func measurementOutcome(measurements map[string]MeasurementSummary) *FixtureOutcome {
	if len(measurements) == 0 {
		return nil
	}
	outcome := &FixtureOutcome{Status: OutcomePASS}
	var failures []string
	for name, measurement := range measurements {
		switch measurement.Status {
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
		if measurement.Error != "" {
			failures = append(failures, fmt.Sprintf("%s: %s", name, measurement.Error))
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
	result.Outcomes.Measurements = measurementOutcome(result.Measurements)
	result.Status = task.StatusPASS
	var errors []string
	for _, outcome := range []struct {
		name  string
		value *FixtureOutcome
	}{
		{name: "command", value: &result.Outcomes.Command},
		{name: "assertions", value: result.Outcomes.Assertions},
		{name: "measurements", value: result.Outcomes.Measurements},
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
