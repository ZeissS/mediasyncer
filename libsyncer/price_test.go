package libsyncer

import (
  "time"
  "testing"
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
    Path: "/testing.txt",
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
