#pragma once
// Redirects items.cpp's GenerateRnd call sites to the probe's recording wrapper
// (defined in probe.cpp) so the probe can observe the pool size the C++ hands to
// GenerateRnd without touching the devilutionX tree.
#define GenerateRnd probeGenerateRnd
