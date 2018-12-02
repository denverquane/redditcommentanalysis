package selection

import (
	"fmt"
	"testing"
)

func TestAverage(t *testing.T) {
	data := []float64{0.0, 2.0, 3.0, 4.0, 6.0}
	if Average(data) != 3.0 {
		t.Fail()
	}
}

func TestMedian(t *testing.T) {
	data := []float64{0.0, 2.0, 3.0, 4.0, 20.0}
	if Median(data) != 3.0 {
		t.Fail()
	}
	data2 := []float64{0.0, 2.0, 3.0, 3.5, 4.0, 20.0}
	if Median(data2) != 3.25 {
		t.Fail()
	}
}

func TestStdDev(t *testing.T) {
	data := []float64{9, 2, 5, 4, 12, 7, 8, 11, 9, 3, 7, 4, 12, 5, 4, 10, 9, 6, 9, 4}

	if StdDev(data) != 2.9832867780352594 {
		t.Fail()
	}
}

func TestMax(t *testing.T) {
	data := []float64{-1000.0, 2.0, 3000.0, 4.0, 20.0}
	if Max(data) != 3000.0 {
		t.Fail()
	}
}

func TestMin(t *testing.T) {
	data := []float64{-1000.0, 2.0, 3000.0, 4.0, 20.0}
	if Min(data) != -1000.0 {
		t.Fail()
	}
}

func TestGetPolarityAndSubjectivityForString(t *testing.T) {
	fmt.Println(GetPolarityAndSubjectivityForString("http://192.168.1.192:8889/api", "you don't want mexico at least then you would get some decent foodd"))
}
