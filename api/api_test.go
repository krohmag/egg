package api

import (
	"egg/datastore"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateEB(t *testing.T) {
	tests := []struct {
		name         string
		user         datastore.User
		calculatedEB string
	}{
		{
			name: "only SE",
			user: datastore.User{
				SoulFood:      0,
				ProphecyBonus: 0,
				SoulEggs:      1000000000000000000,
				ProphecyEggs:  0,
			},
			calculatedEB: "10.000Q",
		},
		{
			name: "SE with some SF",
			user: datastore.User{
				SoulFood:      70,
				ProphecyBonus: 0,
				SoulEggs:      1000000000000000000,
				ProphecyEggs:  0,
			},
			calculatedEB: "80.000Q",
		},
		{
			name: "SE with PE, no ER",
			user: datastore.User{
				SoulFood:      0,
				ProphecyBonus: 0,
				SoulEggs:      1000000000000000000,
				ProphecyEggs:  1,
			},
			calculatedEB: "10.500Q",
		},
		{
			name: "SE with PE and PB",
			user: datastore.User{
				SoulFood:      0,
				ProphecyBonus: 1,
				SoulEggs:      1000000000000000000,
				ProphecyEggs:  1,
			},
			calculatedEB: "10.600Q",
		},
		{
			name: "SE with PE and both ER",
			user: datastore.User{
				SoulFood:      70,
				ProphecyBonus: 1,
				SoulEggs:      1000000000000000000,
				ProphecyEggs:  1,
			},
			calculatedEB: "84.800Q",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.calculatedEB, calculateEB(test.user))
		})
	}
}

func TestForPeople(t *testing.T) {
	tests := []struct {
		name                 string
		bigAssNumber         float64
		humanReadableVersion string
	}{
		{
			name:                 "sub-kilo",
			bigAssNumber:         1,
			humanReadableVersion: "1.000",
		},
		{
			name:                 "kilo",
			bigAssNumber:         1000,
			humanReadableVersion: "1.000k",
		},
		{
			name:                 "mega",
			bigAssNumber:         1000000,
			humanReadableVersion: "1.000m",
		},
		{
			name:                 "giga",
			bigAssNumber:         1000000000,
			humanReadableVersion: "1.000b",
		},
		{
			name:                 "tera",
			bigAssNumber:         1000000000000,
			humanReadableVersion: "1.000T",
		},
		{
			name:                 "peta",
			bigAssNumber:         1000000000000000,
			humanReadableVersion: "1.000q",
		},
		{
			name:                 "exa",
			bigAssNumber:         1000000000000000000,
			humanReadableVersion: "1.000Q",
		},
		{
			name:                 "zetta",
			bigAssNumber:         1000000000000000000000,
			humanReadableVersion: "1.000s",
		},
		{
			name:                 "yotta",
			bigAssNumber:         1000000000000000000000000,
			humanReadableVersion: "1.000S",
		},
		{
			name:                 "xenna",
			bigAssNumber:         1000000000000000000000000000,
			humanReadableVersion: "1.000o",
		},
		{
			name:                 "wecca",
			bigAssNumber:         1000000000000000000000000000000,
			humanReadableVersion: "1.000N",
		},
		{
			name:                 "venda",
			bigAssNumber:         1000000000000000000000000000000000,
			humanReadableVersion: "1.000d",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.humanReadableVersion, forPeople(test.bigAssNumber))
		})
	}
}
