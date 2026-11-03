package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHumanAge(t *testing.T) {
	cases := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero is empty", 0, ""},
		{"negative is empty", -time.Second, ""},
		{"sub-minute uses seconds", 30 * time.Second, "30s"},
		{"sub-hour uses minutes", 5 * time.Minute, "5m"},
		{"sub-day uses hours", 3 * time.Hour, "3h"},
		{"day boundary", 48 * time.Hour, "2d"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, HumanAge(tc.d))
		})
	}
}
