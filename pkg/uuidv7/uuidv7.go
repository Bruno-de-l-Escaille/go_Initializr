package uuidv7

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Advantages of UUIDv7:
// 1. Time-based: UUIDv7 includes a timestamp, which allows for chronological ordering of UUIDs. This is useful for event logging and time-series data.
// 2. Uniqueness: UUIDv7 ensures global uniqueness, making it suitable for distributed systems where unique identifiers are required without central coordination.
// 3. Compactness: When encoded in base62, UUIDv7 becomes more compact, reducing storage and transmission overhead compared to traditional UUID formats.
// 4. Readability: Base62 encoding produces a URL-safe and human-readable string, making it easier to use in web applications and APIs.
// 5. Compatibility: UUIDv7 maintains compatibility with existing UUID standards and libraries, ensuring seamless integration with existing systems and tools.

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Generate generates a UUID version 7 (time-sorted UUID)
func Generate() (string, error) {
	// Get current Unix timestamp in milliseconds
	now := time.Now()
	unixTsMs := uint64(now.UnixNano() / int64(time.Millisecond))

	// Create a 16-byte array for the UUID
	var uuidBytes [16]byte

	// Set timestamp (48 bits)
	binary.BigEndian.PutUint64(uuidBytes[0:8], unixTsMs<<16)

	// Generate 12 bytes of random data for the rest
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Copy random bytes to positions 6-15
	copy(uuidBytes[6:], randomBytes)

	// Set version (4 bits) - version 7
	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x70

	// Set variant (2 bits) - RFC 4122 variant
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80

	// Convert to UUID and return as string
	u, err := uuid.FromBytes(uuidBytes[:])
	if err != nil {
		return "", fmt.Errorf("failed to create UUID from bytes: %w", err)
	}

	return u.String(), nil
}

// MustGenerate generates a UUID version 7 and panics if it fails
func MustGenerate() string {
	id, err := Generate()
	if err != nil {
		panic(err)
	}
	return id
}

// IsValid checks if a string is a valid UUID v7
func IsValid(uuidStr string) bool {
	u, err := uuid.Parse(uuidStr)
	if err != nil {
		return false
	}

	// Check if it's version 7
	return u.Version() == 7
}

// ExtractTimestamp extracts the timestamp from a UUID v7
func ExtractTimestamp(uuidStr string) (time.Time, error) {
	u, err := uuid.Parse(uuidStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid UUID: %w", err)
	}

	if u.Version() != 7 {
		return time.Time{}, fmt.Errorf("UUID is not version 7")
	}

	// Extract timestamp from the first 48 bits
	uuidBytes := u[:]
	timestampMs := binary.BigEndian.Uint64(uuidBytes[0:8]) >> 16

	return time.Unix(0, int64(timestampMs)*int64(time.Millisecond)), nil
}

// EncodeBase62 encodes a byte slice to a base62 string
func EncodeBase62(input []byte) string {
	bi := new(big.Int).SetBytes(input)
	var result []byte
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := &big.Int{}

	for bi.Cmp(zero) != 0 {
		bi.DivMod(bi, base, mod)
		result = append(result, base62Chars[mod.Int64()])
	}

	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// DecodeBase62 decodes a base62 string to a byte slice
func DecodeBase62(input string, bytes int) ([]byte, error) {
	bi := big.NewInt(0)
	base := big.NewInt(62)

	for _, char := range input {
		index := int64(strings.IndexRune(base62Chars, char))
		bi.Mul(bi, base)
		bi.Add(bi, big.NewInt(index))
	}

	decoded := bi.Bytes()
	// Ensure the byte slice has the correct length
	if len(decoded) < bytes {
		padded := make([]byte, bytes-len(decoded))
		decoded = append(padded, decoded...)
	}

	return decoded, nil
}

// GenerateB62 generates a UUIDv7 and encodes it in base62
func GenerateB62() (string, error) {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	// Generate a UUIDv7
	uuidBytes, err := uuidV7.MarshalBinary()
	if err != nil {
		return "", err
	}

	// Encode bytes to base62
	base62UUID := EncodeBase62(uuidBytes)

	return base62UUID, nil
}

// GetTimestampFromBase62UUIDv7 extracts the timestamp from a base62 encoded UUIDv7
func GetTimestampFromBase62UUIDv7(base62UUID string) (time.Time, error) {
	uuidBytes, err := DecodeBase62(base62UUID, 16)
	if err != nil {
		return time.Time{}, err
	}

	// Convert the byte slice to a UUID
	var uuidV7 uuid.UUID
	copy(uuidV7[:], uuidBytes)

	// Extract the timestamp using the Time method
	timestamp, nsec := uuidV7.Time().UnixTime()

	return time.Unix(timestamp, nsec), nil
}

// Base62ToUUID converts a base62 encoded UUIDv7 to standard UUID string format
func Base62ToUUID(base62UUID string) (string, error) {
	uuidBytes, err := DecodeBase62(base62UUID, 16)
	if err != nil {
		return "", fmt.Errorf("failed to decode base62 UUID: %w", err)
	}

	// Convert the byte slice to a UUID
	var uuidV7 uuid.UUID
	copy(uuidV7[:], uuidBytes)

	return uuidV7.String(), nil
}

// UUIDToBase62 converts a standard UUID string to base62 encoded format
func UUIDToBase62(uuidStr string) (string, error) {
	u, err := uuid.Parse(uuidStr)
	if err != nil {
		return "", fmt.Errorf("invalid UUID: %w", err)
	}

	uuidBytes, err := u.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal UUID: %w", err)
	}

	return EncodeBase62(uuidBytes), nil
}
