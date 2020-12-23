package plugins

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/utils"
)

func TestDateFloorCeil(t *testing.T) {
	timeLayout := time.RFC3339 // 2006-01-02T15:04:05Z
	testCases := []struct {
		input     string
		wantFloor string
		wantCeil  string
	}{
		{
			input:     "2006-01-02T15:04:05Z",
			wantFloor: "2006-01-02T00:00:00Z",
			wantCeil:  "2006-01-02T23:59:59Z",
		}, {
			input:     "2006-01-02T15:04:05+07:00",
			wantFloor: "2006-01-02T00:00:00+07:00",
			wantCeil:  "2006-01-02T23:59:59+07:00",
		}, {
			input:     "2020-02-28T10:34:59-04:30",
			wantFloor: "2020-02-28T00:00:00-04:30",
			wantCeil:  "2020-02-28T23:59:59-04:30",
		}, {
			input:     "2020-02-29T10:34:59-04:30",
			wantFloor: "2020-02-29T00:00:00-04:30",
			wantCeil:  "2020-02-29T23:59:59-04:30",
		},
	}

	for _, k := range testCases {
		t.Run(k.input, func(t *testing.T) {
			input, e := time.Parse(timeLayout, k.input)
			if !assert.NoError(t, e) {
				return
			}

			floor := floorDate(input).Format(timeLayout)
			assert.Equal(t, k.wantFloor, floor, "floor")

			ceil := ceilDate(input).Format(timeLayout)
			assert.Equal(t, k.wantCeil, ceil, "ceil")
		})
	}
}

func makeTestSellTwapLevelProvider(seed int64) *sellTwapLevelProvider {
	return makeTestSellTwapLevelProvider2(
		seed,
		2,
		60,
		0.2,
	)
}

func makeTestSellTwapLevelProvider2(
	seed int64,
	numHoursToSell int,
	parentBucketSizeSeconds int,
	minChildOrderSizePercentOfParent float64,
) *sellTwapLevelProvider {
	startPf, _ := newFixedFeed("10.0")
	offset := rateOffset{
		percent:      0.0,
		absolute:     0.0,
		percentFirst: true,
	}
	p, e := makeSellTwapLevelProvider(
		startPf,
		offset,
		model.MakeOrderConstraints(7, 7, 0.1),
		[7]volumeFilter{
			volumeFilter{configValue: "/sell/base/"},
			volumeFilter{configValue: "/sell/base/"},
			volumeFilter{configValue: "/sell/base/"},
			volumeFilter{configValue: "/sell/base/"},
			volumeFilter{configValue: "/sell/base/"},
			volumeFilter{configValue: "/sell/base/"},
			volumeFilter{configValue: "/sell/base/"}},
		numHoursToSell,
		parentBucketSizeSeconds,
		0.05,
		0.5,
		minChildOrderSizePercentOfParent,
		seed,
		false,
	)
	if e != nil {
		panic(e)
	}
	return p.(*sellTwapLevelProvider)
}

func TestMakeFirstBucketFrame(t *testing.T) {
	testCases := []struct {
		name                      string
		bucketID                  int64
		roundID                   int64
		dayBaseCapacity           float64
		dailyVolumeBase           float64
		wantDayBaseCapacity       float64
		wantDayBaseRemaining      float64
		wantDayBaseSoldStart      float64
		wantTotalBaseSurplusStart float64
		wantBaseSurplusIncluded   float64
		wantBaseCapacity          float64
		wantMinOrderSizeBase      float64
	}{
		/*
			the following test cases cover combinations of the following:
			- bucketID
				- bucketID = 0
				- 0 < bucketID < lastBucketID
				- bucketID = lastBucketID
				- bucketID = lastBucketID + 1
				- bucketID = lastBucketID + 2
			- expectedSold
				- less than expected (positive surplus)
				- exactly expected (no surplus)
				- more than expected (negative surplus)
				- full day capacity sold (negative surplus)
		*/
		{
			name:                      "bucket0 no surplus",
			bucketID:                  0,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           0.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      1000.0,
			wantDayBaseSoldStart:      0.0,
			wantTotalBaseSurplusStart: 0.0,
			wantBaseSurplusIncluded:   0.0,
			wantBaseCapacity:          8.333333333333334,
			wantMinOrderSizeBase:      1.666666666666667,
		}, {
			name:                      "bucket0 roundID has no effect",
			bucketID:                  0,
			roundID:                   1,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           0.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      1000.0,
			wantDayBaseSoldStart:      0.0,
			wantTotalBaseSurplusStart: 0.0,
			wantBaseSurplusIncluded:   0.0,
			wantBaseCapacity:          8.333333333333334,
			wantMinOrderSizeBase:      1.666666666666667,
		}, {
			name:                      "bucket0 starts off with positive amount sold, i.e. negative surplus",
			bucketID:                  0,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           5.0, // there is already an amount sold when starting off this bucket, which should be distributed
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      995.0,
			wantDayBaseSoldStart:      5.0,
			wantTotalBaseSurplusStart: -5.0,
			wantBaseSurplusIncluded:   -2.5396825397, // -5.0*(.5-1)/((.5^ceil(.05*(120-0)))-1)
			wantBaseCapacity:          5.7936507936,  // 8.333333333333334 - 2.5396825397
			wantMinOrderSizeBase:      1.1587301587,  // 0.2 * 5.7936507936
		}, {
			name:                      "non-zero bucket within selling hours, exact amount sold as expected, i.e. no surplus",
			bucketID:                  5,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           41.66666666666667,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      958.3333333333334,
			wantDayBaseSoldStart:      41.66666666666667,
			wantTotalBaseSurplusStart: 0.0,               // there is no surplus becuase we have sold the exact amount expected
			wantBaseSurplusIncluded:   0.0,               // therefore, no surplus to be included
			wantBaseCapacity:          8.333333333333334, // therefore this is the average selling amount
			wantMinOrderSizeBase:      1.666666666666667, // 0.2 * 8.333333333333334
		}, {
			name:                      "non-zero bucket within selling hours, nothing sold, i.e. everything as a surplus",
			bucketID:                  5,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           0.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      1000.0,
			wantDayBaseSoldStart:      0.0,
			wantTotalBaseSurplusStart: 41.66666666666667,  // 5 * 8.333333333333334 - 0.0
			wantBaseSurplusIncluded:   21.164021164021168, // 41.66666666666667*(.5-1)/((.5^ceil(.05*(120-5)))-1)
			wantBaseCapacity:          29.4973544973545,   //  8.333333333333334 + 21.164021164021168
			wantMinOrderSizeBase:      5.8994708994709,    // 0.2 * 29.4973544973545
		}, {
			name:                      "non-zero bucket within selling hours, partial amount sold, i.e. partial surplus",
			bucketID:                  5,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           40.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      960.0,
			wantDayBaseSoldStart:      40.0,         // always equal to dailyVolumeBase
			wantTotalBaseSurplusStart: 1.6666666667, // 5 * 8.333333333333334 - 40.0
			wantBaseSurplusIncluded:   0.8465608466, // 1.6666666667*(.5-1)/((.5^ceil(.05*(120-5)))-1)
			wantBaseCapacity:          9.1798941799, // 8.333333333333334 + 0.8465608466
			wantMinOrderSizeBase:      1.835978836,  // 0.2 * 9.1798941799
		}, {
			name:                      "non-zero bucket within selling hours, extra amount sold, i.e. negative surplus",
			bucketID:                  5,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           45.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      955.0,
			wantDayBaseSoldStart:      45.0,          // always equal to dailyVolumeBase
			wantTotalBaseSurplusStart: -3.3333333333, // 5 * 8.333333333333334 - 45.0
			wantBaseSurplusIncluded:   -1.6931216931, // -3.3333333333*(.5-1)/((.5^ceil(.05*(120-5)))-1)
			wantBaseCapacity:          6.6402116402,  // 8.333333333333334 - 1.6931216931
			wantMinOrderSizeBase:      1.328042328,   // 0.2 * 6.6402116402
		}, {
			name:                      "last bucket within selling hours, nothing sold, i.e. everything as a surplus included in current bucket",
			bucketID:                  119,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           0.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      1000.0,
			wantDayBaseSoldStart:      0.0,
			wantTotalBaseSurplusStart: 991.6666666666667, // 119 * 8.333333333333334 - 0.0
			wantBaseSurplusIncluded:   991.6666666666667, // 991.6666666666667*(.5-1)/((.5^ceil(.05*(120-119)))-1)
			wantBaseCapacity:          1000.0,            // 8.333333333333334 + 991.6666666666667
			wantMinOrderSizeBase:      200.0,             // 0.2 * 1000.0
		}, {
			name:                      "last bucket within selling hours, partial sold, i.e. everything as a surplus included in current bucket",
			bucketID:                  119,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           5.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      995.0,
			wantDayBaseSoldStart:      5.0,
			wantTotalBaseSurplusStart: 986.6666666666667, // 119 * 8.333333333333334 - 5.0
			wantBaseSurplusIncluded:   986.6666666666667, // 986.6666666666667*(.5-1)/((.5^ceil(.05*(120-119)))-1)
			wantBaseCapacity:          995.0,             // 8.333333333333334 + 986.6666666666667
			wantMinOrderSizeBase:      199.0,             // 0.2 * 995.0
		}, {
			name:                      "last bucket within selling hours, exact sold as expected, i.e. no surplus",
			bucketID:                  119,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           991.6666666667,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      8.3333333333,
			wantDayBaseSoldStart:      991.6666666667,
			wantTotalBaseSurplusStart: 0.0,               // 119 * 8.333333333333334 - 991.6666666667
			wantBaseSurplusIncluded:   0.0,               // 0.0*(.5-1)/((.5^ceil(.05*(120-119)))-1)
			wantBaseCapacity:          8.333333333333334, // 8.333333333333334 + 0.0
			wantMinOrderSizeBase:      1.6666666667,      // 0.2 * 8.333333333333334
		}, {
			name:                      "last bucket within selling hours, a little extra sold but less than everything, i.e. negative surplus",
			bucketID:                  119,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           995.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      5.0,
			wantDayBaseSoldStart:      995.0,
			wantTotalBaseSurplusStart: -3.3333333333, // 119 * 8.333333333333334 - 995.0
			wantBaseSurplusIncluded:   -3.3333333333, // -3.3333333333*(.5-1)/((.5^ceil(.05*(120-119)))-1)
			wantBaseCapacity:          5.0,           // 8.333333333333334 - 3.3333333333
			wantMinOrderSizeBase:      1.0,           // 0.2 * 5.0
		}, {
			name:                      "last bucket within selling hours, everything sold, i.e. negative surplus",
			bucketID:                  119,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           1000.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      0.0,
			wantDayBaseSoldStart:      1000.0,
			wantTotalBaseSurplusStart: -8.333333333333334, // 119 * 8.333333333333334 - 1000.0
			wantBaseSurplusIncluded:   -8.333333333333334, // -8.333333333333334*(.5-1)/((.5^ceil(.05*(120-119)))-1)
			wantBaseCapacity:          0.0,                // 8.333333333333334 - 8.333333333333334
			wantMinOrderSizeBase:      0.0,                // 0.2 * 0.0
		}, {
			name:                      "first bucket outside selling hours, nothing sold, i.e. everything as a surplus included in current bucket",
			bucketID:                  120,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           0.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      1000.0,
			wantDayBaseSoldStart:      0.0,
			wantTotalBaseSurplusStart: 1000.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   1000.0, // we have special logic to do this assignment equal to the full surplus when outside selling hours
			wantBaseCapacity:          1000.0, // no allotted capacity so this is only the suruplus
			wantMinOrderSizeBase:      200.0,  // 0.2 * 1000.0
		}, {
			name:                      "second bucket outside selling hours, nothing sold, i.e. everything as a surplus included in current bucket",
			bucketID:                  121,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           0.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      1000.0,
			wantDayBaseSoldStart:      0.0,
			wantTotalBaseSurplusStart: 1000.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   1000.0, // we have special logic to do this assignment equal to the full surplus when outside selling hours
			wantBaseCapacity:          1000.0, // no allotted capacity so this is only the suruplus
			wantMinOrderSizeBase:      200.0,  // 0.2 * 1000.0
		}, {
			name:                      "first bucket outside selling hours, everything sold, i.e. no surplus",
			bucketID:                  120,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           1000.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      0.0,
			wantDayBaseSoldStart:      1000.0,
			wantTotalBaseSurplusStart: 0.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   0.0,
			wantBaseCapacity:          0.0,
			wantMinOrderSizeBase:      0.0,
		}, {
			name:                      "second bucket outside selling hours, everything sold, i.e. no surplus",
			bucketID:                  121,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           1000.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      0.0,
			wantDayBaseSoldStart:      1000.0,
			wantTotalBaseSurplusStart: 0.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   0.0,
			wantBaseCapacity:          0.0,
			wantMinOrderSizeBase:      0.0,
		}, {
			name:                      "first bucket outside selling hours, partial sold, i.e. partial surplus all of which included in current bucket",
			bucketID:                  120,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           200.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      800.0,
			wantDayBaseSoldStart:      200.0,
			wantTotalBaseSurplusStart: 800.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   800.0, // we have special logic to do this assignment equal to the full surplus when outside selling hours
			wantBaseCapacity:          800.0, // no allotted capacity so this is only the suruplus
			wantMinOrderSizeBase:      160.0, // 0.2 * 800.0
		}, {
			name:                      "second bucket outside selling hours, partial sold, i.e. partial surplus all of which included in current bucket",
			bucketID:                  121,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           200.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      800.0,
			wantDayBaseSoldStart:      200.0,
			wantTotalBaseSurplusStart: 800.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   800.0, // we have special logic to do this assignment equal to the full surplus when outside selling hours
			wantBaseCapacity:          800.0, // no allotted capacity so this is only the suruplus
			wantMinOrderSizeBase:      160.0, // 0.2 * 800.0
		}, {
			name:                      "first bucket outside selling hours, extra amount sold, i.e. negative surplus",
			bucketID:                  120,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           1001.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      -1.0,
			wantDayBaseSoldStart:      1001.0,
			wantTotalBaseSurplusStart: -1.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   -1.0, // always include full surplus start outside selling hours,
			wantBaseCapacity:          -1.0, // 0 - 1.0	(we let this be negative and do not have extra logic to "fix this up" in the bucket math)
			wantMinOrderSizeBase:      -0.2, // 0.2 * -1.0 (we don't "fix up" baseCapacity in the bucket math)
		}, {
			name:                      "second bucket outside selling hours, extra amount sold, i.e. negative surplus",
			bucketID:                  121,
			roundID:                   0,
			dayBaseCapacity:           1000.0,
			dailyVolumeBase:           1001.0,
			wantDayBaseCapacity:       1000.0,
			wantDayBaseRemaining:      -1.0,
			wantDayBaseSoldStart:      1001.0,
			wantTotalBaseSurplusStart: -1.0, // always equal to dayBaseRemaining outside selling hours
			wantBaseSurplusIncluded:   -1.0, // always include full surplus start outside selling hours,
			wantBaseCapacity:          -1.0, // 0 - 1.0	(we let this be negative and do not have extra logic to "fix this up" in the bucket math)
			wantMinOrderSizeBase:      -0.2, // 0.2 * -1.0 (we don't "fix up" baseCapacity in the bucket math)
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
			startDate := now.Add(time.Minute * -5)
			endDate := now.Add(time.Minute * 5)
			p := makeTestSellTwapLevelProvider(0)
			bucketInfo, e := p.makeFirstBucketFrame(
				now,
				startDate,
				endDate,
				bucketID(k.bucketID),
				roundID(k.roundID),
				k.dayBaseCapacity,
				&queries.DailyVolume{
					BaseVol:  k.dailyVolumeBase,
					QuoteVol: 0.0, // this value does not matter
				},
			)
			if !assert.NoError(t, e) {
				return
			}

			assert.Equal(t, bucketID(k.bucketID), bucketInfo.ID)
			utils.AssetFloatEquals(t, k.wantBaseCapacity, bucketInfo.baseCapacity)
			utils.AssetFloatEquals(t, 0.0, bucketInfo.dynamicValues.baseSold)         // this is always 0.0
			utils.AssetFloatEquals(t, k.wantBaseCapacity, bucketInfo.baseRemaining()) // this is always equal to baseCapacity because nothing has been sold in this bucket yet
			utils.AssetFloatEquals(t, k.wantBaseSurplusIncluded, bucketInfo.baseSurplusIncluded)
			utils.AssetFloatEquals(t, k.wantDayBaseCapacity, bucketInfo.dayBaseCapacity)
			utils.AssetFloatEquals(t, k.wantDayBaseRemaining, bucketInfo.dayBaseRemaining())
			utils.AssetFloatEquals(t, k.wantDayBaseSoldStart, bucketInfo.dayBaseSoldStart)
			utils.AssetFloatEquals(t, k.wantDayBaseSoldStart, bucketInfo.dynamicValues.dayBaseSold) // this is always equal to dayBaseSoldStart because the bucket has not sold anything
			assert.Equal(t, true, bucketInfo.dynamicValues.isNew)
			assert.Equal(t, false, bucketInfo.dynamicValues.isLast)
			assert.Equal(t, now, bucketInfo.dynamicValues.now)
			assert.Equal(t, roundID(k.roundID), bucketInfo.dynamicValues.roundID)
			assert.Equal(t, endDate, bucketInfo.endTime)
			utils.AssetFloatEquals(t, k.wantMinOrderSizeBase, bucketInfo.minOrderSizeBase)
			assert.Equal(t, 60, bucketInfo.sizeSeconds)
			assert.Equal(t, startDate, bucketInfo.startTime)
			utils.AssetFloatEquals(t, k.wantTotalBaseSurplusStart, bucketInfo.totalBaseSurplusStart)
			assert.Equal(t, int64(1440), bucketInfo.totalBuckets)
			assert.Equal(t, int64(120), bucketInfo.totalBucketsToSell)
		})
	}
}

func TestUpdateExistingBucket(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startDate := now.Add(time.Minute * -5)
	endDate := now.Add(time.Minute * 5)
	p := makeTestSellTwapLevelProvider(0)
	bucketInfo, e := p.makeFirstBucketFrame(
		now,
		startDate,
		endDate,
		bucketID(0),
		roundID(0),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  0.0,
			QuoteVol: 0.0,
		},
	)
	if e != nil {
		panic(e)
	}

	p.activeBucket = bucketInfo
	now2 := now.Add(time.Second * 30)
	updatedBucketInfo, e := p.updateExistingBucket(
		now2,
		&queries.DailyVolume{
			BaseVol:  5.0,
			QuoteVol: 1.0,
		},
		roundID(3),
	)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, bucketID(0), updatedBucketInfo.ID)
	assert.Equal(t, 8.333333333333334, updatedBucketInfo.baseCapacity)
	assert.Equal(t, 3.333333333333334, updatedBucketInfo.baseRemaining())
	assert.Equal(t, 0.0, updatedBucketInfo.baseSurplusIncluded)
	assert.Equal(t, 1000.0, updatedBucketInfo.dayBaseCapacity)
	assert.Equal(t, 995.0, updatedBucketInfo.dayBaseRemaining())
	assert.Equal(t, 0.0, updatedBucketInfo.dayBaseSoldStart)
	assert.Equal(t, 5.0, updatedBucketInfo.dynamicValues.baseSold)
	assert.Equal(t, 5.0, updatedBucketInfo.dynamicValues.dayBaseSold)
	assert.Equal(t, false, updatedBucketInfo.dynamicValues.isNew)
	assert.Equal(t, false, updatedBucketInfo.dynamicValues.isLast)
	assert.Equal(t, now2, updatedBucketInfo.dynamicValues.now)
	assert.Equal(t, roundID(3), updatedBucketInfo.dynamicValues.roundID)
	assert.Equal(t, endDate, updatedBucketInfo.endTime)
	assert.Equal(t, 1.666666666666667, updatedBucketInfo.minOrderSizeBase)
	assert.Equal(t, 60, updatedBucketInfo.sizeSeconds)
	assert.Equal(t, startDate, updatedBucketInfo.startTime)
	assert.Equal(t, 0.0, updatedBucketInfo.totalBaseSurplusStart)
	assert.Equal(t, int64(1440), updatedBucketInfo.totalBuckets)
	assert.Equal(t, int64(120), updatedBucketInfo.totalBucketsToSell)
}

func TestFirstDistributionOfBaseSurplus(t *testing.T) {
	testCases := []struct {
		name                   string
		totalSurplus           float64
		remainingBucketsToSell int64
		want                   float64
	}{
		{
			name:                   "no buckets remaining",
			totalSurplus:           100.0,
			remainingBucketsToSell: 0,
			want:                   100.0,
		}, {
			name:                   "negative buckets remaining",
			totalSurplus:           7236.24,
			remainingBucketsToSell: -1,
			want:                   7236.24,
		}, {
			name:                   "1 bucket remaining",
			totalSurplus:           7236.24,
			remainingBucketsToSell: 1,
			want:                   7236.24,
		}, {
			name:                   "20 buckets remaining (n = 1)",
			totalSurplus:           7236.24,
			remainingBucketsToSell: 20,
			want:                   7236.24,
		}, {
			name:                   "21 buckets remaining (n = 2)",
			totalSurplus:           7236.24,
			remainingBucketsToSell: 21,
			want:                   4824.16, // 7236.24 * (0.5-1.0) / (0.5^ceil(0.05*21) - 1.0)
		}, {
			name:                   "22 buckets remaining (n = 2)",
			totalSurplus:           7236.24,
			remainingBucketsToSell: 22,
			want:                   4824.16, // 7236.24 * (0.5-1.0) / (0.5^ceil(0.05*22) - 1.0)
		}, {
			name:                   "170 buckets remaining (n = 9)",
			totalSurplus:           7236.24,
			remainingBucketsToSell: 170,
			want:                   3625.200469667319, // 7236.24 * (0.5-1.0) / (0.5^ceil(0.05*170) - 1.0)
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			p := makeTestSellTwapLevelProvider(0)
			output := p.firstDistributionOfBaseSurplus(k.totalSurplus, k.remainingBucketsToSell)
			assert.Equal(t, k.want, output)
		})
	}
}

func TestBucketInfoString(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startTime := now.Add(time.Minute * -5)
	endTime := now.Add(time.Minute * 5)
	bucket := makeBucketInfo(
		bucketID(12),
		startTime,
		endTime,
		60,
		1440,
		120,
		5.0,
		1000.0,
		0.0,
		0.0,
		8.33333333,
		1.66666667,
		&dynamicBucketValues{
			isNew:       true,
			isLast:      false,
			roundID:     roundID(16),
			dayBaseSold: 5.0,
			baseSold:    0.0,
			now:         now,
		},
	)

	wantString := "BucketInfo[UUID=63129753083917721e25e22fb6f25b9ccd8f8aaa, date=2020-05-21, dayID=4 (Thursday), bucketID=12, startTime=2020-05-21T14:55:00Z, endTime=2020-05-21T15:05:00Z, sizeSeconds=60, totalBuckets=1440, totalBucketsToSell=120," +
		" dayBaseSoldStart=5.00000000, dayBaseCapacity=1000.00000000, totalBaseSurplusStart=0.00000000, baseSurplusIncluded=0.00000000, baseCapacity=8.33333333, minOrderSizeBase=1.66666667," +
		" DynamicBucketValues[isNew=true, isLast=false, roundID=16, dayBaseSold=5.00000000, dayBaseRemaining=995.00000000, baseSold=0.00000000, baseRemaining=8.33333333, bucketProgress=0.00%, bucketTimeElapsed=50.00%]]"
	assert.Equal(t, wantString, bucket.String())
}

func TestBucketInfoUUID(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	testCases := []struct {
		startTime                        time.Time
		endTime                          time.Time
		numHoursToSell                   int
		parentBucketSizeSeconds          int
		minChildOrderSizePercentOfParent float64
		bucketID                         int64
		roundID                          int64
		want                             string
	}{
		{
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         1,
			roundID:                          1,
			want:                             "63129753083917721e25e22fb6f25b9ccd8f8aaa",
		}, {
			startTime:                        now.Add(time.Minute * -6),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         1,
			roundID:                          1,
			want:                             "ef23fdd4c61c66befc7dec21f68fdb0ce48f9ee2",
		}, {
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 6),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         1,
			roundID:                          1,
			want:                             "4eaf99082e5a81d600d38384045addda16ed653b",
		}, {
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   3,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         1,
			roundID:                          1,
			want:                             "582488a2f7adf089ae2a655032d436dbc4fc4507",
		}, {
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          12,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         1,
			roundID:                          1,
			want:                             "15b49c42214e1ce40d02a762c84c246058242d45",
		}, {
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.3,
			bucketID:                         1,
			roundID:                          1,
			want:                             "63129753083917721e25e22fb6f25b9ccd8f8aaa",
		}, {
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         2,
			roundID:                          1,
			want:                             "63129753083917721e25e22fb6f25b9ccd8f8aaa",
		}, {
			startTime:                        now.Add(time.Minute * -5),
			endTime:                          now.Add(time.Minute * 5),
			numHoursToSell:                   2,
			parentBucketSizeSeconds:          60,
			minChildOrderSizePercentOfParent: 0.2,
			bucketID:                         1,
			roundID:                          2,
			want:                             "63129753083917721e25e22fb6f25b9ccd8f8aaa",
		},
	}

	for _, k := range testCases {
		t.Run(k.want, func(t *testing.T) {
			bucket := makeBucketInfo(
				bucketID(k.bucketID),
				k.startTime,
				k.endTime,
				k.parentBucketSizeSeconds,
				int64(24*60*60/k.parentBucketSizeSeconds),
				int64(k.numHoursToSell*60*60/k.parentBucketSizeSeconds),
				5.0,
				1000.0,
				0.0,
				0.0,
				8.33333333,
				k.minChildOrderSizePercentOfParent*8.33333333,
				&dynamicBucketValues{
					isNew:       true,
					isLast:      false,
					roundID:     roundID(k.roundID),
					dayBaseSold: 5.0,
					baseSold:    0.0,
					now:         now,
				},
			)

			assert.Equal(t, k.want, bucket.UUID())
		})
	}
}

func TestFinalizeBucket(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startTime := now.Add(time.Minute * -5)
	endTime := now.Add(time.Minute * 5)
	bucket := makeBucketInfo(
		bucketID(12),
		startTime,
		endTime,
		60,
		1440,
		120,
		5.0,
		1000.0,
		0.0,
		0.0,
		8.33333333,
		1.66666667,
		&dynamicBucketValues{
			isNew:       true,
			isLast:      false,
			roundID:     roundID(16),
			dayBaseSold: 5.0,
			baseSold:    0.0,
			now:         now,
		},
	)

	// make the call
	bucket = finalizeBucket(bucket)
	// ensure field we care about changed
	if !assert.True(t, bucket.dynamicValues.isLast) {
		return
	}

	// ensure nothing else changed
	wantString := "BucketInfo[UUID=63129753083917721e25e22fb6f25b9ccd8f8aaa, date=2020-05-21, dayID=4 (Thursday), bucketID=12, startTime=2020-05-21T14:55:00Z, endTime=2020-05-21T15:05:00Z, sizeSeconds=60, totalBuckets=1440, totalBucketsToSell=120," +
		" dayBaseSoldStart=5.00000000, dayBaseCapacity=1000.00000000, totalBaseSurplusStart=0.00000000, baseSurplusIncluded=0.00000000, baseCapacity=8.33333333, minOrderSizeBase=1.66666667," +
		" DynamicBucketValues[isNew=true, isLast=true, roundID=16, dayBaseSold=5.00000000, dayBaseRemaining=995.00000000, baseSold=0.00000000, baseRemaining=8.33333333, bucketProgress=0.00%, bucketTimeElapsed=50.00%]]"
	assert.Equal(t, wantString, bucket.String())
}
