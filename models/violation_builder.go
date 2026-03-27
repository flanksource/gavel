package models

import "time"

// ViolationBuilder provides a fluent API for constructing Violations from linter output.
type ViolationBuilder struct {
	violation Violation
}

func NewViolationBuilder() *ViolationBuilder {
	return &ViolationBuilder{
		violation: Violation{CreatedAt: time.Now()},
	}
}

func (vb *ViolationBuilder) WithFile(file string) *ViolationBuilder {
	vb.violation.File = file
	return vb
}

func (vb *ViolationBuilder) WithLocation(line, column int) *ViolationBuilder {
	vb.violation.Line = line
	vb.violation.Column = column
	return vb
}

func (vb *ViolationBuilder) WithMessage(message string) *ViolationBuilder {
	vb.violation.Message = &message
	return vb
}

func (vb *ViolationBuilder) WithSource(source string) *ViolationBuilder {
	vb.violation.Source = source
	return vb
}

func (vb *ViolationBuilder) WithCaller(pkg, method string) *ViolationBuilder {
	return vb
}

func (vb *ViolationBuilder) WithCalled(pkg, method string) *ViolationBuilder {
	return vb
}

func (vb *ViolationBuilder) WithRuleFromLinter(linterName, ruleName string) *ViolationBuilder {
	vb.violation.Rule = &Rule{
		Type:    RuleTypeDeny,
		Package: linterName,
		Method:  ruleName,
	}
	return vb
}

func (vb *ViolationBuilder) WithFixable(fixable bool) *ViolationBuilder {
	vb.violation.Fixable = fixable
	return vb
}

func (vb *ViolationBuilder) WithFixApplicability(applicability string) *ViolationBuilder {
	vb.violation.FixApplicability = applicability
	return vb
}

func (vb *ViolationBuilder) WithCode(code string) *ViolationBuilder {
	vb.violation.Code = &code
	return vb
}

func (vb *ViolationBuilder) Build() Violation {
	return vb.violation
}
