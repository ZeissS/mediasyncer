package libsyncer

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(int64(4714123 + time.Now().Nanosecond()))
}

// Clock defines a function that tells us the current time.
type Clock func() time.Time

// PriceFormula calculates a price to bid when storing the file. Returns -1 if the file should not be stored.
type PriceFormula func(file FileID, stats FileStats, freeSpace ByteSize) Price

// PriceFormulaStatic always returns the staticPrice when calculating a price.
func PriceFormulaStatic(staticPrice Price) PriceFormula {
	return func(file FileID, stats FileStats, freeSpace ByteSize) Price {
		return staticPrice
	}
}

// PriceFormulaRandom returns a random price between 0 and 2.
func PriceFormulaRandom() PriceFormula {
	return func(file FileID, stats FileStats, freeSpace ByteSize) Price {
		return Price(rand.Float32() * 2)
	}
}

// PriceFormulaAge returns a PriceFormula that bids with agePrice if the file is
// older (if preferOlder is true) than age. If preferOlder is false, the file must
// be younger than age, to return agePrice. Otherwise defaultPrice is returned.
// If no ModTime is given, -1 is returned.
//
// This can be used to bid high on old files (e.g. for a rarely online storage)
// or to keep new files on the server that downloaded them.
//
// Pass time.Now as the clock to use the current system time for this function.
func PriceFormulaAge(preferOlder bool, age time.Duration, agePrice, defaultPrice Price, clock Clock) PriceFormula {
	return func(file FileID, stats FileStats, freeSpace ByteSize) Price {
		if stats.ModTime == nil {
			return -1
		}

		timeBoundary := clock().Add(-1*age)
		if preferOlder && stats.ModTime.Before(timeBoundary) {
			return agePrice
		}

		if !preferOlder && stats.ModTime.After(timeBoundary) {
			return agePrice
		}
		return defaultPrice
	}
}
