# frames/ — timedemo frame dumps

One index/palette pair per frame from the scripted `demo_0.dmo` timedemo
run: `frame_<n>.idx.bin` (2×u16 LE dims then row-major 8-bit back buffer)
and `frame_<n>.pal.bin` (256 `SDL_Color` RGBA values).

Consumers: devilGo-7aul (render parity).

Produced by `dvg_probe frames` in the vsx3 probe run committed here.
