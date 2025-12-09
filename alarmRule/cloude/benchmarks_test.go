package alarmrule

// import (
// 	"testing"

// 	"github.com/Knetic/govaluate"
// )

// // BenchmarkTransformAlarmRuleSimple benchmarks simple alarm transformation.
// func BenchmarkTransformAlarmRuleSimple(b *testing.B) {
// 	input := `ALARM("simple-alarm")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkTransformAlarmRuleComplex benchmarks complex alarm rule transformation.
// func BenchmarkTransformAlarmRuleComplex(b *testing.B) {
// 	input := `((ALARM("A") AND ALARM("B")) OR (OK("C") AND OK("D"))) AND NOT INSUFFICIENT_DATA("E")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkTransformAlarmRuleWithNewlines benchmarks transformation with newlines.
// func BenchmarkTransformAlarmRuleWithNewlines(b *testing.B) {
// 	input := "ALARM(\n\"cpu-alarm\"\n) AND ALARM(\n\"memory-alarm\"\n)"

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkTransformAlarmRuleMultiple benchmarks transformation with multiple alarms.
// func BenchmarkTransformAlarmRuleMultiple(b *testing.B) {
// 	input := `ALARM("A") AND ALARM("B") AND ALARM("C") AND ALARM("D") AND ALARM("E")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkTransformAlarmRuleDeeplyNested benchmarks deeply nested expressions.
// func BenchmarkTransformAlarmRuleDeeplyNested(b *testing.B) {
// 	input := `((((ALARM("A") OR ALARM("B")) AND ALARM("C")) OR ALARM("D")) AND ALARM("E"))`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkTransformAlarmRuleWithSpecialChars benchmarks transformation with special characters.
// func BenchmarkTransformAlarmRuleWithSpecialChars(b *testing.B) {
// 	input := `ALARM("arn:aws:cloudwatch:us-east-1:123456:alarm:prod/web/cpu-high")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkFullPipeline benchmarks the complete pipeline: Transform + Govaluate.
// func BenchmarkFullPipeline(b *testing.B) {
// 	input := `ALARM("cpu-high") AND ALARM("memory-high")`

// 	functions := map[string]govaluate.ExpressionFunction{
// 		"ALARM": func(args ...interface{}) (interface{}, error) {
// 			return true, nil
// 		},
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		transformed := TransformAlarmRule(input)
// 		expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
// 		_, _ = expr.Evaluate(nil)
// 	}
// }

// // BenchmarkFullPipelineComplex benchmarks full pipeline for complex expressions.
// func BenchmarkFullPipelineComplex(b *testing.B) {
// 	input := `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`

// 	functions := map[string]govaluate.ExpressionFunction{
// 		"ALARM": func(args ...interface{}) (interface{}, error) {
// 			return true, nil
// 		},
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		transformed := TransformAlarmRule(input)
// 		expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
// 		_, _ = expr.Evaluate(nil)
// 	}
// }

// // BenchmarkGovaluateOnly benchmarks only govaluate evaluation (without transformation).
// func BenchmarkGovaluateOnly(b *testing.B) {
// 	expression := `ALARM('cpu-high') && ALARM('memory-high')`

// 	functions := map[string]govaluate.ExpressionFunction{
// 		"ALARM": func(args ...interface{}) (interface{}, error) {
// 			return true, nil
// 		},
// 	}

// 	expr, _ := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, _ = expr.Evaluate(nil)
// 	}
// }

// // BenchmarkTransformOnly benchmarks only transformation (without govaluate).
// func BenchmarkTransformOnly(b *testing.B) {
// 	input := `ALARM("cpu-high") AND ALARM("memory-high")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkTransformWithQuotes benchmarks transformation with different quote styles.
// func BenchmarkTransformWithQuotes(b *testing.B) {
// 	testCases := []struct {
// 		name  string
// 		input string
// 	}{
// 		{
// 			name:  "DoubleQuotes",
// 			input: `ALARM("test")`,
// 		},
// 		{
// 			name:  "SingleQuotes",
// 			input: `ALARM('test')`,
// 		},
// 		{
// 			name:  "NoQuotes",
// 			input: `ALARM(test)`,
// 		},
// 		{
// 			name:  "NestedQuotes",
// 			input: `ALARM("'test'")`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				_ = TransformAlarmRule(tc.input)
// 			}
// 		})
// 	}
// }

// // BenchmarkTransformOperators benchmarks transformation of different operators.
// func BenchmarkTransformOperators(b *testing.B) {
// 	testCases := []struct {
// 		name  string
// 		input string
// 	}{
// 		{
// 			name:  "AND",
// 			input: `ALARM("A") AND ALARM("B")`,
// 		},
// 		{
// 			name:  "OR",
// 			input: `ALARM("A") OR ALARM("B")`,
// 		},
// 		{
// 			name:  "NOT",
// 			input: `NOT ALARM("A")`,
// 		},
// 		{
// 			name:  "Mixed",
// 			input: `ALARM("A") AND NOT ALARM("B") OR ALARM("C")`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				_ = TransformAlarmRule(tc.input)
// 			}
// 		})
// 	}
// }

// // BenchmarkTransformAlarmCount benchmarks transformation based on number of alarms.
// func BenchmarkTransformAlarmCount(b *testing.B) {
// 	testCases := []struct {
// 		name  string
// 		input string
// 	}{
// 		{
// 			name:  "1Alarm",
// 			input: `ALARM("alarm1")`,
// 		},
// 		{
// 			name:  "2Alarms",
// 			input: `ALARM("alarm1") AND ALARM("alarm2")`,
// 		},
// 		{
// 			name:  "5Alarms",
// 			input: `ALARM("alarm1") AND ALARM("alarm2") AND ALARM("alarm3") AND ALARM("alarm4") AND ALARM("alarm5")`,
// 		},
// 		{
// 			name:  "10Alarms",
// 			input: `ALARM("alarm1") AND ALARM("alarm2") AND ALARM("alarm3") AND ALARM("alarm4") AND ALARM("alarm5") AND ALARM("alarm6") AND ALARM("alarm7") AND ALARM("alarm8") AND ALARM("alarm9") AND ALARM("alarm10")`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				_ = TransformAlarmRule(tc.input)
// 			}
// 		})
// 	}
// }

// // BenchmarkTransformNestingDepth benchmarks transformation based on nesting depth.
// func BenchmarkTransformNestingDepth(b *testing.B) {
// 	testCases := []struct {
// 		name  string
// 		input string
// 	}{
// 		{
// 			name:  "Depth1",
// 			input: `ALARM("A")`,
// 		},
// 		{
// 			name:  "Depth2",
// 			input: `(ALARM("A") AND ALARM("B"))`,
// 		},
// 		{
// 			name:  "Depth3",
// 			input: `((ALARM("A") AND ALARM("B")) AND ALARM("C"))`,
// 		},
// 		{
// 			name:  "Depth5",
// 			input: `((((ALARM("A") AND ALARM("B")) AND ALARM("C")) AND ALARM("D")) AND ALARM("E"))`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				_ = TransformAlarmRule(tc.input)
// 			}
// 		})
// 	}
// }

// // BenchmarkTransformAlarmNameLength benchmarks transformation based on alarm name length.
// func BenchmarkTransformAlarmNameLength(b *testing.B) {
// 	testCases := []struct {
// 		name  string
// 		input string
// 	}{
// 		{
// 			name:  "Short",
// 			input: `ALARM("cpu")`,
// 		},
// 		{
// 			name:  "Medium",
// 			input: `ALARM("production-web-server-cpu-high")`,
// 		},
// 		{
// 			name:  "Long",
// 			input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:production-web-server-cpu-utilization-high")`,
// 		},
// 		{
// 			name:  "VeryLong",
// 			input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:production-environment-web-application-server-cluster-cpu-utilization-threshold-exceeded")`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				_ = TransformAlarmRule(tc.input)
// 			}
// 		})
// 	}
// }

// // BenchmarkFullPipelineVariations benchmarks full pipeline with different scenarios.
// func BenchmarkFullPipelineVariations(b *testing.B) {
// 	testCases := []struct {
// 		name  string
// 		input string
// 	}{
// 		{
// 			name:  "Simple",
// 			input: `ALARM("cpu")`,
// 		},
// 		{
// 			name:  "TwoAlarmsAND",
// 			input: `ALARM("cpu") AND ALARM("memory")`,
// 		},
// 		{
// 			name:  "TwoAlarmsOR",
// 			input: `ALARM("cpu") OR ALARM("memory")`,
// 		},
// 		{
// 			name:  "WithNOT",
// 			input: `ALARM("cpu") AND NOT ALARM("maintenance")`,
// 		},
// 		{
// 			name:  "Nested",
// 			input: `(ALARM("cpu1") OR ALARM("cpu2")) AND NOT ALARM("deploy")`,
// 		},
// 		{
// 			name:  "Complex",
// 			input: `((ALARM("A") AND ALARM("B")) OR (OK("C") AND OK("D"))) AND NOT INSUFFICIENT_DATA("E")`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			functions := map[string]govaluate.ExpressionFunction{
// 				"ALARM": func(args ...interface{}) (interface{}, error) {
// 					return true, nil
// 				},
// 				"OK": func(args ...interface{}) (interface{}, error) {
// 					return true, nil
// 				},
// 				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
// 					return true, nil
// 				},
// 			}

// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				transformed := TransformAlarmRule(tc.input)
// 				expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
// 				_, _ = expr.Evaluate(nil)
// 			}
// 		})
// 	}
// }

// // BenchmarkMemoryAllocation benchmarks memory allocations during transformation.
// func BenchmarkMemoryAllocation(b *testing.B) {
// 	input := `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }

// // BenchmarkConcurrentTransform benchmarks concurrent transformation operations.
// func BenchmarkConcurrentTransform(b *testing.B) {
// 	input := `ALARM("cpu-high") AND ALARM("memory-high")`

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			_ = TransformAlarmRule(input)
// 		}
// 	})
// }

// // BenchmarkConcurrentFullPipeline benchmarks concurrent full pipeline operations.
// func BenchmarkConcurrentFullPipeline(b *testing.B) {
// 	input := `ALARM("cpu-high") AND ALARM("memory-high")`

// 	functions := map[string]govaluate.ExpressionFunction{
// 		"ALARM": func(args ...interface{}) (interface{}, error) {
// 			return true, nil
// 		},
// 	}

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			transformed := TransformAlarmRule(input)
// 			expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
// 			_, _ = expr.Evaluate(nil)
// 		}
// 	})
// }

// // BenchmarkRealWorldScenarios benchmarks real-world AWS alarm scenarios.
// func BenchmarkRealWorldScenarios(b *testing.B) {
// 	testCases := []struct {
// 		name     string
// 		input    string
// 		scenario string
// 	}{
// 		{
// 			name:     "HighAvailability",
// 			input:    `ALARM("primary-down") AND ALARM("backup-down")`,
// 			scenario: "Both primary and backup servers down",
// 		},
// 		{
// 			name:     "MaintenanceWindow",
// 			input:    `ALARM("cpu-high") AND NOT ALARM("maintenance")`,
// 			scenario: "High CPU but not during maintenance",
// 		},
// 		{
// 			name:     "MultiRegion",
// 			input:    `ALARM("us-east-down") OR ALARM("eu-west-down") OR ALARM("ap-southeast-down")`,
// 			scenario: "Any region is down",
// 		},
// 		{
// 			name:     "ServiceDependency",
// 			input:    `ALARM("service-down") AND OK("database-healthy")`,
// 			scenario: "Service down but database healthy",
// 		},
// 		{
// 			name:     "ComplexProduction",
// 			input:    `(ALARM("prod-cpu") OR ALARM("prod-mem")) AND NOT (ALARM("known-issue") OR ALARM("deployment"))`,
// 			scenario: "Production issues excluding known problems",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		b.Run(tc.name, func(b *testing.B) {
// 			functions := map[string]govaluate.ExpressionFunction{
// 				"ALARM": func(args ...interface{}) (interface{}, error) {
// 					return true, nil
// 				},
// 				"OK": func(args ...interface{}) (interface{}, error) {
// 					return true, nil
// 				},
// 			}

// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				transformed := TransformAlarmRule(tc.input)
// 				expr, _ := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
// 				_, _ = expr.Evaluate(nil)
// 			}
// 		})
// 	}
// }
