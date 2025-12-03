package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
)

func BenchmarkTransformAlarmRuleSimple(b *testing.B) {
	input := `ALARM("SimpleAlarm")`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

func BenchmarkTransformAlarmRuleComplex(b *testing.B) {
	input := `((ALARM("A") AND ALARM("B")) OR (OK("C") AND OK("D"))) AND NOT INSUFFICIENT_DATA("E")`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkTransformAlarmRuleWithNewlines - benchmark z newlines
func BenchmarkTransformAlarmRuleWithNewlines(b *testing.B) {
	input := "ALARM(\n\"cpu-alarm\"\n) AND ALARM(\n\"memory-alarm\"\n)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkTransformAlarmRuleMultiple - benchmark wielu alarmów
func BenchmarkTransformAlarmRuleMultiple(b *testing.B) {
	input := `ALARM("A") AND ALARM("B") AND ALARM("C") AND ALARM("D") AND ALARM("E")`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkTransformAlarmRuleDeeplyNested - benchmark głęboko zagnieżdżonych wyrażeń
func BenchmarkTransformAlarmRuleDeeplyNested(b *testing.B) {
	input := `((((ALARM("A") OR ALARM("B")) AND ALARM("C")) OR ALARM("D")) AND ALARM("E"))`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkTransformAlarmRuleWithSpecialChars - benchmark ze specjalnymi znakami
func BenchmarkTransformAlarmRuleWithSpecialChars(b *testing.B) {
	input := `ALARM("arn:aws:cloudwatch:us-east-1:123456:alarm:prod/web/cpu-high")`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkFullPipeline - benchmark pełnego pipeline: Transform + Govaluate
func BenchmarkFullPipeline(b *testing.B) {
	input := `ALARM("cpu-high") AND ALARM("memory-high")`

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			return true, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transformed := TransformAlarmRule(input)
		expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
		_, _ = expr.Evaluate(nil)
	}
}

// BenchmarkFullPipelineComplex - benchmark pełnego pipeline dla złożonych wyrażeń
func BenchmarkFullPipelineComplex(b *testing.B) {
	input := `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			return true, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transformed := TransformAlarmRule(input)
		expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
		_, _ = expr.Evaluate(nil)
	}
}

// BenchmarkGovaluateOnly - benchmark tylko govaluate (bez transformacji)
func BenchmarkGovaluateOnly(b *testing.B) {
	expression := `ALARM('cpu-high') && ALARM('memory-high')`

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			return true, nil
		},
	}

	expr, _ := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(nil)
	}
}

// BenchmarkTransformOnly - benchmark tylko transformacji (bez govaluate)
func BenchmarkTransformOnly(b *testing.B) {
	input := `ALARM("cpu-high") AND ALARM("memory-high")`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkTransformWithQuotes - benchmark transformacji z różnymi cudzysłowami
func BenchmarkTransformWithQuotes(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "DoubleQuotes",
			input: `ALARM("test")`,
		},
		{
			name:  "SingleQuotes",
			input: `ALARM('test')`,
		},
		{
			name:  "NoQuotes",
			input: `ALARM(test)`,
		},
		{
			name:  "NestedQuotes",
			input: `ALARM("'test'")`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TransformAlarmRule(tc.input)
			}
		})
	}
}

// BenchmarkTransformOperators - benchmark transformacji różnych operatorów
func BenchmarkTransformOperators(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "AND",
			input: `ALARM("A") AND ALARM("B")`,
		},
		{
			name:  "OR",
			input: `ALARM("A") OR ALARM("B")`,
		},
		{
			name:  "NOT",
			input: `NOT ALARM("A")`,
		},
		{
			name:  "Mixed",
			input: `ALARM("A") AND NOT ALARM("B") OR ALARM("C")`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TransformAlarmRule(tc.input)
			}
		})
	}
}

// BenchmarkTransformAlarmCount - benchmark w zależności od liczby alarmów
func BenchmarkTransformAlarmCount(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "1Alarm",
			input: `ALARM("alarm1")`,
		},
		{
			name:  "2Alarms",
			input: `ALARM("alarm1") AND ALARM("alarm2")`,
		},
		{
			name:  "5Alarms",
			input: `ALARM("alarm1") AND ALARM("alarm2") AND ALARM("alarm3") AND ALARM("alarm4") AND ALARM("alarm5")`,
		},
		{
			name:  "10Alarms",
			input: `ALARM("alarm1") AND ALARM("alarm2") AND ALARM("alarm3") AND ALARM("alarm4") AND ALARM("alarm5") AND ALARM("alarm6") AND ALARM("alarm7") AND ALARM("alarm8") AND ALARM("alarm9") AND ALARM("alarm10")`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TransformAlarmRule(tc.input)
			}
		})
	}
}

// BenchmarkTransformNestingDepth - benchmark w zależności od głębokości zagnieżdżenia
func BenchmarkTransformNestingDepth(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Depth1",
			input: `ALARM("A")`,
		},
		{
			name:  "Depth2",
			input: `(ALARM("A") AND ALARM("B"))`,
		},
		{
			name:  "Depth3",
			input: `((ALARM("A") AND ALARM("B")) AND ALARM("C"))`,
		},
		{
			name:  "Depth5",
			input: `((((ALARM("A") AND ALARM("B")) AND ALARM("C")) AND ALARM("D")) AND ALARM("E"))`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TransformAlarmRule(tc.input)
			}
		})
	}
}

// BenchmarkTransformAlarmNameLength - benchmark w zależności od długości nazwy alarmu
func BenchmarkTransformAlarmNameLength(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Short",
			input: `ALARM("cpu")`,
		},
		{
			name:  "Medium",
			input: `ALARM("production-web-server-cpu-high")`,
		},
		{
			name:  "Long",
			input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:production-web-server-cpu-utilization-high")`,
		},
		{
			name:  "VeryLong",
			input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:production-environment-web-application-server-cluster-cpu-utilization-threshold-exceeded")`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = TransformAlarmRule(tc.input)
			}
		})
	}
}

// BenchmarkFullPipelineVariations - benchmark różnych scenariuszy pełnego pipeline
func BenchmarkFullPipelineVariations(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Simple",
			input: `ALARM("cpu")`,
		},
		{
			name:  "TwoAlarmsAND",
			input: `ALARM("cpu") AND ALARM("memory")`,
		},
		{
			name:  "TwoAlarmsOR",
			input: `ALARM("cpu") OR ALARM("memory")`,
		},
		{
			name:  "WithNOT",
			input: `ALARM("cpu") AND NOT ALARM("maintenance")`,
		},
		{
			name:  "Nested",
			input: `(ALARM("cpu1") OR ALARM("cpu2")) AND NOT ALARM("deploy")`,
		},
		{
			name:  "Complex",
			input: `((ALARM("A") AND ALARM("B")) OR (OK("C") AND OK("D"))) AND NOT INSUFFICIENT_DATA("E")`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					return true, nil
				},
				"OK": func(args ...interface{}) (interface{}, error) {
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					return true, nil
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				transformed := TransformAlarmRule(tc.input)
				expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
				_, _ = expr.Evaluate(nil)
			}
		})
	}
}

// BenchmarkMemoryAllocation - benchmark alokacji pamięci
func BenchmarkMemoryAllocation(b *testing.B) {
	input := `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = TransformAlarmRule(input)
	}
}

// BenchmarkConcurrentTransform - benchmark współbieżnych transformacji
func BenchmarkConcurrentTransform(b *testing.B) {
	input := `ALARM("cpu-high") AND ALARM("memory-high")`

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = TransformAlarmRule(input)
		}
	})
}

// BenchmarkConcurrentFullPipeline - benchmark współbieżnego pełnego pipeline
func BenchmarkConcurrentFullPipeline(b *testing.B) {
	input := `ALARM("cpu-high") AND ALARM("memory-high")`

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			return true, nil
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			transformed := TransformAlarmRule(input)
			expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			_, _ = expr.Evaluate(nil)
		}
	})
}

// BenchmarkRealWorldScenarios - benchmark rzeczywistych scenariuszy
func BenchmarkRealWorldScenarios(b *testing.B) {
	testCases := []struct {
		name     string
		input    string
		scenario string
	}{
		{
			name:     "HighAvailability",
			input:    `ALARM("primary-down") AND ALARM("backup-down")`,
			scenario: "Both primary and backup servers down",
		},
		{
			name:     "MaintenanceWindow",
			input:    `ALARM("cpu-high") AND NOT ALARM("maintenance")`,
			scenario: "High CPU but not during maintenance",
		},
		{
			name:     "MultiRegion",
			input:    `ALARM("us-east-down") OR ALARM("eu-west-down") OR ALARM("ap-southeast-down")`,
			scenario: "Any region is down",
		},
		{
			name:     "ServiceDependency",
			input:    `ALARM("service-down") AND OK("database-healthy")`,
			scenario: "Service down but database healthy",
		},
		{
			name:     "ComplexProduction",
			input:    `(ALARM("prod-cpu") OR ALARM("prod-mem")) AND NOT (ALARM("known-issue") OR ALARM("deployment"))`,
			scenario: "Production issues excluding known problems",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					return true, nil
				},
				"OK": func(args ...interface{}) (interface{}, error) {
					return true, nil
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				transformed := TransformAlarmRule(tc.input)
				expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
				_, _ = expr.Evaluate(nil)
			}
		})
	}
}
