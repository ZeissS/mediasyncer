package libsyncer

import (
	"testing"
	"time"
)

func staticClock(t time.Time) Clock {
	return func() time.Time {
		return t
	}
}

func TestPriceFormulaAge_PreferYounger(t *testing.T) {
	agePrice := Price(1.0)
	age := time.Duration(60 * 24 * time.Hour)
	defaultPrice := Price(0.5)
	now := time.Date(2016, 01, 01, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	f := PriceFormulaAge(false, age, agePrice, defaultPrice, staticClock(now))

	file := FileID{
		VolumeID: "vol1",
		Path:     "/testing.txt",
	}
	modTime := time.Date(2015, 12, 31, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	stats := FileStats{
		ModTime: &modTime,
	}
	freeSpace := ByteSize(1024)

	price := f(file, stats, freeSpace)
	if price != agePrice {
		t.Fatalf("Expected agePrice, but got defaultPrice for young file.")
	}
}
func TestPriceFormulaAge_PreferYounger_TooOld(t *testing.T) {
	agePrice := Price(1.0)
	age := time.Duration(60 * 24 * time.Hour)
	defaultPrice := Price(0.5)
	now := time.Date(2016, 01, 01, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	f := PriceFormulaAge(false, age, agePrice, defaultPrice, staticClock(now))

	file := FileID{
		VolumeID: "vol1",
		Path:     "/testing.txt",
	}
	modTime := time.Date(2015, 03, 10, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	stats := FileStats{
		ModTime: &modTime,
	}
	freeSpace := ByteSize(1024)

	price := f(file, stats, freeSpace)
	if price != defaultPrice {
		t.Fatalf("Expected defaultPrice, but got agePrice for too young file.")
	}
}



func TestPriceFormulaAge_PreferOlder_TooYoung(t *testing.T) {
	agePrice := Price(1.0)
	age := time.Duration(60 * 24 * time.Hour)
	defaultPrice := Price(0.5)
	now := time.Date(2016, 01, 01, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	f := PriceFormulaAge(true, age, agePrice, defaultPrice, staticClock(now))

	file := FileID{
		VolumeID: "vol1",
		Path:     "/testing.txt",
	}
	modTime := time.Date(2015, 12, 31, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	stats := FileStats{
		ModTime: &modTime,
	}
	freeSpace := ByteSize(1024)

	price := f(file, stats, freeSpace)
  if price != defaultPrice {
		t.Fatalf("Expected defaultPrice, but got agePrice for old file.")
	}
}
func TestPriceFormulaAge_PreferOlder(t *testing.T) {
	agePrice := Price(1.0)
	age := time.Duration(60 * 24 * time.Hour)
	defaultPrice := Price(0.5)
	now := time.Date(2016, 01, 01, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	f := PriceFormulaAge(true, age, agePrice, defaultPrice, staticClock(now))

	file := FileID{
		VolumeID: "vol1",
		Path:     "/testing.txt",
	}
	modTime := time.Date(2015, 03, 10, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	stats := FileStats{
		ModTime: &modTime,
	}
	freeSpace := ByteSize(1024)

	price := f(file, stats, freeSpace)
  if price != agePrice {
		t.Fatalf("Expected agePrice, but got defaultPrice for old file.")
	}

}
