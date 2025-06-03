package chain_selectors

import (
	"strconv"
	"testing"
)

func TestCustomChainSelectorGeneration(t *testing.T) {
	// Test your custom chain IDs
	testChains := []uint64{9388201, 9250445, 1234567, 7777777}

	for _, chainID := range testChains {
		t.Run("Chain_"+strconv.FormatUint(chainID, 10), func(t *testing.T) {
			// Test selector generation
			selector1 := generateCustomChainSelector(chainID)
			selector2 := generateCustomChainSelector(chainID)

			// Should be deterministic
			if selector1 != selector2 {
				t.Errorf("Selector generation not deterministic for chain %d: %d != %d",
					chainID, selector1, selector2)
			}

			// Should be in custom range (high bit set)
			if !isCustomSelector(selector1) {
				t.Errorf("Generated selector %d for chain %d is not marked as custom",
					selector1, chainID)
			}

			// Test reverse lookup
			extractedChainID, err := extractChainIdFromCustomSelector(selector1)
			if err != nil {
				t.Errorf("Failed to extract chain ID from selector %d: %v", selector1, err)
			}

			if extractedChainID != chainID {
				t.Errorf("Extracted chain ID %d doesn't match original %d",
					extractedChainID, chainID)
			}
		})
	}
}

func TestGetChainDetailsByChainIDAndFamilyCustom(t *testing.T) {
	// Test your specific chain IDs
	testChains := []uint64{9388201, 9250445}

	for _, chainID := range testChains {
		t.Run("Details_"+strconv.FormatUint(chainID, 10), func(t *testing.T) {
			details, err := GetChainDetailsByChainIDAndFamily(
				strconv.FormatUint(chainID, 10),
				FamilyEVM,
			)

			if err != nil {
				t.Errorf("Failed to get details for custom chain %d: %v", chainID, err)
			}

			if details.ChainSelector == 0 {
				t.Errorf("Got zero selector for chain %d", chainID)
			}

			if details.ChainName == "" {
				t.Errorf("Got empty name for chain %d", chainID)
			}

			expectedName := generateCustomChainName(chainID)
			if details.ChainName != expectedName {
				t.Errorf("Name mismatch for chain %d: got %s, expected %s",
					chainID, details.ChainName, expectedName)
			}
		})
	}
}

func TestSelectorFromChainIdCustom(t *testing.T) {
	testChains := []uint64{9388201, 9250445}

	for _, chainID := range testChains {
		t.Run("Selector_"+strconv.FormatUint(chainID, 10), func(t *testing.T) {
			selector, err := SelectorFromChainId(chainID)

			if err != nil {
				t.Errorf("Failed to get selector for custom chain %d: %v", chainID, err)
			}

			if selector == 0 {
				t.Errorf("Got zero selector for chain %d", chainID)
			}

			// Should be marked as custom
			if !isCustomSelector(selector) {
				t.Errorf("Selector %d for chain %d is not marked as custom", selector, chainID)
			}
		})
	}
}

func TestGetCustomChainSelector(t *testing.T) {
	testChains := []uint64{9388201, 9250445, 1234567}

	for _, chainID := range testChains {
		t.Run("GetCustom_"+strconv.FormatUint(chainID, 10), func(t *testing.T) {
			selector, err := GetCustomChainSelector(chainID)

			if err != nil {
				t.Errorf("Failed to get custom selector for chain %d: %v", chainID, err)
			}

			if selector == 0 {
				t.Errorf("Got zero selector for chain %d", chainID)
			}

			// Verify consistency
			selector2, err2 := GetCustomChainSelector(chainID)
			if err2 != nil || selector != selector2 {
				t.Errorf("GetCustomChainSelector not consistent for chain %d", chainID)
			}
		})
	}
}

func TestOfficialChainsStillWork(t *testing.T) {
	// Test that official chains still work (e.g., Ethereum mainnet)
	selector, err := SelectorFromChainId(1) // Ethereum mainnet
	if err != nil {
		t.Errorf("Official Ethereum mainnet chain should still work: %v", err)
	}

	if selector == 0 {
		t.Errorf("Got zero selector for Ethereum mainnet")
	}
}

func BenchmarkCustomSelectorGeneration(b *testing.B) {
	chainID := uint64(9388201)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateCustomChainSelector(chainID)
	}
}
