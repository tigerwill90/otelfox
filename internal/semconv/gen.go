// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package semconv // import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin/internal/semconv"

// Generate semconv package:
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/bench_test.go.tmpl "--data={}" --out=bench_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/common_test.go.tmpl "--data={ \"pkg\": \"github.com/tigerwill90/otelfox\" }" --out=common_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/env.go.tmpl "--data={ \"pkg\": \"github.com/tigerwill90/otelfox\" }" --out=env.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/env_test.go.tmpl "--data={}" --out=env_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/httpconv.go.tmpl "--data={ \"pkg\": \"github.com/tigerwill90/otelfox\" }" --out=httpconv.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/httpconv_test.go.tmpl "--data={}" --out=httpconv_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/httpconvtest_test.go.tmpl "--data={ \"pkg\": \"github.com/tigerwill90/otelfox\" }" --out=httpconvtest_test.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/util.go.tmpl "--data={ \"pkg\": \"github.com/tigerwill90/otelfox\" }" --out=util.go
//go:generate go tool -modfile=../../go.tool.mod gotmpl --body=../shared/semconv/util_test.go.tmpl "--data={}" --out=util_test.go
