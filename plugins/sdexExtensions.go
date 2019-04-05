package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/go/exp/clients/horizon"
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
	acceptedFee, e := getAcceptedFee(&feeStats, percentile)
	if e != nil {
		return 0, fmt.Errorf("could not fetch accepted fee: %s", e)
	}
	acceptedFeeInt64 := uint64(acceptedFee)

	if acceptedFeeInt64 <= maxOpFeeStroops {
		log.Printf("acceptedFeeInt64 <= maxOpFeeStroops; using acceptedFee of %d stroops at percentile=%d (maxOpFeeStroops=%d)\n", acceptedFeeInt64, percentile, maxOpFeeStroops)
		return acceptedFeeInt64, nil
	}
	log.Printf("acceptedFeeInt64 > maxOpFeeStroops; using maxOpFeeStroops of %d stroops (percentile=%d, acceptedFee=%d stroops)\n", maxOpFeeStroops, percentile, acceptedFeeInt64)
	return maxOpFeeStroops, nil
}

func getAcceptedFee(fs *hProtocol.FeeStats, percentile uint8) (int, error) {
	switch percentile {
	case 10:
		return fs.P10AcceptedFee, nil
	case 20:
		return fs.P20AcceptedFee, nil
	case 30:
		return fs.P30AcceptedFee, nil
	case 40:
		return fs.P40AcceptedFee, nil
	case 50:
		return fs.P50AcceptedFee, nil
	case 60:
		return fs.P60AcceptedFee, nil
	case 70:
		return fs.P70AcceptedFee, nil
	case 80:
		return fs.P80AcceptedFee, nil
	case 90:
		return fs.P90AcceptedFee, nil
	case 95:
		return fs.P95AcceptedFee, nil
	case 99:
		return fs.P99AcceptedFee, nil
	}
	return 0, fmt.Errorf("unable to get accepted fee for percentile: %d", percentile)
}
