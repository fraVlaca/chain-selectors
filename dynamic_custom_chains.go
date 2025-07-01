package chain_selectors

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
)

// CUSTOM_CHAIN_RANGE defines the range for custom/testnet chains
// Any chain ID above this value will be treated as a custom chain
const CUSTOM_CHAIN_RANGE = uint64(1000000)

// No longer needed - using direct O(1) encoding/decoding
// Keeping imports for backward compatibility if needed

// deterministically generates a chain selector for any custom chain ID
func generateCustomChainSelector(chainID uint64) uint64 {
	// Use direct encoding with 0xE prefix for O(1) bidirectional transformation
	// This avoids collision with existing 0xD selectors and eliminates need for caching

	// Ensure chain ID fits in 60 bits (leaving 4 for 0xE marker)
	if chainID > 0x0FFFFFFFFFFFFFFF {
		// For very large chain IDs, fall back to hash-based approach
		hash := sha256.Sum256([]byte(fmt.Sprintf("custom-testnet-chain-%d", chainID)))
		selector := binary.BigEndian.Uint64(hash[:8])
		return 0xE000000000000000 | (selector & 0x0FFFFFFFFFFFFFFF)
	}

	// Direct encoding: 0xE prefix + chain ID (O(1) reversible)
	return 0xE000000000000000 | chainID
}

// generateCustomChainName creates a name for custom chains
func generateCustomChainName(chainID uint64) string {
	return fmt.Sprintf("custom-testnet-%d", chainID)
}

// isCustomChain determines if a chain ID should be treated as custom
func isCustomChain(chainID uint64) bool {
	// Check if it's not in official selectors (any non-official chain is custom)
	return !isInOfficialSelectors(chainID)
}

// isCustomSelector determines if a selector looks like a custom one
func isCustomSelector(selector uint64) bool {
	// Check if it has our custom 0xE prefix pattern
	return (selector & 0xF000000000000000) == 0xE000000000000000
}

// isInOfficialSelectors checks if chain ID exists in official selectors
func isInOfficialSelectors(chainID uint64) bool {
	_, exists := evmChainIdToChainSelector[chainID]
	return exists
}

// ExtractChainIdFromCustomSelector extracts chain ID from custom selector
func extractChainIdFromCustomSelector(selector uint64) (uint64, error) {
	if !isCustomSelector(selector) {
		return 0, fmt.Errorf("not a custom selector: %d", selector)
	}

	// Direct decoding: remove 0xE prefix to get chain ID (O(1) operation)
	chainID := selector & 0x0FFFFFFFFFFFFFFF

	// Verify the selector was generated with direct encoding
	// by checking if re-encoding produces the same selector
	if generateCustomChainSelector(chainID) == selector {
		return chainID, nil
	}

	// If verification fails, this selector might use the old hash-based method
	// or be from a very large chain ID that used hash fallback
	return 0, fmt.Errorf("could not reverse custom selector: %d (possibly hash-based)", selector)
}

// populateCommonCustomChains no longer needed with direct encoding
// Keeping function signature for backward compatibility
func populateCommonCustomChains() {
	// No-op: direct encoding eliminates need for pre-population
}

// Enhanced GetChainDetailsByChainIDAndFamily that supports custom chains
func GetChainDetailsByChainIDAndFamilyWithCustom(chainID string, family string) (ChainDetails, error) {
	// First try the standard function
	details, err := GetChainDetailsByChainIDAndFamily(chainID, family)
	if err == nil {
		return details, nil
	}

	// If not found, check if it's a custom chain
	if family == FamilyEVM {
		evmChainId, parseErr := strconv.ParseUint(chainID, 10, 64)
		if parseErr != nil {
			return ChainDetails{}, fmt.Errorf("invalid chain id %s for %s", chainID, family)
		}

		if isCustomChain(evmChainId) {
			// Generate deterministic selector for custom chain
			selector := generateCustomChainSelector(evmChainId)
			name := generateCustomChainName(evmChainId)

			// Check if custom chain support is enabled
			if os.Getenv("ENABLE_CUSTOM_CHAINS") != "false" {
				fmt.Printf("🔧 Generated custom chain selector: %s (ID: %d, Selector: %d)\n",
					name, evmChainId, selector)

				return ChainDetails{
					ChainSelector: selector,
					ChainName:     name,
				}, nil
			} else {
				fmt.Printf("⚠️  Custom chain %d detected but ENABLE_CUSTOM_CHAINS is disabled\n", evmChainId)
			}
		}
	}

	// Return original error if not a custom chain or custom chains disabled
	return ChainDetails{}, err
}

// Enhanced GetChainIDFromSelector that supports custom chains
func GetChainIDFromSelectorWithCustom(selector uint64) (string, error) {
	// First try the standard function
	chainID, err := GetChainIDFromSelector(selector)
	if err == nil {
		return chainID, nil
	}

	// Check if it's a custom selector
	if isCustomSelector(selector) {
		evmChainId, extractErr := extractChainIdFromCustomSelector(selector)
		if extractErr == nil {
			return strconv.FormatUint(evmChainId, 10), nil
		}
	}

	// Return original error if not custom
	return "", err
}

// RegisterCustomChain manually registers a custom chain for immediate use
func RegisterCustomChain(chainID uint64, name string) uint64 {
	selector := generateCustomChainSelector(chainID)

	fmt.Printf("✅ Registered custom chain: %s (ID: %d, Selector: %d)\n",
		name, chainID, selector)

	return selector
}

// GetCustomChainSelector is the main function to get selector for any chain
func GetCustomChainSelector(chainID uint64) (uint64, error) {
	// First check if it's in official selectors
	if details, exists := evmChainIdToChainSelector[chainID]; exists {
		return details.ChainSelector, nil
	}

	// Generate deterministic selector for custom chains
	if isCustomChain(chainID) {
		if os.Getenv("ENABLE_CUSTOM_CHAINS") != "false" {
			selector := generateCustomChainSelector(chainID)
			name := generateCustomChainName(chainID)

			fmt.Printf("🔧 Generated custom chain selector: %s (ID: %d, Selector: %d)\n",
				name, chainID, selector)

			return selector, nil
		} else {
			return 0, fmt.Errorf("custom chain %d detected but ENABLE_CUSTOM_CHAINS is disabled", chainID)
		}
	}

	return 0, fmt.Errorf("chain selector not found for chain %d", chainID)
}

// ListAllChains returns both official and custom chains in a range
func ListAllChains(startChainID, endChainID uint64) []ChainDetails {
	var chains []ChainDetails

	// Add official chains in range
	for chainID, details := range evmChainIdToChainSelector {
		if chainID >= startChainID && chainID <= endChainID {
			chains = append(chains, details)
		}
	}

	// Add custom chains in range (if enabled)
	if os.Getenv("ENABLE_CUSTOM_CHAINS") != "false" {
		for chainID := startChainID; chainID <= endChainID; chainID++ {
			if isCustomChain(chainID) && !isInOfficialSelectors(chainID) {
				selector := generateCustomChainSelector(chainID)
				name := generateCustomChainName(chainID)

				chains = append(chains, ChainDetails{
					ChainSelector: selector,
					ChainName:     name,
				})
			}
		}
	}

	return chains
}
