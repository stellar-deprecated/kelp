package plugins

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
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
		2,
		60,
		0.05,
		0.5,
		0.2,
		seed,
	)
	if e != nil {
		panic(e)
	}
	return p.(*sellTwapLevelProvider)
}

func TestMakeFirstBucketFrame1(t *testing.T) {
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
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, bucketID(0), bucketInfo.ID)
	assert.Equal(t, 8.333333333333334, bucketInfo.baseCapacity)
	assert.Equal(t, 8.333333333333334, bucketInfo.baseRemaining())
	assert.Equal(t, 0.0, bucketInfo.baseSurplusIncluded)
	assert.Equal(t, 1000.0, bucketInfo.dayBaseCapacity)
	assert.Equal(t, 1000.0, bucketInfo.dayBaseRemaining())
	assert.Equal(t, 0.0, bucketInfo.dayBaseSoldStart)
	assert.Equal(t, 0.0, bucketInfo.dynamicValues.baseSold)
	assert.Equal(t, 0.0, bucketInfo.dynamicValues.dayBaseSold)
	assert.Equal(t, true, bucketInfo.dynamicValues.isNew)
	assert.Equal(t, now, bucketInfo.dynamicValues.now)
	assert.Equal(t, roundID(0), bucketInfo.dynamicValues.roundID)
	assert.Equal(t, endDate, bucketInfo.endTime)
	assert.Equal(t, 1.666666666666667, bucketInfo.minOrderSizeBase)
	assert.Equal(t, 60, bucketInfo.sizeSeconds)
	assert.Equal(t, startDate, bucketInfo.startTime)
	assert.Equal(t, 0.0, bucketInfo.totalBaseSurplusStart)
	assert.Equal(t, int64(1440), bucketInfo.totalBuckets)
	assert.Equal(t, int64(120), bucketInfo.totalBucketsToSell)
}

func TestMakeFirstBucketFrame2(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startDate := now.Add(time.Minute * -5)
	endDate := now.Add(time.Minute * 5)
	p := makeTestSellTwapLevelProvider(0)
	bucketInfo, e := p.makeFirstBucketFrame(
		now,
		startDate,
		endDate,
		bucketID(0),
		roundID(1),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  5.0,
			QuoteVol: 1.0,
		},
	)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, bucketID(0), bucketInfo.ID)
	assert.Equal(t, 8.333333333333334, bucketInfo.baseCapacity)
	assert.Equal(t, 8.333333333333334, bucketInfo.baseRemaining())
	assert.Equal(t, 0.0, bucketInfo.baseSurplusIncluded)
	assert.Equal(t, 1000.0, bucketInfo.dayBaseCapacity)
	assert.Equal(t, 995.0, bucketInfo.dayBaseRemaining())
	assert.Equal(t, 5.0, bucketInfo.dayBaseSoldStart)
	assert.Equal(t, 0.0, bucketInfo.dynamicValues.baseSold)
	assert.Equal(t, 5.0, bucketInfo.dynamicValues.dayBaseSold)
	assert.Equal(t, true, bucketInfo.dynamicValues.isNew)
	assert.Equal(t, now, bucketInfo.dynamicValues.now)
	assert.Equal(t, roundID(1), bucketInfo.dynamicValues.roundID)
	assert.Equal(t, endDate, bucketInfo.endTime)
	assert.Equal(t, 1.666666666666667, bucketInfo.minOrderSizeBase)
	assert.Equal(t, 60, bucketInfo.sizeSeconds)
	assert.Equal(t, startDate, bucketInfo.startTime)
	assert.Equal(t, 0.0, bucketInfo.totalBaseSurplusStart)
	assert.Equal(t, int64(1440), bucketInfo.totalBuckets)
	assert.Equal(t, int64(120), bucketInfo.totalBucketsToSell)
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

func TestCutoverToNewBucketSameDay_InsideNumHoursToSell(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startDate := now.Add(time.Minute * -5)
	endDate := now.Add(time.Minute * 5)
	p := makeTestSellTwapLevelProvider(0)
	activeBucket, e := p.makeFirstBucketFrame(
		now,
		startDate,
		endDate,
		bucketID(3),
		roundID(1),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  5.0,
			QuoteVol: 1.0,
		},
	)
	if e != nil {
		panic(e)
	}

	p.activeBucket = activeBucket
	now2 := now.Add(time.Second * 30)
	startDate2 := now2.Add(time.Minute * -5)
	endDate2 := now2.Add(time.Minute * 5)
	newBucket, e := p.makeFirstBucketFrame(
		now2,
		startDate2,
		endDate2,
		bucketID(4),
		roundID(2),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  5.0,
			QuoteVol: 1.0,
		},
	)
	if e != nil {
		panic(e)
	}
	cutoverBucket, e := p.cutoverToNewBucketSameDay(newBucket)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, bucketID(4), cutoverBucket.ID)
	assert.Equal(t, 22.724867724867728, cutoverBucket.baseCapacity) // 8.333333333333334 + 14.391534391534393
	assert.Equal(t, 22.724867724867728, cutoverBucket.baseRemaining())
	assert.Equal(t, 14.391534391534393, cutoverBucket.baseSurplusIncluded) // 28.333333333333334*(.5-1)/((.5^ceil(.05*(120-4)))-1)
	assert.Equal(t, 1000.0, cutoverBucket.dayBaseCapacity)
	assert.Equal(t, 995.0, cutoverBucket.dayBaseRemaining())
	assert.Equal(t, 5.0, cutoverBucket.dayBaseSoldStart)
	assert.Equal(t, 0.0, cutoverBucket.dynamicValues.baseSold)
	assert.Equal(t, 5.0, cutoverBucket.dynamicValues.dayBaseSold)
	assert.Equal(t, true, cutoverBucket.dynamicValues.isNew)
	assert.Equal(t, now2, cutoverBucket.dynamicValues.now)
	assert.Equal(t, roundID(2), cutoverBucket.dynamicValues.roundID)
	assert.Equal(t, endDate2, cutoverBucket.endTime)
	assert.Equal(t, 4.544973544973546, cutoverBucket.minOrderSizeBase)
	assert.Equal(t, 60, cutoverBucket.sizeSeconds)
	assert.Equal(t, startDate2, cutoverBucket.startTime)
	assert.Equal(t, 28.333333333333334, cutoverBucket.totalBaseSurplusStart)
	assert.Equal(t, int64(1440), cutoverBucket.totalBuckets)
	assert.Equal(t, int64(120), cutoverBucket.totalBucketsToSell)
}

func TestCutoverToNewBucketSameDay_OutsideNumHoursToSell(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startDate := now.Add(time.Minute * -5)
	endDate := now.Add(time.Minute * 5)
	p := makeTestSellTwapLevelProvider(0)
	activeBucket, e := p.makeFirstBucketFrame(
		now,
		startDate,
		endDate,
		bucketID(119), // last bucket inside num hours to sell based on test levelProvider
		roundID(3),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  983.333412,
			QuoteVol: 98.3333412,
		},
	)
	if e != nil {
		panic(e)
	}

	p.activeBucket = activeBucket
	now2 := now.Add(time.Second * 30)
	startDate2 := now2.Add(time.Minute * -5)
	endDate2 := now2.Add(time.Minute * 5)
	newBucket, e := p.makeFirstBucketFrame(
		now2,
		startDate2,
		endDate2,
		bucketID(120),
		roundID(4),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  983.333412,
			QuoteVol: 98.3333412,
		},
	)
	if e != nil {
		panic(e)
	}
	cutoverBucket, e := p.cutoverToNewBucketSameDay(newBucket)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, bucketID(120), cutoverBucket.ID)
	assert.Equal(t, 16.66658800000016, cutoverBucket.baseCapacity) // remainingBucketsToSell == 0, so equals exactly the entire surplus
	assert.Equal(t, 16.66658800000016, cutoverBucket.baseRemaining())
	assert.Equal(t, 16.66658800000016, cutoverBucket.baseSurplusIncluded) // remainingBucketsToSell == 0, so include entire surplus
	assert.Equal(t, 1000.0, cutoverBucket.dayBaseCapacity)
	assert.Equal(t, 16.666588000000047, cutoverBucket.dayBaseRemaining())
	assert.Equal(t, 983.333412, cutoverBucket.dayBaseSoldStart)
	assert.Equal(t, 0.0, cutoverBucket.dynamicValues.baseSold)
	assert.Equal(t, 983.333412, cutoverBucket.dynamicValues.dayBaseSold)
	assert.Equal(t, true, cutoverBucket.dynamicValues.isNew)
	assert.Equal(t, now2, cutoverBucket.dynamicValues.now)
	assert.Equal(t, roundID(4), cutoverBucket.dynamicValues.roundID)
	assert.Equal(t, endDate2, cutoverBucket.endTime)
	assert.Equal(t, 3.3333176000000324, cutoverBucket.minOrderSizeBase)
	assert.Equal(t, 60, cutoverBucket.sizeSeconds)
	assert.Equal(t, startDate2, cutoverBucket.startTime)
	assert.Equal(t, 16.66658800000016, cutoverBucket.totalBaseSurplusStart) // 120*8.333333333333334 - 983.333412
	assert.Equal(t, int64(1440), cutoverBucket.totalBuckets)
	assert.Equal(t, int64(120), cutoverBucket.totalBucketsToSell)
}

func TestCutoverToNewBucketSameDay_WithNewVolumeSold(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-05-21T15:00:00Z")
	startDate := now.Add(time.Minute * -5)
	endDate := now.Add(time.Minute * 5)
	p := makeTestSellTwapLevelProvider(0)
	activeBucket, e := p.makeFirstBucketFrame(
		now,
		startDate,
		endDate,
		bucketID(3),
		roundID(1),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  5.0,
			QuoteVol: 1.0,
		},
	)
	if e != nil {
		panic(e)
	}

	p.activeBucket = activeBucket
	now2 := now.Add(time.Second * 30)
	startDate2 := now2.Add(time.Minute * -5)
	endDate2 := now2.Add(time.Minute * 5)
	newBucket, e := p.makeFirstBucketFrame(
		now2,
		startDate2,
		endDate2,
		bucketID(4),
		roundID(2),
		1000.0,
		&queries.DailyVolume{
			BaseVol:  6.0,
			QuoteVol: 1.1,
		},
	)
	if e != nil {
		panic(e)
	}
	cutoverBucket, e := p.cutoverToNewBucketSameDay(newBucket)
	if !assert.NoError(t, e) {
		return
	}

	// when there is additional volume in a cutover scenario then we include the volume in the newly created bucket for simplicity of calculations
	assert.Equal(t, bucketID(4), cutoverBucket.ID)
	assert.Equal(t, 22.724867724867728, cutoverBucket.baseCapacity) // 8.333333333333334 + 14.391534391534393
	assert.Equal(t, 21.724867724867728, cutoverBucket.baseRemaining())
	assert.Equal(t, 14.391534391534393, cutoverBucket.baseSurplusIncluded) // 28.333333333333334*(.5-1)/((.5^ceil(.05*(120-4)))-1)
	assert.Equal(t, 1000.0, cutoverBucket.dayBaseCapacity)
	assert.Equal(t, 994.0, cutoverBucket.dayBaseRemaining())
	assert.Equal(t, 5.0, cutoverBucket.dayBaseSoldStart)
	assert.Equal(t, 1.0, cutoverBucket.dynamicValues.baseSold)
	assert.Equal(t, 6.0, cutoverBucket.dynamicValues.dayBaseSold)
	assert.Equal(t, true, cutoverBucket.dynamicValues.isNew)
	assert.Equal(t, now2, cutoverBucket.dynamicValues.now)
	assert.Equal(t, roundID(2), cutoverBucket.dynamicValues.roundID)
	assert.Equal(t, endDate2, cutoverBucket.endTime)
	assert.Equal(t, 4.544973544973546, cutoverBucket.minOrderSizeBase)
	assert.Equal(t, 60, cutoverBucket.sizeSeconds)
	assert.Equal(t, startDate2, cutoverBucket.startTime)
	assert.Equal(t, 28.333333333333334, cutoverBucket.totalBaseSurplusStart)
	assert.Equal(t, int64(1440), cutoverBucket.totalBuckets)
	assert.Equal(t, int64(120), cutoverBucket.totalBucketsToSell)
}
