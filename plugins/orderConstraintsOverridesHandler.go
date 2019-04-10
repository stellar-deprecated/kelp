package plugins

import "github.com/stellar/kelp/model"

// OrderConstraintsOverridesHandler knows how to capture overrides and apply them onto OrderConstraints
type OrderConstraintsOverridesHandler struct {
	overrides map[string]*model.OrderConstraintsOverride
}

// MakeEmptyOrderConstraintsOverridesHandler is a factory method
func MakeEmptyOrderConstraintsOverridesHandler() *OrderConstraintsOverridesHandler {
	return &OrderConstraintsOverridesHandler{
		overrides: map[string]*model.OrderConstraintsOverride{},
	}
}

// MakeOrderConstraintsOverridesHandler is a factory method
func MakeOrderConstraintsOverridesHandler(inputs map[model.TradingPair]model.OrderConstraints) *OrderConstraintsOverridesHandler {
	overrides := map[string]*model.OrderConstraintsOverride{}
	for p, oc := range inputs {
		overrides[p.String()] = model.MakeOrderConstraintsOverrideFromConstraints(&oc)
	}

	return &OrderConstraintsOverridesHandler{
		overrides: overrides,
	}
}

// Apply creates a new order constraints after checking for any existing overrides
func (ocHandler *OrderConstraintsOverridesHandler) Apply(pair *model.TradingPair, oc *model.OrderConstraints) *model.OrderConstraints {
	override, has := ocHandler.overrides[pair.String()]
	if !has {
		return oc
	}
	return model.MakeOrderConstraintsWithOverride(*oc, override)
}

// Get impl, panics if the override does not exist
func (ocHandler *OrderConstraintsOverridesHandler) Get(pair *model.TradingPair) *model.OrderConstraintsOverride {
	return ocHandler.overrides[pair.String()]
}

// Upsert allows you to set overrides to partially override values for specific pairs
func (ocHandler *OrderConstraintsOverridesHandler) Upsert(pair *model.TradingPair, override *model.OrderConstraintsOverride) {
	existingOverride, exists := ocHandler.overrides[pair.String()]
	if !exists {
		ocHandler.overrides[pair.String()] = override
		return
	}

	existingOverride.Augment(override)
	ocHandler.overrides[pair.String()] = existingOverride
}

// IsCompletelyOverriden returns true if the override exists and is complete for the given trading pair
func (ocHandler *OrderConstraintsOverridesHandler) IsCompletelyOverriden(pair *model.TradingPair) bool {
	override, has := ocHandler.overrides[pair.String()]
	if !has {
		return false
	}
	return override.IsComplete()
}
