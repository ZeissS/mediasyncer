package mediasyncer

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(int64(4714123 + time.Now().Nanosecond()))
}

// Returns a price to bid when storing the file. returns -1 if the file should not be stored.
type PriceFormula func(file FileID, stats FileStats, freeSpace ByteSize) Price

func PriceFormulaStatic(staticPrice Price) PriceFormula {
	return func(file FileID, stats FileStats, freeSpace ByteSize) Price {
		return staticPrice
	}
}

func PriceFormulaRandom() PriceFormula {
	return func(file FileID, stats FileStats, freeSpace ByteSize) Price {
		return Price(rand.Float32() * 2)
	}
}
