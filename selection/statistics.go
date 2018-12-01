package selection

import "math"

type BoxPlotStatistics struct {
	Min           float64
	Max           float64
	Median        float64
	Average       float64
	ThirdQuartile float64
	FirstQuartile float64
}

func GetBoxPlotStats(data []float64) BoxPlotStatistics {
	median := Median(data)
	avg := Average(data)
	q1 := Q1(data)
	q3 := Q3(data)
	return BoxPlotStatistics{Min(data), Max(data), median, avg, q3, q1}
}

func Average(floats []float64) float64 {
	var sum float64 = 0
	for _, v := range floats {
		sum += v
	}
	return sum / float64(len(floats))
}

func StdDev(floats []float64) float64 {
	difSquaredSum := 0.0
	avg := Average(floats)

	for _, v := range floats {
		difSquaredSum += math.Pow(v-avg, 2.0)
	}

	return math.Sqrt(difSquaredSum / float64(len(floats)))
}

func Median(floats []float64) float64 {
	elements := len(floats)
	median := 0.0
	if elements%2 == 0 {
		median = floats[(elements-1)/2] + (floats[(elements+1)/2]-floats[(elements-1)/2])/2.0
	} else {
		median = floats[elements/2]
	}

	return median
}

func Min(floats []float64) float64 {
	min := math.MaxFloat64
	for _, v := range floats {
		if min > v {
			min = v
		}
	}
	return min
}

func Max(floats []float64) float64 {
	max := math.Inf(-1)
	for _, v := range floats {
		if max < v {
			max = v
		}
	}
	return max
}

func Q1(floats []float64) float64 {
	medianIndex := len(floats) / 2
	return Median(floats[:medianIndex])
}

func Q3(floats []float64) float64 {
	medianIndex := len(floats) / 2
	return Median(floats[medianIndex:])
}

func Outliers(floats []float64) []float64 {
	outliers := make([]float64, 0)
	q1 := Q1(floats)
	q3 := Q3(floats)
	outerfence := (q3 - q1) * 4.5
	lowOuterFence := q1 - outerfence
	highOuterFence := q3 + outerfence

	for _, v := range floats {
		if v > highOuterFence || v < lowOuterFence {
			outliers = append(outliers, v)
		}
	}
	return outliers
}
