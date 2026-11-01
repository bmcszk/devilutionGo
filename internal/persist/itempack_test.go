package persist_test

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// itemPackRow is one row of testdata/item_pack_golden.tsv.
type itemPackRow struct {
	suite string
	pack  persist.ItemPack
	label string
}

func TestItemPackSize(t *testing.T) {
	// when
	size := persist.ItemPackSize
	// then the layout is pack.h:20-31: 4+2+2+1+1+1+1+1+2+4
	if size != 19 {
		t.Fatalf("ItemPackSize = %d, want 19", size)
	}
}

func TestItemPackGoldenRoundTrip(t *testing.T) {
	for _, row := range loadItemPackGolden(t) {
		t.Run(row.suite+"/"+row.label, func(t *testing.T) {
			// when the packed_test.cpp vector is encoded and decoded back
			err := assertItemPackRoundTrip(row.pack)
			// then the byte image and every field survive
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestItemPackEncodeEarVector(t *testing.T) {
	// given the ItemPack from pack_test.cpp:880 (UnPackItem_ear)
	pack := persist.ItemPack{
		Seed:       1633955154,
		CreateInfo: 17509,
		Idx:        23,
		BId:        111,
		BDur:       103,
		BMDur:      117,
		BCh:        101,
		BMCh:       68,
		Value:      19843,
	}
	// when
	encoded := hex.EncodeToString(pack.Encode())
	// then every multi-byte field is little-endian (endian_swap.hpp:27-49)
	const want = "522d6461654417006f67756544834d00000000"
	if encoded != want {
		t.Fatalf("Encode = %s, want %s", encoded, want)
	}
}

func TestItemPackEncodeEmptyVector(t *testing.T) {
	// given the empty ItemPack from pack_test.cpp:818 (UnPackItem_empty)
	pack := persist.ItemPack{Idx: persist.PackedItemNone}
	// when
	encoded := hex.EncodeToString(pack.Encode())
	// then idx 0xFFFF marks the empty slot, stored little-endian
	const want = "000000000000ffff0000000000000000000000"
	if encoded != want {
		t.Fatalf("Encode = %s, want %s", encoded, want)
	}
}

func TestDecodeItemPackShortBuffer(t *testing.T) {
	// given a truncated record
	buf := make([]byte, persist.ItemPackSize-1)
	// when
	_, err := persist.DecodeItemPack(buf)
	// then the decode fails instead of reading past the end
	if err == nil {
		t.Fatal("DecodeItemPack(short) = nil error, want error")
	}
}

// assertItemPackRoundTrip checks one encoded record against its fields.
func assertItemPackRoundTrip(pack persist.ItemPack) error {
	encoded := pack.Encode()
	decoded, err := persist.DecodeItemPack(encoded)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	if len(encoded) != persist.ItemPackSize {
		return fmt.Errorf("encode len = %d, want %d", len(encoded), persist.ItemPackSize)
	}

	if decoded != pack {
		return fmt.Errorf("round trip = %+v, want %+v", decoded, pack)
	}

	return nil
}

// loadItemPackGolden reads the committed pack_test.cpp vectors.
func loadItemPackGolden(t *testing.T) []itemPackRow {
	t.Helper()

	file, err := os.Open("testdata/item_pack_golden.tsv")
	if err != nil {
		t.Fatalf("open golden: %v", err)
	}

	defer file.Close() //nolint:errcheck // read-only fixture

	var rows []itemPackRow

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		rows = append(rows, parseItemPackRow(t, line))
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan golden: %v", err)
	}

	if len(rows) < 100 {
		t.Fatalf("golden has %d rows, want at least 100", len(rows))
	}

	return rows
}

// parseItemPackRow converts one tab-separated golden line.
func parseItemPackRow(t *testing.T, line string) itemPackRow {
	t.Helper()

	fields := strings.Split(line, "\t")
	if len(fields) != 12 {
		t.Fatalf("golden row has %d fields, want 12: %q", len(fields), line)
	}

	nums := make([]uint64, 10)
	for i := range nums {
		value, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
			t.Fatalf("parse field %d of %q: %v", i, line, err)
		}

		nums[i] = value
	}

	return itemPackRow{
		suite: fields[0],
		pack: persist.ItemPack{
			Seed:       uint32(nums[0]),
			CreateInfo: uint16(nums[1]),
			Idx:        uint16(nums[2]),
			BId:        uint8(nums[3]),
			BDur:       uint8(nums[4]),
			BMDur:      uint8(nums[5]),
			BCh:        uint8(nums[6]),
			BMCh:       uint8(nums[7]),
			Value:      uint16(nums[8]),
			Buff:       uint32(nums[9]),
		},
		label: fields[11],
	}
}
