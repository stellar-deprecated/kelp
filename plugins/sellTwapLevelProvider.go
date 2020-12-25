package plugins

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/postgresdb"
)

const secondsInHour = 60 * 60
const secondsInDay = 24 * secondsInHour
const timeFormat = time.RFC3339

// sellTwapLevelProvider provides a fixed number of levels using a static percentage spread
type sellTwapLevelProvider struct {
	startPf                                               api.PriceFeed
	offset                                                rateOffset
	orderConstraints                                      *model.OrderConstraints
	dowFilter                                             [7]volumeFilter
	numHoursToSell                                        int
	parentBucketSizeSeconds                               int
	distributeSurplusOverRemainingIntervalsPercentCeiling float64
	exponentialSmoothingFactor                            float64
	minChildOrderSizePercentOfParent                      float64
	random                                                *rand.Rand
	isBuySide                                             bool

	// uninitialized
	activeBucket    *bucketInfo
	previousRoundID *roundID
}

// ensure it implements the LevelProvider interface
var _ api.LevelProvider = &sellTwapLevelProvider{}

// makeSellTwapLevelProvider is a factory method
func makeSellTwapLevelProvider(
	startPf api.PriceFeed,
	offset rateOffset,
	orderConstraints *model.OrderConstraints,
	dowFilter [7]volumeFilter,
	numHoursToSell int,
	parentBucketSizeSeconds int,
	distributeSurplusOverRemainingIntervalsPercentCeiling float64,
	exponentialSmoothingFactor float64,
	minChildOrderSizePercentOfParent float64,
	randSeed int64,
	isBuySide bool,
) (api.LevelProvider, error) {
	if numHoursToSell <= 0 || numHoursToSell > 24 {
		return nil, fmt.Errorf("invalid number of hours to sell, expected 0 < numHoursToSell <= 24; was %d", numHoursToSell)
	}

	if parentBucketSizeSeconds <= 0 || parentBucketSizeSeconds > secondsInDay {
		return nil, fmt.Errorf("invalid value for parentBucketSizeSeconds, expected 0 < parentBucketSizeSeconds <= %d (secondsInDay); was %d", secondsInDay, parentBucketSizeSeconds)
	}

	if (secondsInDay % parentBucketSizeSeconds) != 0 {
		return nil, fmt.Errorf("parentBucketSizeSeconds needs to perfectly divide secondsInDay but it does not; secondsInDay is %d and parentBucketSizeSeconds was %d", secondsInDay, parentBucketSizeSeconds)
	}

	if distributeSurplusOverRemainingIntervalsPercentCeiling < 0.0 || distributeSurplusOverRemainingIntervalsPercentCeiling > 1.0 {
		return nil, fmt.Errorf("distributeSurplusOverRemainingIntervalsPercentCeiling is invalid, expected 0.0 <= distributeSurplusOverRemainingIntervalsPercentCeiling <= 1.0; was %.f", distributeSurplusOverRemainingIntervalsPercentCeiling)
	}

	if exponentialSmoothingFactor < 0.0 || exponentialSmoothingFactor > 1.0 {
		return nil, fmt.Errorf("exponentialSmoothingFactor is invalid, expected 0.0 <= exponentialSmoothingFactor <= 1.0; was %.f", exponentialSmoothingFactor)
	}

	if minChildOrderSizePercentOfParent < 0.0 || minChildOrderSizePercentOfParent > 1.0 {
		return nil, fmt.Errorf("minChildOrderSizePercentOfParent is invalid, expected 0.0 <= minChildOrderSizePercentOfParent <= 1.0; was %.f", exponentialSmoothingFactor)
	}

	for i, f := range dowFilter {
		if !f.isBase() {
			return nil, fmt.Errorf("volume filter at index %d was not constrained on the base asset as expected: %s (we currently only allow buy and sell constraints in base units)", i, f.configValue)
		}
	}

	random := rand.New(rand.NewSource(randSeed))
	return &sellTwapLevelProvider{
		startPf:                 startPf,
		offset:                  offset,
		orderConstraints:        orderConstraints,
		dowFilter:               dowFilter,
		numHoursToSell:          numHoursToSell,
		parentBucketSizeSeconds: parentBucketSizeSeconds,
		distributeSurplusOverRemainingIntervalsPercentCeiling: distributeSurplusOverRemainingIntervalsPercentCeiling,
		exponentialSmoothingFactor:                            exponentialSmoothingFactor,
		minChildOrderSizePercentOfParent:                      minChildOrderSizePercentOfParent,
		random:                                                random,
		isBuySide:                                             isBuySide,
	}, nil
}

type bucketID int64

type dynamicBucketValues struct {
	isNew       bool
	isLast      bool
	roundID     roundID
	dayBaseSold float64
	baseSold    float64
	now         time.Time
}

type bucketInfo struct {
	ID                 bucketID
	startTime          time.Time
	endTime            time.Time
	sizeSeconds        int
	totalBuckets       int64
	totalBucketsToSell int64
	dayBaseSoldStart   float64
	// currently we only allow dayBaseCapacity and not dayQuoteCapacity (or dayCapacity as a common field)
	// TODO NS allow quote capacity to work for twap and adjust log lines accordingly
	dayBaseCapacity float64
	// surplus can be negative because offers are outstanding and can be consumed while we run these level calculations. i.e. it can never be atomic.
	// the probability of this happening is small and as the execution speed of the update loop improves (with better code) the probability will go down.
	// It can be made logically atomic (with more guarantees than the fix in #456) by deleting outstanding offers in the pre-update data synchronization.
	// I think that is unnecessary since we'd rather keep the offer open for simplicity since we are not promising a perfect twap execution.
	// Alternatively, if we want to be perfect but also keep an offer outstanding at all times then in these level calculations we can assume that 100%
	// of the outstanding offer is consumed and we can pick an order size between baseCapacity - existingOfferBaseAmount. I think this is unnecessary too.
	totalBaseSurplusStart float64
	baseSurplusIncluded   float64
	baseCapacity          float64
	minOrderSizeBase      float64
	dynamicValues         *dynamicBucketValues
}

// makeBucketInfo is a factory method to ensure that we do not create a bucketInfo with missing inputs in the event
// that the struct is modified. Nobody should create a new bucketInfo outside of this function.
func makeBucketInfo(
	ID bucketID,
	startTime time.Time,
	endTime time.Time,
	sizeSeconds int,
	totalBuckets int64,
	totalBucketsToSell int64,
	dayBaseSoldStart float64,
	dayBaseCapacity float64,
	totalBaseSurplusStart float64,
	baseSurplusIncluded float64,
	baseCapacity float64,
	minOrderSizeBase float64,
	dynamicValues *dynamicBucketValues,
) *bucketInfo {
	return &bucketInfo{
		ID:                    ID,
		startTime:             startTime,
		endTime:               endTime,
		sizeSeconds:           sizeSeconds,
		totalBuckets:          totalBuckets,
		totalBucketsToSell:    totalBucketsToSell,
		dayBaseSoldStart:      dayBaseSoldStart,
		dayBaseCapacity:       dayBaseCapacity,
		totalBaseSurplusStart: totalBaseSurplusStart,
		baseSurplusIncluded:   baseSurplusIncluded,
		baseCapacity:          baseCapacity,
		minOrderSizeBase:      minOrderSizeBase,
		dynamicValues:         dynamicValues,
	}
}

func (b *bucketInfo) dayBaseRemaining() float64 {
	return b.dayBaseCapacity - b.dynamicValues.dayBaseSold
}

func (b *bucketInfo) baseRemaining() float64 {
	return b.baseCapacity - b.dynamicValues.baseSold
}

// String is the Stringer method
func (b *bucketInfo) String() string {
	return fmt.Sprintf("BucketInfo[UUID=%s, date=%s, dayID=%d (%s), bucketID=%d, startTime=%s, endTime=%s, sizeSeconds=%d, totalBuckets=%d, totalBucketsToSell=%d, dayBaseSoldStart=%.8f, dayBaseCapacity=%.8f, totalBaseSurplusStart=%.8f, baseSurplusIncluded=%.8f, baseCapacity=%.8f, minOrderSizeBase=%.8f, DynamicBucketValues[isNew=%v, isLast=%v, roundID=%d, dayBaseSold=%.8f, dayBaseRemaining=%.8f, baseSold=%.8f, baseRemaining=%.8f, bucketProgress=%.2f%%, bucketTimeElapsed=%.2f%%]]",
		b.UUID(),
		b.startTime.Format("2006-01-02"),
		b.startTime.Weekday(),
		b.startTime.Weekday().String(),
		b.ID,
		b.startTime.Format(timeFormat),
		b.endTime.Format(timeFormat),
		b.sizeSeconds,
		b.totalBuckets,
		b.totalBucketsToSell,
		b.dayBaseSoldStart,
		b.dayBaseCapacity,
		b.totalBaseSurplusStart,
		b.baseSurplusIncluded,
		b.baseCapacity,
		b.minOrderSizeBase,
		b.dynamicValues.isNew,
		b.dynamicValues.isLast,
		b.dynamicValues.roundID,
		b.dynamicValues.dayBaseSold,
		b.dayBaseRemaining(),
		b.dynamicValues.baseSold,
		b.baseRemaining(),
		100.0*b.dynamicValues.baseSold/b.baseCapacity,
		100.0*float64(b.dynamicValues.now.Unix()-b.startTime.Unix())/float64(b.endTime.Unix()-b.startTime.Unix()),
	)
}

// UUID gives a unique hash ID for this bucket that is unique to this specific configuration and time interval
// this should be constant for all bucket instances that overlap with this time interval and configuration
func (b *bucketInfo) UUID() string {
	timePartition := fmt.Sprintf("startTime=%s_endTime=%s", b.startTime.Format(time.RFC3339Nano), b.endTime.Format(time.RFC3339Nano))
	configPartition := fmt.Sprintf("totalBuckets=%d_totalBucketsToSell=%d", b.totalBuckets, b.totalBucketsToSell)
	s := fmt.Sprintf("timePartition=%s__configPartition=%s", timePartition, configPartition)

	hash := sha1.Sum([]byte(s))
	return fmt.Sprintf("%x", hash)
}

type roundID uint64

type roundInfo struct {
	ID                  roundID
	bucketID            bucketID
	bucketUUID          string
	now                 time.Time
	secondsElapsedToday int64
	sizeBaseCapped      float64
	price               float64
}

// String is the Stringer method
func (r *roundInfo) String() string {
	return fmt.Sprintf(
		"RoundInfo[roundID=%d, bucketID=%d, bucketUUID=%s, now=%s (day=%s, secondsElapsedToday=%d), sizeBaseCapped=%.8f, price=%.8f]",
		r.ID,
		r.bucketID,
		r.bucketUUID,
		r.now.Format(timeFormat),
		r.now.Weekday().String(),
		r.secondsElapsedToday,
		r.sizeBaseCapped,
		r.price,
	)
}

// GetLevels impl.
func (p *sellTwapLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	now := time.Now().UTC()
	log.Printf("GetLevels, unix timestamp for 'now' in UTC = %d (%s)\n", now.Unix(), now)

	volFilter := p.dowFilter[now.Weekday()]
	log.Printf("volumeFilter = %s\n", volFilter.String())

	rID := p.makeRoundID()
	oldBucket, activeBucket, e := p.makeActiveBucket(now, volFilter, rID)
	if e != nil {
		return nil, fmt.Errorf("unable to make bucketInfo: %s", e)
	}

	round, e := p.makeRoundInfo(rID, now, activeBucket)
	if e != nil {
		return nil, fmt.Errorf("unable to make roundInfo: %s", e)
	}

	// structured log line for metric tracking via log files
	if oldBucket != nil {
		log.Printf("bucketInfo: %s; roundInfo: %s\n", oldBucket, round)
	}
	log.Printf("bucketInfo: %s; roundInfo: %s\n", activeBucket, round)

	// save activeBucket and round for future rounds
	p.activeBucket = activeBucket
	p.previousRoundID = &round.ID

	if round.sizeBaseCapped < p.orderConstraints.MinBaseVolume.AsFloat() {
		return []api.Level{}, nil
	}

	// we invert the price for buy side
	price := round.price
	if p.isBuySide {
		price = 1 / price
	}

	return []api.Level{{
		Price:  *model.NumberFromFloat(price, p.orderConstraints.PricePrecision),
		Amount: *model.NumberFromFloat(round.sizeBaseCapped, p.orderConstraints.VolumePrecision),
	}}, nil
}

func (p *sellTwapLevelProvider) makeFirstBucketFrame(
	now time.Time,
	startTime time.Time,
	endTime time.Time,
	bID bucketID,
	rID roundID,
	dayBaseCapacity float64,
	dailyVolumeValues *queries.DailyVolume,
) (*bucketInfo, error) {
	dayStartTime := floorDate(now)
	dayEndTime := ceilDate(now)
	totalBuckets := int64(math.Ceil(float64(dayEndTime.Unix()-dayStartTime.Unix()) / float64(p.parentBucketSizeSeconds)))
	totalBucketsToSell := int64(math.Ceil(float64(p.numHoursToSell*secondsInHour) / float64(p.parentBucketSizeSeconds)))
	dayBaseSoldStart := dailyVolumeValues.BaseVol

	// the total surplus remaining up until this point gets distributed over the remaining buckets
	averageBaseCapacity := float64(dayBaseCapacity) / float64(totalBucketsToSell)
	numPreviousBuckets := bID // buckets are 0-indexed, so bucketID is equal to numbers of previous buckets
	expectedSold := averageBaseCapacity * float64(numPreviousBuckets)
	// we have special logic for buckets after selling hours to ensure we don't expect a larger amount sold
	if int64(numPreviousBuckets) >= totalBucketsToSell {
		expectedSold = dayBaseCapacity
	}
	totalBaseSurplusStart := expectedSold - dayBaseSoldStart
	remainingBucketsToSell := totalBucketsToSell - int64(numPreviousBuckets)
	baseSurplusIncluded := p.firstDistributionOfBaseSurplus(totalBaseSurplusStart, remainingBucketsToSell)
	baseCapacity := baseSurplusIncluded
	if remainingBucketsToSell > 0 {
		// only include the averageBaseCapacity if we are within the number of total buckets to sell
		// else we are in a state where there is no "new" capacity for every bucket and we are only
		// trying to get rid of past surplus values
		baseCapacity += averageBaseCapacity
	}
	minOrderSizeBase := p.minChildOrderSizePercentOfParent * baseCapacity
	// upon instantiation the first bucket frame does not have anything sold beyond the starting values
	dynamicValues := &dynamicBucketValues{
		isNew:       true,
		isLast:      false,
		roundID:     rID,
		dayBaseSold: dayBaseSoldStart,
		baseSold:    0.0, // always 0.0 for a new bucket
		now:         now,
	}

	newBucket := makeBucketInfo(
		bID,
		startTime,
		endTime,
		p.parentBucketSizeSeconds,
		totalBuckets,
		totalBucketsToSell,
		dayBaseSoldStart,
		dayBaseCapacity,
		totalBaseSurplusStart,
		baseSurplusIncluded,
		baseCapacity,
		minOrderSizeBase,
		dynamicValues,
	)
	return newBucket, nil
}

func (p *sellTwapLevelProvider) updateExistingBucket(now time.Time, dailyVolumeValues *queries.DailyVolume, rID roundID) (*bucketInfo, error) {
	bucketCopy := *p.activeBucket
	bucket := &bucketCopy
	dayBaseSold := dailyVolumeValues.BaseVol

	bucket.dynamicValues = &dynamicBucketValues{
		isNew: false,
		//isLast stays same
		roundID:     rID,
		dayBaseSold: dayBaseSold,
		baseSold:    dayBaseSold - bucket.dayBaseSoldStart,
		now:         now,
	}
	return bucket, nil
}

func finalizeBucket(bucket *bucketInfo) *bucketInfo {
	bucket.dynamicValues.isLast = true
	return bucket
}

func (p *sellTwapLevelProvider) makeActiveBucket(now time.Time, volFilter volumeFilter, rID roundID) ( /*oldBucket*/ *bucketInfo /*activeBucket*/, *bucketInfo, error) {
	dayStartTime := floorDate(now)
	secondsElapsedToday := now.Unix() - dayStartTime.Unix()
	bID := bucketID(secondsElapsedToday / int64(p.parentBucketSizeSeconds))
	startTime := dayStartTime.Add(time.Second * time.Duration(bID) * time.Duration(p.parentBucketSizeSeconds))
	endTime := startTime.Add(time.Second*time.Duration(p.parentBucketSizeSeconds) - time.Nanosecond)

	dayBaseCapacity, e := volFilter.mustGetBaseAssetCapInBaseUnits()
	if e != nil {
		return nil, nil, fmt.Errorf("could not fetch base asset cap in base units: %s", e)
	}
	queryResult, e := volFilter.dailyVolumeByDateQuery.QueryRow(now.Format(postgresdb.DateFormatString))
	if e != nil {
		return nil, nil, fmt.Errorf("could not fetch daily values for today: %s", e)
	}
	dailyVolumeValues, ok := queryResult.(*queries.DailyVolume)
	if !ok {
		return nil, nil, fmt.Errorf("could not cast query result from dailyValuesByDateQuery as a *queries.DailyVolume, was type '%T'", queryResult)
	}

	// bucket on bot load
	if p.activeBucket == nil {
		bucket, e := p.makeFirstBucketFrame(now, startTime, endTime, bID, rID, dayBaseCapacity, dailyVolumeValues)
		if e != nil {
			return nil, nil, fmt.Errorf("could not make first bucket: %s", e)
		}
		return nil, bucket, nil
	}

	// always update existing bucket with latest volume numbers
	bucket, e := p.updateExistingBucket(now, dailyVolumeValues, rID)
	if e != nil {
		return nil, nil, fmt.Errorf("could not update existing bucket (ID=%d): %s", bID, e)
	}

	// new round in the same bucket should be returned
	if bID == p.activeBucket.ID {
		return nil, bucket, nil
	}

	// finalize the current bucket
	oldBucket := finalizeBucket(bucket)

	// new bucket needs to be created
	newBucket, e := p.makeFirstBucketFrame(now, startTime, endTime, bID, rID, dayBaseCapacity, dailyVolumeValues)
	if e != nil {
		return nil, nil, fmt.Errorf("unable to make first bucket frame when cutting over with new bucketID (ID=%d): %s", bID, e)
	}
	// always return oldBucket along with the newBucket (do not check bucketID validity)
	return oldBucket, newBucket, nil
}

/*
Using a geometric series calculation:
Sn = a * (r^n - 1) / (r - 1)
a = Sn * (r - 1) / (r^n - 1)
a = 8,000 * (0.5 - 1) / (0.5^4 - 1)
a = 8,000 * (-0.5) / (0.0625 - 1)
a = 8,000 * (0.5/0.9375)
a = 4,266.67
*/
func (p *sellTwapLevelProvider) firstDistributionOfBaseSurplus(totalSurplus float64, remainingBucketsToSell int64) float64 {
	if remainingBucketsToSell <= 0 {
		return totalSurplus
	}

	Sn := totalSurplus
	r := p.exponentialSmoothingFactor
	n := math.Ceil(p.distributeSurplusOverRemainingIntervalsPercentCeiling * float64(remainingBucketsToSell))

	a := Sn * (r - 1.0) / (math.Pow(r, n) - 1.0)
	return a
}

func (p *sellTwapLevelProvider) makeRoundID() roundID {
	if p.previousRoundID == nil {
		return roundID(0)
	}
	return *p.previousRoundID + 1
}

func (p *sellTwapLevelProvider) makeRoundInfo(rID roundID, now time.Time, bucket *bucketInfo) (*roundInfo, error) {
	dayStartTime := floorDate(now)
	secondsElapsedToday := now.Unix() - dayStartTime.Unix()

	var sizeBaseCapped float64
	if bucket.baseRemaining() <= bucket.minOrderSizeBase {
		sizeBaseCapped = bucket.baseRemaining()
	} else {
		sizeBaseCapped = bucket.minOrderSizeBase + (p.random.Float64() * (bucket.baseRemaining() - bucket.minOrderSizeBase))
	}

	price, e := p.startPf.GetPrice()
	if e != nil {
		return nil, fmt.Errorf("could not get price from feed: %s", e)
	}
	adjustedPrice, wasModified := p.offset.apply(price)
	if wasModified {
		log.Printf("feed price (adjusted): %.8f\n", adjustedPrice)
	}

	return &roundInfo{
		ID:                  rID,
		bucketID:            bucket.ID,
		bucketUUID:          bucket.UUID(),
		now:                 now,
		secondsElapsedToday: secondsElapsedToday,
		sizeBaseCapped:      sizeBaseCapped,
		price:               adjustedPrice,
	}, nil
}

// GetFillHandlers impl
func (p *sellTwapLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return nil, nil
}

func floorDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func ceilDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, -1, t.Location())
}
