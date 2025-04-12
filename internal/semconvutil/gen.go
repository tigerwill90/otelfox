// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package semconvutil // import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin/internal/semconvutil"

// Generate semconvutil package:
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconvutil/httpconv_test.go.tmpl "--data={}" --out=httpconv_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconvutil/httpconv.go.tmpl "--data={}" --out=httpconv.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconvutil/netconv_test.go.tmpl "--data={}" --out=netconv_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconvutil/netconv.go.tmpl "--data={}" --out=netconv.go
