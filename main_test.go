package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekdayNumber17(t *testing.T) {
	date := time.Date(2022, 8, 17, 0, 0, 0, 0, time.Now().Location())
	assert.Equal(t, 3, weekdayNumber(date), "Wrong weekday of month number")
}

func TestWeekdayNumber1(t *testing.T) {
	date := time.Date(2022, 8, 1, 0, 0, 0, 0, time.Now().Location())
	assert.Equal(t, 1, weekdayNumber(date), "Wrong weekday of month number")
}

func TestWeekdayNumber7(t *testing.T) {
	date := time.Date(2022, 8, 7, 0, 0, 0, 0, time.Now().Location())
	assert.Equal(t, 1, weekdayNumber(date), "Wrong weekday of month number")
}

func TestWeekdayNumber8(t *testing.T) {
	date := time.Date(2022, 8, 8, 0, 0, 0, 0, time.Now().Location())
	assert.Equal(t, 2, weekdayNumber(date), "Wrong weekday of month number")
}

func TestWeekdayNumber31(t *testing.T) {
	date := time.Date(2022, 8, 31, 0, 0, 0, 0, time.Now().Location())
	assert.Equal(t, 5, weekdayNumber(date), "Wrong weekday of month number")
}
