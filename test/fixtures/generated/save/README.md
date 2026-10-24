# save/ — vanilla spawn save archive

`spawn_0.sv`: vanilla spawn `.sv` hero archive written by the scripted run
(`pfile_write_hero(true)`); new-player `PlayerPack` timestamp fields
(`pack.h:34-35`) are zeroed by `PackPlayer` initialization, so the pin
holds.

Consumers: devilGo-6xu2 / devilGo-oqho (save roundtrip).

Produced by `dvg_probe save` in the vsx3 probe run committed here.
