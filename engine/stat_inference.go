package engine

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DataType represents the inferred type of a column
type DataType int

const (
	TypeUnknown DataType = iota
	TypeInteger
	TypeFloat
	TypeString
	TypeDate
	TypeBoolean
)

// String returns string representation of DataType
func (dt DataType) String() string {
	switch dt {
	case TypeInteger:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeDate:
		return "date"
	case TypeBoolean:
		return "boolean"
	default:
		return "unknown"
	}
}

// Locale represents regional data patterns
type Locale string

const (
	LocaleUS     Locale = "en_US"
	LocaleEU     Locale = "en_EU"
	LocaleASIA   Locale = "en_ASIA"
	LocaleGlobal Locale = "global"
)

// TypePrior represents Bayesian prior probabilities for each type
type TypePrior struct {
	Integer float64
	Float   float64
	String  float64
	Date    float64
	Boolean float64
}

// LocalePriors defines prior probabilities for different locales
var LocalePriors = map[Locale]TypePrior{
	LocaleUS: {
		Integer: 0.25,
		Float:   0.20,
		String:  0.40,
		Date:    0.10,
		Boolean: 0.05,
	},
	LocaleEU: {
		Integer: 0.25,
		Float:   0.20,
		String:  0.40,
		Date:    0.10,
		Boolean: 0.05,
	},
	LocaleASIA: {
		Integer: 0.30,
		Float:   0.15,
		String:  0.40,
		Date:    0.10,
		Boolean: 0.05,
	},
	LocaleGlobal: {
		Integer: 0.25,
		Float:   0.20,
		String:  0.40,
		Date:    0.10,
		Boolean: 0.05,
	},
}

// InferenceConfig holds configuration for statistical inference
type InferenceConfig struct {
	Locale           Locale
	SampleSize       int
	ConfidenceThresh float64
	RandomSeed       int64
	EnableSIMD       bool
}

// DefaultConfig returns default inference configuration
func DefaultConfig() *InferenceConfig {
	return &InferenceConfig{
		Locale:           LocaleGlobal,
		SampleSize:       1000,
		ConfidenceThresh: 0.80,
		RandomSeed:       42, // Deterministic by default
		EnableSIMD:       true,
	}
}

// TypeProbability represents probability of a value being a specific type
type TypeProbability struct {
	Type        DataType
	Probability float64
	Evidence    float64
}

// InferenceResult contains the result of statistical type inference
type InferenceResult struct {
	InferredType  DataType
	Confidence    float64
	Probabilities []TypeProbability
	SampleSize    int
}

// BayesianInferenceEngine performs statistical type inference using Bayesian methods
type BayesianInferenceEngine struct {
	config *InferenceConfig
	rng    *rand.Rand
	mu     sync.RWMutex
}

// NewBayesianInferenceEngine creates a new inference engine
func NewBayesianInferenceEngine(config *InferenceConfig) *BayesianInferenceEngine {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Create deterministic random number generator
	source := rand.NewSource(config.RandomSeed)
	rng := rand.New(source)
	
	return &BayesianInferenceEngine{
		config: config,
		rng:    rng,
	}
}

// InferType performs Bayesian statistical inference on a column of values
func (e *BayesianInferenceEngine) InferType(values []string) *InferenceResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Sample values if dataset is large
	sampled := e.sampleValues(values)
	
	// Calculate likelihood for each type
	likelihoods := e.calculateLikelihoods(sampled)
	
	// Apply Bayesian inference with priors
	posteriors := e.calculatePosteriors(likelihoods)
	
	// Determine most likely type
	inferredType, confidence := e.selectType(posteriors)
	
	return &InferenceResult{
		InferredType:  inferredType,
		Confidence:    confidence,
		Probabilities: posteriors,
		SampleSize:    len(sampled),
	}
}

// sampleValues samples values from the dataset
func (e *BayesianInferenceEngine) sampleValues(values []string) []string {
	if len(values) <= e.config.SampleSize {
		return values
	}
	
	// Deterministic sampling using configured random seed
	sampled := make([]string, e.config.SampleSize)
	indices := e.rng.Perm(len(values))[:e.config.SampleSize]
	
	for i, idx := range indices {
		sampled[i] = values[idx]
	}
	
	return sampled
}

// calculateLikelihoods calculates likelihood of data for each type
func (e *BayesianInferenceEngine) calculateLikelihoods(values []string) map[DataType]float64 {
	if e.config.EnableSIMD && len(values) > 100 {
		return CalculateLikelihoodsSIMD(values)
	}
	
	likelihoods := make(map[DataType]float64)
	
	// Count successful parses for each type
	counts := map[DataType]int{
		TypeInteger: 0,
		TypeFloat:   0,
		TypeString:  0,
		TypeDate:    0,
		TypeBoolean: 0,
	}
	
	nonEmpty := 0
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		nonEmpty++
		
		isInt := false
		isFloat := false
		isBool := false
		isDate := false
		
		// Check integer first (most specific numeric type)
		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			counts[TypeInteger]++
			isInt = true
			// Don't check float if it's an integer (integers are more specific)
		} else {
			// Only check float if not an integer
			if _, err := strconv.ParseFloat(v, 64); err == nil {
				counts[TypeFloat]++
				isFloat = true
			}
		}
		
		// Check boolean
		lower := strings.ToLower(v)
		if lower == "true" || lower == "false" || lower == "t" || lower == "f" ||
			lower == "yes" || lower == "no" || lower == "y" || lower == "n" {
			counts[TypeBoolean]++
			isBool = true
		}
		
		// Check date
		if e.isDate(v) {
			counts[TypeDate]++
			isDate = true
		}
		
		// Count as string only if it's NOT a specific type
		if !isInt && !isFloat && !isBool && !isDate {
			counts[TypeString]++
		}
	}
	
	// Calculate likelihoods as proportion of successful parses
	if nonEmpty > 0 {
		for typ, count := range counts {
			likelihoods[typ] = float64(count) / float64(nonEmpty)
		}
	} else {
		// All empty - default to string
		likelihoods[TypeString] = 1.0
	}
	
	return likelihoods
}

// isDate checks if a value is a valid date
func (e *BayesianInferenceEngine) isDate(value string) bool {
	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"02-01-2006",
		"Jan 2 2006",
		"January 2 2006",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	
	for _, format := range dateFormats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}
	
	return false
}

// calculatePosteriors applies Bayesian inference to calculate posterior probabilities
func (e *BayesianInferenceEngine) calculatePosteriors(likelihoods map[DataType]float64) []TypeProbability {
	priors := LocalePriors[e.config.Locale]
	
	// Apply Bayes' theorem: P(Type|Data) ∝ P(Data|Type) * P(Type)
	unnormalized := map[DataType]float64{
		TypeInteger: likelihoods[TypeInteger] * priors.Integer,
		TypeFloat:   likelihoods[TypeFloat] * priors.Float,
		TypeString:  likelihoods[TypeString] * priors.String,
		TypeDate:    likelihoods[TypeDate] * priors.Date,
		TypeBoolean: likelihoods[TypeBoolean] * priors.Boolean,
	}
	
	// Normalize to get probabilities
	total := 0.0
	for _, prob := range unnormalized {
		total += prob
	}
	
	probabilities := make([]TypeProbability, 0, len(unnormalized))
	for typ, prob := range unnormalized {
		normalized := 0.0
		if total > 0 {
			normalized = prob / total
		}
		probabilities = append(probabilities, TypeProbability{
			Type:        typ,
			Probability: normalized,
			Evidence:    likelihoods[typ],
		})
	}
	
	// Sort by probability (descending)
	sort.Slice(probabilities, func(i, j int) bool {
		return probabilities[i].Probability > probabilities[j].Probability
	})
	
	return probabilities
}

// selectType selects the most likely type based on posterior probabilities
func (e *BayesianInferenceEngine) selectType(posteriors []TypeProbability) (DataType, float64) {
	if len(posteriors) == 0 {
		return TypeUnknown, 0.0
	}
	
	best := posteriors[0]
	
	// Special handling: integers are also valid floats
	// If integer probability is high but float is also high, prefer int
	if best.Type == TypeFloat && len(posteriors) > 1 {
		for _, p := range posteriors[1:] {
			if p.Type == TypeInteger && p.Probability > 0.90*best.Probability {
				return TypeInteger, p.Probability
			}
		}
	}
	
	// If best type has very low confidence AND string has some evidence, fall back to string
	if best.Probability < e.config.ConfidenceThresh && best.Probability < 0.50 {
		// Check if string type exists in posteriors and has some probability
		for _, p := range posteriors {
			if p.Type == TypeString && p.Probability > 0.30 {
				return TypeString, p.Probability
			}
		}
	}
	
	return best.Type, best.Probability
}

// InferTypeSimple provides a simple interface matching the existing API
func (e *BayesianInferenceEngine) InferTypeSimple(values []string) string {
	result := e.InferType(values)
	return result.InferredType.String()
}

// SetLocale changes the locale for inference
func (e *BayesianInferenceEngine) SetLocale(locale Locale) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.Locale = locale
}

// SetSeed sets the random seed for deterministic behavior
func (e *BayesianInferenceEngine) SetSeed(seed int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.RandomSeed = seed
	source := rand.NewSource(seed)
	e.rng = rand.New(source)
}

// GetConfidence returns the confidence of the last inference
func (e *BayesianInferenceEngine) GetConfidence(values []string) float64 {
	result := e.InferType(values)
	return result.Confidence
}

// CalculateEntropy calculates Shannon entropy of type probabilities
func CalculateEntropy(probabilities []TypeProbability) float64 {
	entropy := 0.0
	for _, p := range probabilities {
		if p.Probability > 0 {
			entropy -= p.Probability * math.Log2(p.Probability)
		}
	}
	return entropy
}
