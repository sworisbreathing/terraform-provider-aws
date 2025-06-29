// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

const (
	summaryInvalidValue     = "Invalid value"
	summaryInvalidValueType = "Invalid value type"
)

func NewIncorrectValueTypeAttributeError(path cty.Path, expected string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		summaryInvalidValueType,
		"Expected type to be "+expected,
	)
}

func NewInvalidValueAttributeErrorf(path cty.Path, format string, a ...any) diag.Diagnostic {
	return NewInvalidValueAttributeError(
		path,
		fmt.Sprintf(format, a...),
	)
}

func NewInvalidValueAttributeError(path cty.Path, detail string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		summaryInvalidValue,
		detail,
	)
}

func NewAttributeErrorDiagnostic(path cty.Path, summary, detail string) diag.Diagnostic {
	return withPath(
		NewErrorDiagnostic(summary, detail),
		path,
	)
}

func NewAttributeWarningDiagnostic(path cty.Path, summary, detail string) diag.Diagnostic {
	return withPath(
		NewWarningDiagnostic(summary, detail),
		path,
	)
}

func NewErrorDiagnostic(summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   detail,
	}
}

func NewWarningDiagnostic(summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  summary,
		Detail:   detail,
	}
}

func withPath(d diag.Diagnostic, path cty.Path) diag.Diagnostic {
	d.AttributePath = path
	return d
}

// NewAttributeConflictsWhenError returns an error diagnostic indicating that the attribute at the given path cannot be
// specified when the attribute at otherPath has the given value.
func NewAttributeConflictsWhenError(path, otherPath cty.Path, otherValue string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		path,
		"Invalid Attribute Combination",
		fmt.Sprintf("Attribute %q cannot be specified when %q is %q.",
			PathString(path),
			PathString(otherPath),
			otherValue,
		),
	)
}

// NewAttributeRequiredWhenError returns an error diagnostic indicating that the attribute at neededPath is required when the
// attribute at otherPath has the given value.
func NewAttributeRequiredWhenError(neededPath, otherPath cty.Path, value string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		otherPath,
		"Invalid Attribute Combination",
		fmt.Sprintf("Attribute %q must be specified when %q is %q.",
			PathString(neededPath),
			PathString(otherPath),
			value,
		),
	)
}

// NewAtLeastOneOfChildrenError returns an error diagnostic indicating that at least on of the named children of
// parentPath is required.
func NewAtLeastOneOfChildrenError(parentPath cty.Path, paths ...cty.Path) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		parentPath,
		"Invalid Attribute Combination",
		fmt.Sprintf("At least one attribute out of [%s] must be specified", strings.Join(tfslices.ApplyToAll(paths, PathString), ", ")),
	)
}

// NewAttributeRequiredWhenError should only be used for apply-time validation, as it replicates
// the functionality of a `Required` attribute
func NewAttributeRequiredError(parentPath cty.Path, attrname string) diag.Diagnostic {
	return NewAttributeErrorDiagnostic(
		parentPath,
		"Missing required argument",
		fmt.Sprintf("The argument %q is required, but no definition was found.", attrname),
	)
}

// NewAttributeRequiredWillBeError returns a warning diagnostic indicating that the attribute at the given path is required.
// This is intended to be used for situations where the missing attribute will be an error in a future release.
func NewAttributeRequiredWillBeError(parentPath cty.Path, attrname string) diag.Diagnostic {
	return willBeError(
		NewAttributeRequiredError(parentPath, attrname),
	)
}

// NewAttributeConflictsWhenWillBeError returns a warning diagnostic indicating that the attribute at the given path cannot be
// specified when the attribute at otherPath has the given value.
// This is intended to be used for situations where the conflict will become an error in a future release.
func NewAttributeConflictsWhenWillBeError(path, otherPath cty.Path, otherValue string) diag.Diagnostic {
	return willBeError(
		NewAttributeConflictsWhenError(path, otherPath, otherValue),
	)
}

func PathString(path cty.Path) string {
	var buf strings.Builder
	for i, step := range path {
		switch x := step.(type) {
		case cty.GetAttrStep:
			if i != 0 {
				buf.WriteString(".")
			}
			buf.WriteString(x.Name)
		case cty.IndexStep:
			var s string
			switch val := x.Key; val.Type() {
			case cty.String:
				s = val.AsString()
			case cty.Number:
				num := val.AsBigFloat()
				s = num.String()
			default:
				s = fmt.Sprintf("<unexpected index: %s>", val.Type().FriendlyName())
			}
			fmt.Fprintf(&buf, "[%s]", s)
		default:
			if i != 0 {
				buf.WriteString(".")
			}
			fmt.Fprintf(&buf, "<unexpected step: %[1]T %[1]v>", x)
		}
	}
	return buf.String()
}

func errorToWarning(d diag.Diagnostic) diag.Diagnostic {
	d.Severity = diag.Warning
	return d
}

func willBeError(d diag.Diagnostic) diag.Diagnostic {
	d.Detail += "\n\nThis will be an error in a future release."
	return errorToWarning(d)
}
