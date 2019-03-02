package plugins

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/stellar/kelp/support/networking"
)

// OpFeeStroops computes fees per operation
type OpFeeStroops func() (uint64, error)

// SdexFixedFeeFn returns a fixedFee in stroops
func SdexFixedFeeFn(fixedFeeStroops uint64) OpFeeStroops {
	return func() (uint64, error) {
		return fixedFeeStroops, nil
	}
}

const baseFeeStroops = 100

var validPercentiles = []uint8{10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99}

// SdexFeeFnFromStats returns an OpFeeStroops that uses the /fee_stats endpoint
func SdexFeeFnFromStats(
	horizonBaseURL string,
	capacityTrigger float64,
	percentile uint8,
	maxOpFeeStroops uint64,
) (OpFeeStroops, error) {
	isValid := false
	for _, p := range validPercentiles {
		if percentile == p {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, fmt.Errorf("unable to create SdexFeeFnFromStats since percentile is invalid (%d). Allowed values: %v", percentile, validPercentiles)
	}

	if capacityTrigger <= 0 {
		return nil, fmt.Errorf("unable to create SdexFeeFnFromStats, capacityTrigger should be > 0: %f", capacityTrigger)
	}

	if maxOpFeeStroops < baseFeeStroops {
		return nil, fmt.Errorf("unable to create SdexFeeFnFromStats, maxOpFeeStroops should be >= %d (baseFeeStroops): %d", baseFeeStroops, maxOpFeeStroops)
	}

	return func() (uint64, error) {
		return getFeeFromStats(horizonBaseURL, capacityTrigger, percentile, maxOpFeeStroops)
	}, nil
}

func getFeeFromStats(horizonBaseURL string, capacityTrigger float64, percentile uint8, maxOpFeeStroops uint64) (uint64, error) {
	feeStatsURL := horizonBaseURL + "/fee_stats"

	feeStatsResponseMap := map[string]string{}
	e := networking.JSONRequest(http.DefaultClient, "GET", feeStatsURL, "", map[string]string{}, &feeStatsResponseMap, "")
	if e != nil {
		return 0, fmt.Errorf("error fetching fee stats (URL=%s): %s", feeStatsURL, e)
	}

	// parse ledgerCapacityUsage
	ledgerCapacityUsage, e := strconv.ParseFloat(feeStatsResponseMap["ledger_capacity_usage"], 64)
	if e != nil {
		return 0, fmt.Errorf("could not parse ledger_capacity_usage ('%s') as float64: %s", feeStatsResponseMap["ledger_capacity_usage"], e)
	}

	// case where we don't trigger the dynamic fees logic
	if ledgerCapacityUsage < capacityTrigger {
		var lastFeeInt int
		lastFeeInt, e = strconv.Atoi(feeStatsResponseMap["last_ledger_base_fee"])
		if e != nil {
			return 0, fmt.Errorf("could not parse last_ledger_base_fee ('%s') as int: %s", feeStatsResponseMap["last_ledger_base_fee"], e)
		}
		lastFee := uint64(lastFeeInt)
		if lastFee <= maxOpFeeStroops {
			log.Printf("lastFee <= maxOpFeeStroops; using last_ledger_base_fee of %d stroops (maxOpFeeStroops = %d)\n", lastFee, maxOpFeeStroops)
			return lastFee, nil
		}
		log.Printf("lastFee > maxOpFeeStroops; using maxOpFeeStroops of %d stroops (lastFee = %d)\n", maxOpFeeStroops, lastFee)
		return maxOpFeeStroops, nil
	}

	// parse percentile value
	var pStroopsInt64 uint64
	pKey := fmt.Sprintf("p%d_accepted_fee", percentile)
	if pStroops, ok := feeStatsResponseMap[pKey]; ok {
		var pStroopsInt int
		pStroopsInt, e = strconv.Atoi(pStroops)
		if e != nil {
			return 0, fmt.Errorf("could not parse percentile value (%s='%s'): %s", pKey, pStroops, e)
		}
		pStroopsInt64 = uint64(pStroopsInt)
	} else {
		return 0, fmt.Errorf("could not fetch percentile value (%s) from feeStatsResponseMap: %s", pKey, e)
	}

	if pStroopsInt64 <= maxOpFeeStroops {
		log.Printf("pStroopsInt64 <= maxOpFeeStroops; using %s of %d stroops (maxOpFeeStroops = %d)\n", pKey, pStroopsInt64, maxOpFeeStroops)
		return pStroopsInt64, nil
	}
	log.Printf("pStroopsInt64 > maxOpFeeStroops; using maxOpFeeStroops of %d stroops (%s = %d)\n", maxOpFeeStroops, pKey, pStroopsInt64)
	return maxOpFeeStroops, nil
}
