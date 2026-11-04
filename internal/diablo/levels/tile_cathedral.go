package levels

// Cathedral (drlg_l1.cpp:97-178) tile vocabulary (upstream `enum Tile : uint8_t` in the
// anonymous namespace of that translation unit). Names carry the prefix
// 'cat' because Go has one package namespace where C++ had one per TU.
//
//nolint:stylecheck // upstream names, verbatim apart from the prefix
const (
	catVWall          Tile = 1
	catHWall          Tile = 2
	catCorner         Tile = 3
	catDWall          Tile = 4
	catDArch          Tile = 5
	catVWallEnd       Tile = 6
	catHWallEnd       Tile = 7
	catHArchEnd       Tile = 8
	catVArchEnd       Tile = 9
	catHArchVWall     Tile = 10
	catVArch          Tile = 11
	catHArch          Tile = 12
	catFloor          Tile = 13
	catHWallVArch     Tile = 14
	catPillar         Tile = 15
	catVCorner        Tile = 16
	catHCorner        Tile = 17
	catDirtHwall      Tile = 18
	catDirtVwall      Tile = 19
	catVDirtCorner    Tile = 20
	catHDirtCorner    Tile = 21
	catDirt           Tile = 22
	catDirtHwallEnd   Tile = 23
	catDirtVwallEnd   Tile = 24
	catVDoor          Tile = 25
	catHDoor          Tile = 26
	catHFenceVWall    Tile = 27
	catHDoorVDoor     Tile = 28
	catDFence         Tile = 29
	catVDoorEnd       Tile = 30
	catHDoorEnd       Tile = 31
	catVFenceEnd      Tile = 32
	catVArchEnd2      Tile = 33
	catHArchVWall2    Tile = 34
	catVFence         Tile = 35
	catHFence         Tile = 36
	catHWallVFence    Tile = 37
	catHArchVFence    Tile = 38
	catHArchVDoor     Tile = 39
	catHArchVWall3    Tile = 40
	catDWall2         Tile = 41
	catHWallVArch2    Tile = 42
	catDWall3         Tile = 43
	catEntranceStairs Tile = 64
	catVWall2         Tile = 79
	catHWall2         Tile = 80
	catDWall4         Tile = 82
	catVWallEnd2      Tile = 84
	catVWall4         Tile = 89
	catVWall5         Tile = 90
	catHWall4         Tile = 91
	catHWall5         Tile = 92
	catVWall8         Tile = 100
	catFloor12        Tile = 139
	catFloor13        Tile = 140
	catFloor14        Tile = 141
	catFloor15        Tile = 142
	catFloor16        Tile = 143
	catFloor17        Tile = 144
	catFloor18        Tile = 145
	catVWall17        Tile = 146
	catVArch5         Tile = 147
	catHWallShadow    Tile = 148
	catHArchShadow    Tile = 149
	catFloor19        Tile = 150
	catFloor20        Tile = 151
	catFloor21        Tile = 152
	catHArchShadow2   Tile = 153
	catHWallShadow2   Tile = 154
	catFloor22        Tile = 162
	catFloor23        Tile = 163
	catDirtHWall2     Tile = 199
	catDirtVWall2     Tile = 200
	catDirtCorner2    Tile = 202
	catDirtHWallEnd2  Tile = 204
	catDirtVWallEnd2  Tile = 205
)
