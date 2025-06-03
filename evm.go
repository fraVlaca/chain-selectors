package chain_selectors

import (
	_ "embed"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

//go:generate go run genchains_evm.go

//go:embed selectors.yml
var selectorsYml []byte

//go:embed test_selectors.yml
var testSelectorsYml []byte

type ChainDetails struct {
	ChainSelector uint64 `yaml:"selector"`
	ChainName     string `yaml:"name"`
}

var (
	evmSelectorsMap           = parseYml(selectorsYml)
	evmTestSelectorsMap       = parseYml(testSelectorsYml)
	evmChainIdToChainSelector = loadAllEVMSelectors()
	evmChainsBySelector       = make(map[uint64]Chain)
	evmChainsByEvmChainID     = make(map[uint64]Chain)
)

func init() {
	for _, ch := range ALL {
		evmChainsBySelector[ch.Selector] = ch
		evmChainsByEvmChainID[ch.EvmChainID] = ch
	}
}

func loadAllEVMSelectors() map[uint64]ChainDetails {
	output := make(map[uint64]ChainDetails, len(evmSelectorsMap)+len(evmTestSelectorsMap))
	for k, v := range evmSelectorsMap {
		output[k] = v
	}
	for k, v := range evmTestSelectorsMap {
		output[k] = v
	}
	return output
}

func parseYml(ymlFile []byte) map[uint64]ChainDetails {
	type ymlData struct {
		SelectorsByEvmChainId map[uint64]ChainDetails `yaml:"selectors"`
	}

	var data ymlData
	err := yaml.Unmarshal(ymlFile, &data)
	if err != nil {
		panic(err)
	}

	return data.SelectorsByEvmChainId
}

func EvmChainIdToChainSelector() map[uint64]uint64 {
	copyMap := make(map[uint64]uint64, len(evmChainIdToChainSelector))
	for k, v := range evmChainIdToChainSelector {
		copyMap[k] = v.ChainSelector
	}
	return copyMap
}

// Deprecated, this only supports EVM chains, use the chain agnostic `GetChainIDFromSelector` instead
func ChainIdFromSelector(chainSelectorId uint64) (uint64, error) {
	for k, v := range evmChainIdToChainSelector {
		if v.ChainSelector == chainSelectorId {
			return k, nil
		}
	}

	// Try custom selector lookup
	if isCustomSelector(chainSelectorId) {
		return extractChainIdFromCustomSelector(chainSelectorId)
	}

	return 0, fmt.Errorf("chain not found for chain selector %d", chainSelectorId)
}

// Deprecated, this only supports EVM chains, use the chain agnostic `GetChainDetailsByChainIDAndFamily` instead
// ENHANCED: Now supports custom chains with deterministic generation
func SelectorFromChainId(chainId uint64) (uint64, error) {
	if chainSelectorId, exist := evmChainIdToChainSelector[chainId]; exist {
		return chainSelectorId.ChainSelector, nil
	}

	// Try our custom chain selector generation
	return GetCustomChainSelector(chainId)
}

// Deprecated, this only supports EVM chains, use the chain agnostic `NameFromChainId` instead
func NameFromChainId(chainId uint64) (string, error) {
	details, exist := evmChainIdToChainSelector[chainId]
	if !exist {
		// Try custom chain name generation
		if isCustomChain(chainId) {
			return generateCustomChainName(chainId), nil
		}
		return "", fmt.Errorf("chain name not found for chain %d", chainId)
	}
	if details.ChainName == "" {
		return strconv.FormatUint(chainId, 10), nil
	}
	return details.ChainName, nil
}

func ChainIdFromName(name string) (uint64, error) {
	for k, v := range evmChainIdToChainSelector {
		if v.ChainName == name {
			return k, nil
		}
	}
	chainId, err := strconv.ParseUint(name, 10, 64)
	if err == nil {
		if details, exist := evmChainIdToChainSelector[chainId]; exist && details.ChainName == "" {
			return chainId, nil
		}
		// ENHANCED: Check if it's a custom chain
		if isCustomChain(chainId) {
			return chainId, nil
		}
	}
	return 0, fmt.Errorf("chain not found for name %s", name)
}

func TestChainIds() []uint64 {
	chainIds := make([]uint64, 0, len(evmTestSelectorsMap))
	for k := range evmTestSelectorsMap {
		chainIds = append(chainIds, k)
	}
	return chainIds
}

// ENHANCED: Now supports custom chains
func ChainBySelector(sel uint64) (Chain, bool) {
	ch, exists := evmChainsBySelector[sel]
	if exists {
		return ch, true
	}

	// Try custom selector lookup
	if isCustomSelector(sel) {
		chainID, err := extractChainIdFromCustomSelector(sel)
		if err == nil {
			// Create a synthetic Chain for custom chains
			return Chain{
				EvmChainID: chainID,
				Selector:   sel,
				Name:       generateCustomChainName(chainID),
				VarName:    fmt.Sprintf("CUSTOM_TESTNET_%d", chainID),
			}, true
		}
	}

	return Chain{}, false
}

// ENHANCED: Now supports custom chains
func ChainByEvmChainID(evmChainID uint64) (Chain, bool) {
	ch, exists := evmChainsByEvmChainID[evmChainID]
	if exists {
		return ch, true
	}

	// Try custom chain lookup
	if isCustomChain(evmChainID) {
		selector := generateCustomChainSelector(evmChainID)
		name := generateCustomChainName(evmChainID)

		return Chain{
			EvmChainID: evmChainID,
			Selector:   selector,
			Name:       name,
			VarName:    fmt.Sprintf("CUSTOM_TESTNET_%d", evmChainID),
		}, true
	}

	return Chain{}, false
}

// ENHANCED: Now supports custom chains
func IsEvm(chainSel uint64) (bool, error) {
	_, exists := ChainBySelector(chainSel)
	if !exists {
		// Check if it's a custom selector
		if isCustomSelector(chainSel) {
			return true, nil
		}
		return false, fmt.Errorf("chain %d not found", chainSel)
	}
	// We always return true since only evm chains are supported atm.
	return true, nil
}
