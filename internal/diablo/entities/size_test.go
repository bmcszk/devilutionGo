package entities_test

import (
	"testing"
	"unsafe"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
)

// TestStructSizesPinned documents each ported struct's Go size at the time
// of the port. Go sizes do NOT match the C++ ABI (different alignment, no
// packed Item arrays, GC pointers); the pin is a change-detector: any field
// addition or type change that moves a size must update the table here
// deliberately, with a note in the bean's PoW. To re-pin: print
// unsafe.Sizeof of the new layout and update the expected value.
func TestStructSizesPinned(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Player", unsafe.Sizeof(entities.Player{}), 19712},
		{"Monster", unsafe.Sizeof(entities.Monster{}), 224},
		{"Object", unsafe.Sizeof(entities.Object{}), 144},
		{"Item", unsafe.Sizeof(entities.Item{}), 336},
		{"ActorPosition", unsafe.Sizeof(entities.ActorPosition{}), 80},
		{"SpellCastInfo", unsafe.Sizeof(entities.SpellCastInfo{}), 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("size drift: %s is %d bytes, pinned %d — update the table deliberately or revert the change",
					tt.name, tt.got, tt.want)
			}
		})
	}
}
