package storage

import (
	"crypto/sha256"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// ComputeCID generates a deterministic CIDv1 from raw bytes using SHA-256.
// This produces the same CID that IPFS would generate for the content,
// allowing pre-upload verification and deduplication.
func ComputeCID(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("cid: cannot compute CID for empty data")
	}

	// SHA-256 hash of the content.
	hash := sha256.Sum256(data)

	// Encode as a multihash (SHA2-256, 32 bytes).
	mh, err := multihash.Encode(hash[:], multihash.SHA2_256)
	if err != nil {
		return "", fmt.Errorf("cid: failed to encode multihash: %w", err)
	}

	// Build CIDv1 with raw codec (0x55).
	c := cid.NewCidV1(cid.Raw, mh)
	return c.String(), nil
}

// ValidateCID checks that a CID string is well-formed and decodable.
// Returns the decoded CID string (normalized) or an error.
func ValidateCID(cidStr string) (string, error) {
	if cidStr == "" {
		return "", fmt.Errorf("cid: empty CID")
	}

	c, err := cid.Decode(cidStr)
	if err != nil {
		return "", fmt.Errorf("cid: invalid format: %w", err)
	}

	return c.String(), nil
}

// VerifyContent checks that a CID matches the expected content.
// Computes the CID of the data and compares it to the provided CID.
// Returns true if they match (content integrity verified).
func VerifyContent(cidStr string, data []byte) (bool, error) {
	computed, err := ComputeCID(data)
	if err != nil {
		return false, err
	}

	// Normalize both CIDs for comparison.
	expected, err := ValidateCID(cidStr)
	if err != nil {
		return false, err
	}

	return computed == expected, nil
}
