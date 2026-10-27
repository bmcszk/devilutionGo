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
	// given
	tests := []struct {
		name   string
		size   func() uintptr
		pinned uintptr
	}{
		{"Player", func() uintptr { return unsafe.Sizeof(entities.Player{}) }, 19712},
		{"Monster", func() uintptr { return unsafe.Sizeof(entities.Monster{}) }, 224},
		{"Object", func() uintptr { return unsafe.Sizeof(entities.Object{}) }, 144},
		{"Item", func() uintptr { return unsafe.Sizeof(entities.Item{}) }, 336},
		{"ActorPosition", func() uintptr { return unsafe.Sizeof(entities.ActorPosition{}) }, 80},
		{"SpellCastInfo", func() uintptr { return unsafe.Sizeof(entities.SpellCastInfo{}) }, 16},
	}
	// when
	type measurement struct {
		name string
		got  uintptr
		want uintptr
	}
	measurements := make([]measurement, len(tests))
	for i, tt := range tests {
		measurements[i] = measurement{name: tt.name, got: tt.size(), want: tt.pinned}
	}
	// then
	for _, m := range measurements {
		t.Run(m.name, func(t *testing.T) {
			if m.got != m.want {
				t.Fatalf("size drift: %s is %d bytes, pinned %d — update the table deliberately or revert the change",
					m.name, m.got, m.want)
			}
		})
	}
}
