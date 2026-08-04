package trust

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildReport(t *testing.T) {
	expectedFactors := map[FactorName]Factor{
		PrivateContributionFactor:  Factor{Value: 2 * factorReferences[PrivateContributionFactor], TrustPercent: 0.99},
		IssueContributionFactor:    Factor{Value: 2 * factorReferences[IssueContributionFactor], TrustPercent: 0.99},
		CommitContributionFactor:   Factor{Value: 2 * factorReferences[CommitContributionFactor], TrustPercent: 0.99},
		RepoContributionFactor:     Factor{Value: 2 * factorReferences[RepoContributionFactor], TrustPercent: 0.99},
		PRContributionFactor:       Factor{Value: 2 * factorReferences[PRContributionFactor], TrustPercent: 0.99},
		PRReviewContributionFactor: Factor{Value: 2 * factorReferences[PRReviewContributionFactor], TrustPercent: 0.99},
		AccountAgeFactor:           Factor{Value: 2 * factorReferences[AccountAgeFactor], TrustPercent: 0.99},
		ContributionScoreFactor:    Factor{Value: 2 * factorReferences[ContributionScoreFactor], TrustPercent: 0.99},
	}

	trustData := map[FactorName][]float64{
		PrivateContributionFactor:  []float64{0, 4 * factorReferences[PrivateContributionFactor], 2 * factorReferences[PrivateContributionFactor]},
		IssueContributionFactor:    []float64{0, 2 * factorReferences[IssueContributionFactor], 4 * factorReferences[IssueContributionFactor]},
		CommitContributionFactor:   []float64{0, 2 * factorReferences[CommitContributionFactor], 4 * factorReferences[CommitContributionFactor]},
		RepoContributionFactor:     []float64{0, 2 * factorReferences[RepoContributionFactor], 4 * factorReferences[RepoContributionFactor]},
		PRContributionFactor:       []float64{0, 2 * factorReferences[PRContributionFactor], 4 * factorReferences[PRContributionFactor]},
		PRReviewContributionFactor: []float64{0, 2 * factorReferences[PRReviewContributionFactor], 4 * factorReferences[PRReviewContributionFactor]},
		AccountAgeFactor:           []float64{0, 2 * factorReferences[AccountAgeFactor], 4 * factorReferences[AccountAgeFactor]},
		ContributionScoreFactor:    []float64{0, 2 * factorReferences[ContributionScoreFactor], 4 * factorReferences[ContributionScoreFactor]},
	}

	report, err := buildReport(trustData)
	require.NoError(t, err)
	require.NotNil(t, report)

	for factor, expectedTrust := range expectedFactors {
		assert.Equal(t, expectedTrust, report.Factors[factor], "unexpected value for factor %q", factor)
	}
}

// TestBuildReportWithPercentiles feeds a known, evenly spread distribution so
// that each percentile has a distinct expected value. An all-zero sample would
// pass even if the percentiles were wired to the wrong slice or the wrong rank.
func TestBuildReportWithPercentiles(t *testing.T) {
	// Scores 1 to 100, so the Nth percentile lands on roughly the value N.
	scores := make([]float64, 0, 100)
	for i := 1; i <= 100; i++ {
		scores = append(scores, float64(i))
	}

	trustData := map[FactorName][]float64{ContributionScoreFactor: scores}
	for _, factor := range factors {
		if factor != ContributionScoreFactor {
			trustData[factor] = scores
		}
	}

	report, err := buildReport(trustData)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotNil(t, report.Percentiles)
	require.Len(t, report.Percentiles, len(percentiles))

	var previousValue float64
	for _, percentile := range percentiles {
		factor, ok := report.Percentiles[percentile]
		require.True(t, ok, "percentile %q is missing from the report", percentile)

		rank, err := strconv.ParseFloat(string(percentile), 64)
		require.NoError(t, err)

		assert.InDelta(t, rank, factor.Value, 2,
			"the %qth percentile of 1..100 should be close to %v", percentile, rank)
		assert.Greater(t, factor.Value, previousValue,
			"percentile values must increase with the percentile rank")
		assert.InDelta(t, computeTrustFromScore(factor.Value, percentileReferences[percentile]), factor.TrustPercent, 0.0001)

		previousValue = factor.Value
	}
}

// TestBuildReportOverallIsWeighted pins the weighted average down. Neither of
// the other report tests asserts the overall factor, so the weights themselves
// were never verified.
func TestBuildReportOverallIsWeighted(t *testing.T) {
	// Only the weighted contribution factor scores any trust at all.
	trustData := map[FactorName][]float64{
		ContributionScoreFactor: {2 * factorReferences[ContributionScoreFactor]},
	}
	for _, factor := range factors {
		if factor != ContributionScoreFactor {
			trustData[factor] = []float64{0}
		}
	}

	report, err := buildReport(trustData)
	require.NoError(t, err)

	var totalWeight int
	for _, weight := range factorWeights {
		totalWeight += weight
	}

	expected := 0.99 * float64(factorWeights[ContributionScoreFactor]) / float64(totalWeight)

	assert.InDelta(t, expected, report.Factors[Overall].TrustPercent, 0.0001)
	assert.Nil(t, report.Percentiles, "a single sample is too small to have percentiles")
}

// TestBuildReportWithMissingFactor guards against a factor with no computed
// value silently contributing a trust of zero to the overall score.
func TestBuildReportWithMissingFactor(t *testing.T) {
	trustData := map[FactorName][]float64{ContributionScoreFactor: {1}}

	_, err := buildReport(trustData)

	require.Error(t, err)
}

func TestComputeTrustFromScore(t *testing.T) {
	tests := map[string]struct {
		score     float64
		reference float64

		expectedTrust float64
	}{
		"no score":                {score: 0, reference: 100, expectedTrust: 0},
		"half the reference":      {score: 50, reference: 100, expectedTrust: 1.0 / 3.0},
		"exactly the reference":   {score: 100, reference: 100, expectedTrust: 2.0 / 3.0},
		"clamped above the cap":   {score: 1000, reference: 100, expectedTrust: 0.99},
		"exactly at the cap":      {score: 148.5, reference: 100, expectedTrust: 0.99},
		"zero reference":          {score: 42, reference: 0, expectedTrust: 0},
		"zero reference and zero": {score: 0, reference: 0, expectedTrust: 0},
		"negative reference":      {score: 42, reference: -10, expectedTrust: 0},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			t.Parallel()

			trust := computeTrustFromScore(test.score, test.reference)

			assert.False(t, math.IsNaN(trust), "trust must never be NaN")
			assert.False(t, math.IsInf(trust, 0), "trust must never be infinite")
			assert.InDelta(t, test.expectedTrust, trust, 0.0001)
		})
	}
}

func TestSplitTrustReports(t *testing.T) {
	trustData := make(map[FactorName][]float64)

	earlyUser := float64(1)
	randomUser := float64(2)

	trustData = addToTrustData(trustData, 200, earlyUser)
	trustData = addToTrustData(trustData, 800, randomUser)

	earlyUsers, randomUsers := splitTrustData(trustData)
	require.NotNil(t, earlyUsers)
	require.NotNil(t, randomUsers)

	for _, data := range earlyUsers {
		t.Run("early users data", func(t *testing.T) {
			t.Parallel()
			assert.Len(t, data, 200)
			for _, userValue := range data {
				assert.InDelta(t, userValue, earlyUser, 0.0001)
			}
		})
	}

	for _, data := range randomUsers {
		t.Run("random users data", func(t *testing.T) {
			t.Parallel()
			assert.Len(t, data, 800)
			for _, userValue := range data {
				assert.InDelta(t, userValue, randomUser, 0.0001)
			}
		})
	}
}

func addToTrustData(trustData map[FactorName][]float64, amount int, value float64) map[FactorName][]float64 {
	for i := 0; i < amount; i++ {
		for _, factor := range factors {
			trustData[factor] = append(trustData[factor], value)
		}
	}

	return trustData
}

// TestSplitTrustDataSmallerThanFirstBlock covers samples below the 200 user
// first block, which used to index past the end of every factor slice.
func TestSplitTrustDataSmallerThanFirstBlock(t *testing.T) {
	for _, total := range []int{0, 1, 50, 199, 200, 201} {
		t.Run(fmt.Sprint(total), func(t *testing.T) {
			t.Parallel()

			trustData := addToTrustData(make(map[FactorName][]float64), total, 1)

			var early, random map[FactorName][]float64
			require.NotPanics(t, func() {
				early, random = splitTrustData(trustData)
			})

			for _, factor := range factors {
				assert.Len(t, early[factor], min(total, firstStargazersAmount), "factor %q", factor)
				assert.Len(t, random[factor], max(total-firstStargazersAmount, 0), "factor %q", factor)
			}
		})
	}
}

// TestSplitTrustDataWithRaggedFactors covers factor slices of differing
// lengths, which used to be indexed with a length taken from another factor.
func TestSplitTrustDataWithRaggedFactors(t *testing.T) {
	trustData := addToTrustData(make(map[FactorName][]float64), 300, 1)
	trustData[AccountAgeFactor] = trustData[AccountAgeFactor][:10]

	require.NotPanics(t, func() {
		early, random := splitTrustData(trustData)

		assert.Len(t, early[AccountAgeFactor], 10)
		assert.Empty(t, random[AccountAgeFactor])
		assert.Len(t, early[ContributionScoreFactor], firstStargazersAmount)
		assert.Len(t, random[ContributionScoreFactor], 100)
	})
}

// TestBuildComparativeReportKeepsZeroPercentiles guards against a legitimately
// zero percentile being mistaken for a percentile that was never computed.
// Repositories with fake stars are exactly the ones whose low percentiles are
// zero, and dropping the percentile block raised their overall score.
func TestBuildComparativeReportKeepsZeroPercentiles(t *testing.T) {
	trustData := make(map[FactorName][]float64)
	// 200 early stargazers that score well.
	for _, factor := range factors {
		for i := 0; i < firstStargazersAmount; i++ {
			trustData[factor] = append(trustData[factor], 2*factorReferences[factor])
		}
	}
	// 100 later stargazers with no activity at all.
	trustData = addToTrustData(trustData, 100, 0)

	report, err := buildComparativeReport(trustData)
	require.NoError(t, err)
	require.NotNil(t, report)

	require.NotNil(t, report.Percentiles, "percentiles of exactly zero must be kept")
	require.Len(t, report.Percentiles, len(percentiles))

	for _, percentile := range percentiles {
		assert.Zero(t, report.Percentiles[percentile].TrustPercent,
			"the worse of both samples has no contributions at all")
	}

	// With every factor and every percentile at zero, so is the overall trust.
	assert.Zero(t, report.Factors[Overall].TrustPercent)
}

// TestBuildComparativeReportWithoutPercentiles covers the sample being too
// small for percentiles, where a nil map is the correct outcome.
func TestBuildComparativeReportWithoutPercentiles(t *testing.T) {
	trustData := addToTrustData(make(map[FactorName][]float64), firstStargazersAmount+5, 1)

	report, err := buildComparativeReport(trustData)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Nil(t, report.Percentiles, "5 samples are too few to have percentiles")
}

func TestComputeWithoutUsers(t *testing.T) {
	_, err := Compute(context.Background(), nil, nil)

	require.Error(t, err, "an empty scan must fail instead of reporting a score of zero")
}
