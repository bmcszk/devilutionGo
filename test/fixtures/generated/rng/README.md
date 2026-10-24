# rng/ — LCG output vectors

One file per seed in `tools/probe/seeds.txt`, named `<seed>.bin`: 1000
little-endian `uint32_t` raw `GenerateRandomNumber()` outputs.

Consumers: devilGo-m2x7 (RNG port parity).

Produced by `dvg_probe rng` in the vsx3 probe run committed here.
