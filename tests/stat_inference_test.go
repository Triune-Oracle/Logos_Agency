package tests

import (
	"strings"
	"testing"

	"github.com/Triune-Oracle/Logos_Agency/engine"
)

func TestBayesianInferenceEngine_InferType(t *testing.T) {
	tests := []struct {
		name         string
		values       []string
		expectedType engine.DataType
		minConfidence float64
	}{
		{
			name:          "Integer values",
			values:        []string{"1", "2", "3", "100", "200"},
			expectedType:  engine.TypeInteger,
			minConfidence: 0.8,
		},
		{
			name:          "Float values",
			values:        []string{"1.5", "2.3", "3.14", "99.99"},
			expectedType:  engine.TypeFloat,
			minConfidence: 0.8,
		},
		{
			name:          "String values",
			values:        []string{"hello", "world", "test", "data"},
			expectedType:  engine.TypeString,
			minConfidence: 0.8,
		},
		{
			name:          "Date values",
			values:        []string{"2024-01-01", "2024-02-15", "2024-03-20"},
			expectedType:  engine.TypeDate,
			minConfidence: 0.7,
		},
		{
			name:          "Boolean values",
			values:        []string{"true", "false", "true", "false", "true"},
			expectedType:  engine.TypeBoolean,
			minConfidence: 0.7,
		},
		{
			name:          "Mixed with empty",
			values:        []string{"1", "", "2", "", "3"},
			expectedType:  engine.TypeInteger,
			minConfidence: 0.7,
		},
		{
			name:          "All empty",
			values:        []string{"", "", ""},
			expectedType:  engine.TypeString,
			minConfidence: 0.0,
		},
		{
			name:          "Single value",
			values:        []string{"test"},
			expectedType:  engine.TypeString,
			minConfidence: 0.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())
			result := eng.InferType(tc.values)

			if result.InferredType != tc.expectedType {
				t.Errorf("Expected type %v, got %v", tc.expectedType, result.InferredType)
			}

			if result.Confidence < tc.minConfidence {
				t.Errorf("Expected confidence >= %v, got %v", tc.minConfidence, result.Confidence)
			}

			if len(result.Probabilities) == 0 {
				t.Error("Expected non-empty probabilities")
			}
		})
	}
}

func TestBayesianInferenceEngine_Deterministic(t *testing.T) {
	values := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

	// Run inference twice with same seed
	config1 := engine.DefaultConfig()
	config1.RandomSeed = 12345

	config2 := engine.DefaultConfig()
	config2.RandomSeed = 12345

	engine1 := engine.NewBayesianInferenceEngine(config1)
	engine2 := engine.NewBayesianInferenceEngine(config2)

	result1 := engine1.InferType(values)
	result2 := engine2.InferType(values)

	if result1.InferredType != result2.InferredType {
		t.Errorf("Expected deterministic type inference, got %v and %v",
			result1.InferredType, result2.InferredType)
	}

	if result1.Confidence != result2.Confidence {
		t.Errorf("Expected deterministic confidence, got %v and %v",
			result1.Confidence, result2.Confidence)
	}
}

func TestBayesianInferenceEngine_LocaleAware(t *testing.T) {
	values := []string{"100", "200", "300"}

	// Test different locales
	locales := []engine.Locale{engine.LocaleUS, engine.LocaleEU, engine.LocaleASIA, engine.LocaleGlobal}

	for _, locale := range locales {
		t.Run(string(locale), func(t *testing.T) {
			config := engine.DefaultConfig()
			config.Locale = locale

			eng := engine.NewBayesianInferenceEngine(config)
			result := eng.InferType(values)

			if result.InferredType != engine.TypeInteger {
				t.Errorf("Expected TypeInteger for locale %v, got %v", locale, result.InferredType)
			}

			// Verify locale-specific prior was used
			prior := engine.LocalePriors[locale]
			if prior.Integer == 0 {
				t.Errorf("Invalid prior for locale %v", locale)
			}
		})
	}
}

func TestBayesianInferenceEngine_SetLocale(t *testing.T) {
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	// Change locale
	eng.SetLocale(engine.LocaleASIA)

	values := []string{"1", "2", "3"}
	result := eng.InferType(values)

	if result.InferredType != engine.TypeInteger {
		t.Errorf("Expected TypeInteger after locale change, got %v", result.InferredType)
	}
}

func TestBayesianInferenceEngine_SetSeed(t *testing.T) {
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	values := []string{"1", "2", "3", "4", "5"}

	// First inference
	result1 := eng.InferType(values)

	// Change seed and infer again
	eng.SetSeed(99999)
	result2 := eng.InferType(values)

	// Results should be same for simple data
	if result1.InferredType != result2.InferredType {
		t.Errorf("Expected same type, got %v and %v", result1.InferredType, result2.InferredType)
	}
}

func TestBayesianInferenceEngine_LargeSample(t *testing.T) {
	// Create large dataset
	values := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		values[i] = "test"
	}

	config := engine.DefaultConfig()
	config.SampleSize = 100

	eng := engine.NewBayesianInferenceEngine(config)
	result := eng.InferType(values)

	if result.InferredType != engine.TypeString {
		t.Errorf("Expected TypeString, got %v", result.InferredType)
	}

	if result.SampleSize > 100 {
		t.Errorf("Expected sample size <= 100, got %v", result.SampleSize)
	}
}

func TestBayesianInferenceEngine_InferTypeSimple(t *testing.T) {
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	tests := []struct {
		values   []string
		expected string
	}{
		{[]string{"1", "2", "3"}, "int"},
		{[]string{"1.5", "2.5"}, "float"},
		{[]string{"hello"}, "string"},
		{[]string{"2024-01-01"}, "date"},
		{[]string{"true", "false"}, "boolean"},
	}

	for _, tc := range tests {
		result := eng.InferTypeSimple(tc.values)
		if result != tc.expected {
			t.Errorf("Expected %v, got %v for values %v", tc.expected, result, tc.values)
		}
	}
}

func TestDataType_String(t *testing.T) {
	tests := []struct {
		dataType engine.DataType
		expected string
	}{
		{engine.TypeInteger, "int"},
		{engine.TypeFloat, "float"},
		{engine.TypeString, "string"},
		{engine.TypeDate, "date"},
		{engine.TypeBoolean, "boolean"},
		{engine.TypeUnknown, "unknown"},
	}

	for _, tc := range tests {
		result := tc.dataType.String()
		if result != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, result)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := engine.DefaultConfig()

	if config.Locale != engine.LocaleGlobal {
		t.Errorf("Expected LocaleGlobal, got %v", config.Locale)
	}

	if config.SampleSize != 1000 {
		t.Errorf("Expected sample size 1000, got %v", config.SampleSize)
	}

	if config.ConfidenceThresh != 0.80 {
		t.Errorf("Expected confidence threshold 0.80, got %v", config.ConfidenceThresh)
	}

	if config.RandomSeed != 42 {
		t.Errorf("Expected random seed 42, got %v", config.RandomSeed)
	}

	if !config.EnableSIMD {
		t.Error("Expected SIMD to be enabled by default")
	}
}

func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		name          string
		probabilities []engine.TypeProbability
		minEntropy    float64
		maxEntropy    float64
	}{
		{
			name: "Uniform distribution",
			probabilities: []engine.TypeProbability{
				{Type: engine.TypeInteger, Probability: 0.25},
				{Type: engine.TypeFloat, Probability: 0.25},
				{Type: engine.TypeString, Probability: 0.25},
				{Type: engine.TypeDate, Probability: 0.25},
			},
			minEntropy: 1.9,
			maxEntropy: 2.1,
		},
		{
			name: "Deterministic",
			probabilities: []engine.TypeProbability{
				{Type: engine.TypeInteger, Probability: 1.0},
			},
			minEntropy: 0.0,
			maxEntropy: 0.01,
		},
		{
			name: "Binary split",
			probabilities: []engine.TypeProbability{
				{Type: engine.TypeInteger, Probability: 0.5},
				{Type: engine.TypeString, Probability: 0.5},
			},
			minEntropy: 0.9,
			maxEntropy: 1.1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entropy := engine.CalculateEntropy(tc.probabilities)

			if entropy < tc.minEntropy || entropy > tc.maxEntropy {
				t.Errorf("Expected entropy between %v and %v, got %v",
					tc.minEntropy, tc.maxEntropy, entropy)
			}
		})
	}
}

func TestBayesianInferenceEngine_GetConfidence(t *testing.T) {
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	values := []string{"1", "2", "3"}
	confidence := eng.GetConfidence(values)

	if confidence < 0.0 || confidence > 1.0 {
		t.Errorf("Expected confidence between 0 and 1, got %v", confidence)
	}
}

func TestBayesianInferenceEngine_IntegerFloatPreference(t *testing.T) {
	// Test that integers are preferred over floats when appropriate
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	values := []string{"1", "2", "3", "4", "5"}
	result := eng.InferType(values)

	if result.InferredType != engine.TypeInteger {
		t.Errorf("Expected TypeInteger for integer values, got %v", result.InferredType)
	}
}

func TestBayesianInferenceEngine_EdgeCases(t *testing.T) {
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	tests := []struct {
		name   string
		values []string
	}{
		{"Empty slice", []string{}},
		{"Single empty string", []string{""}},
		{"Multiple empty strings", []string{"", "", ""}},
		{"Whitespace only", []string{"   ", "\t", "\n"}},
		{"Very long string", []string{strings.Repeat("a", 10000)}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			result := eng.InferType(tc.values)

			if result.InferredType == engine.TypeUnknown && len(tc.values) > 0 {
				// Most edge cases should fall back to TypeString, not TypeUnknown
				t.Logf("Warning: Got TypeUnknown for edge case: %v", tc.name)
			}
		})
	}
}

func TestLocalePriors(t *testing.T) {
	// Verify all locales have valid priors
	locales := []engine.Locale{engine.LocaleUS, engine.LocaleEU, engine.LocaleASIA, engine.LocaleGlobal}

	for _, locale := range locales {
		prior, ok := engine.LocalePriors[locale]
		if !ok {
			t.Errorf("Missing prior for locale %v", locale)
			continue
		}

		// Check that priors sum to approximately 1.0
		sum := prior.Integer + prior.Float + prior.String + prior.Date + prior.Boolean
		if sum < 0.99 || sum > 1.01 {
			t.Errorf("Priors for locale %v sum to %v, expected ~1.0", locale, sum)
		}

		// Check all priors are non-negative
		if prior.Integer < 0 || prior.Float < 0 || prior.String < 0 ||
			prior.Date < 0 || prior.Boolean < 0 {
			t.Errorf("Negative prior for locale %v", locale)
		}
	}
}
