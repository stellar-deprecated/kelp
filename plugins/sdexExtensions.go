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
	output := FeeStatsResponse{}
	e := networking.JSONRequest(http.DefaultClient, "GET", feeStatsURL, "", map[string]string{}, &output, "")
	if e != nil {
		return 0, fmt.Errorf("error fetching fee stats (URL=%s): %s", feeStatsURL, e)
	}

	lastFeeInt, e := strconv.Atoi(output.LastLedgerBaseFee)
	if e != nil {
		return 0, fmt.Errorf("could not parse last_ledger_base_fee (%s) as int: %s", output.LastLedgerBaseFee, e)
	}
	modeFeeInt, e := strconv.Atoi(output.ModeAcceptedFee)
	if e != nil {
		return 0, fmt.Errorf("could not parse mode_accepted_fee (%s) as int: %s", output.ModeAcceptedFee, e)
	}
	lastFee := uint64(lastFeeInt)
	modeFee := uint64(modeFeeInt)

	if lastFee >= modeFee && lastFee <= maxOpFeeStroops {
		log.Printf("using last_ledger_base_fee of %d stroops (maxBaseFee = %d)\n", lastFee, maxOpFeeStroops)
		return lastFee, nil
	}
	if modeFee >= lastFee && modeFee <= maxOpFeeStroops {
		log.Printf("using mode_accepted_fee of %d stroops (maxBaseFee = %d)\n", modeFee, maxOpFeeStroops)
		return modeFee, nil
	}
	log.Printf("using maxBaseFee of %d stroops (last_ledger_base_fee = %d; mode_accepted_fee = %d)\n", maxOpFeeStroops, lastFee, modeFee)
	return maxOpFeeStroops, nil
}

// FeeStatsResponse represents the response from /fee_stats
type FeeStatsResponse struct {
	LastLedger        string `json:"last_ledger"`          // uint64 as a string
	LastLedgerBaseFee string `json:"last_ledger_base_fee"` // uint64 as a string
	MinAcceptedFee    string `json:"min_accepted_fee"`     // uint64 as a string
	ModeAcceptedFee   string `json:"mode_accepted_fee"`    // uint64 as a string
}
