package plugins

import (
	"fmt"
	"log"
	"strings"

	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
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
	capacityTrigger float64,
	percentile uint8,
	maxOpFeeStroops uint64,
	newClient *horizonclient.Client,
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
		return getFeeFromStats(newClient, capacityTrigger, percentile, maxOpFeeStroops)
	}, nil
}

func getFeeFromStats(horizonClient horizonclient.ClientInterface, capacityTrigger float64, percentile uint8, maxOpFeeStroops uint64) (uint64, error) {
	feeStats, e := horizonClient.FeeStats()
	if e != nil {
		// if the endpoint is not available (horizon-specific error) then use the maxOpFeeStroops
		// the Stellar network will use the base fee and increase it when necessary up to a max of the fee specified
		if strings.Contains(e.Error(), "Endpoint Not Available") {
			log.Printf("endpoint was not available so using maxOpFeeStroops passed in of %d stroops\n", maxOpFeeStroops)
			return maxOpFeeStroops, nil
		}
		return 0, fmt.Errorf("error fetching fee stats: %s", e)
	}

	// case where we don't trigger the dynamic fees logic
	if feeStats.LedgerCapacityUsage < capacityTrigger {
		lastFee := uint64(feeStats.LastLedgerBaseFee)
		if lastFee <= maxOpFeeStroops {
			log.Printf("lastFee <= maxOpFeeStroops; using last_ledger_base_fee of %d stroops (maxOpFeeStroops = %d)\n", lastFee, maxOpFeeStroops)
			return lastFee, nil
		}
		log.Printf("lastFee > maxOpFeeStroops; using maxOpFeeStroops of %d stroops (lastFee = %d)\n", maxOpFeeStroops, lastFee)
		return maxOpFeeStroops, nil
	}

	// parse percentile value
	maxFee, e := getMaxFee(&feeStats, percentile)
	if e != nil {
		return 0, fmt.Errorf("could not fetch max fee: %s", e)
	}
	maxFeeInt64 := uint64(maxFee)

	if maxFeeInt64 <= maxOpFeeStroops {
		log.Printf("maxFeeInt64 <= maxOpFeeStroops; using maxFee of %d stroops at percentile=%d (maxOpFeeStroops=%d)\n", maxFeeInt64, percentile, maxOpFeeStroops)
		return maxFeeInt64, nil
	}
	log.Printf("maxFeeInt64 > maxOpFeeStroops; using maxOpFeeStroops of %d stroops (percentile=%d, maxFee=%d stroops)\n", maxOpFeeStroops, percentile, maxFeeInt64)
	return maxOpFeeStroops, nil
}

func getMaxFee(fs *hProtocol.FeeStats, percentile uint8) (int64, error) {
	switch percentile {
	case 10:
		return fs.MaxFee.P10, nil
	case 20:
		return fs.MaxFee.P20, nil
	case 30:
		return fs.MaxFee.P30, nil
	case 40:
		return fs.MaxFee.P40, nil
	case 50:
		return fs.MaxFee.P50, nil
	case 60:
		return fs.MaxFee.P60, nil
	case 70:
		return fs.MaxFee.P70, nil
	case 80:
		return fs.MaxFee.P80, nil
	case 90:
		return fs.MaxFee.P90, nil
	case 95:
		return fs.MaxFee.P95, nil
	case 99:
		return fs.MaxFee.P99, nil
	}
	return 0, fmt.Errorf("unable to get max fee for percentile: %d", percentile)
}
