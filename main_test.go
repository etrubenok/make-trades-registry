package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPreviousDate(t *testing.T) {
	current := time.Date(int(2019), time.Month(1), int(10), int(0), int(0), int(0), int(0), time.UTC)

	year, month, day := GetPreviousDate(current)
	assert.Equal(t, 2019, year)
	assert.Equal(t, 1, month)
	assert.Equal(t, 9, day)
}

func TestGetPreviousDatePrev(t *testing.T) {
	current := time.Date(int(2019), time.Month(1), int(10), int(0), int(23), int(1), int(0), time.UTC)

	year, month, day := GetPreviousDate(current)
	assert.Equal(t, 2019, year)
	assert.Equal(t, 1, month)
	assert.Equal(t, 9, day)
}

func TestGetPreviousDatePrevMonth(t *testing.T) {
	current := time.Date(int(2019), time.Month(3), int(1), int(0), int(23), int(1), int(0), time.UTC)

	year, month, day := GetPreviousDate(current)
	assert.Equal(t, 2019, year)
	assert.Equal(t, 2, month)
	assert.Equal(t, 28, day)
}

func TestGetPreviousDatePrevYear(t *testing.T) {
	current := time.Date(int(2020), time.Month(1), int(1), int(0), int(23), int(1), int(0), time.UTC)

	year, month, day := GetPreviousDate(current)
	assert.Equal(t, 2019, year)
	assert.Equal(t, 12, month)
	assert.Equal(t, 31, day)
}
