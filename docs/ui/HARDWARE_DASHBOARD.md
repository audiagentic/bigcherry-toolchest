# Hardware dashboard

The left operator rail separates navigation from hardware telemetry.

## Behaviour

- Navigation is fixed and compact.
- The hardware region scrolls independently only when required.
- GPUs render as a two-column card dashboard. Opening a card shows full identity details.
- Initial markup is server-rendered from `/api/monitor`.
- Dynamic values are updated in place from `/api/monitor/stream` SSE; the dashboard is not replaced every three seconds.

## GPU identity

Dynamic metrics and static hardware identity are deliberately separate.

On ROCm, dynamic utilisation/temperature/power remains collected through the normal monitor backend. Stable identity is discovered once per Toolchest process and keyed to llama.cpp's `ROCm<N>` ordering by PCI BDF. `rocm-smi --showbus --showproductname --csv` supplies the product name where available, with the existing monitor name as fallback.

This avoids treating `cardN` or an independently enumerated agent index as the identity of `ROCm<N>` on multi-GPU hosts.
