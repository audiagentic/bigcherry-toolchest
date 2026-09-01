# BigCherry Toolchest integration

This fork keeps BigCherry responsible for producing tuned llama.cpp binaries and uses Toolchest as the launcher, experiment manager, benchmark matrix runner, history store, and comparison UI.

## External build registration

An external build is registered into Toolchest's native `builds.json` registry. Toolchest does not copy or own the supplied binary. Existing build selection, router lifecycle, benchmark jobs, and benchmark `BuildSnapshot` capture then work without a separate BigCherry code path.

Use the import command from this repository:

    go run ./cmd/toolchest-import-build \
      --data-dir /path/to/toolchest-data \
      --id bigcherry-b10227-hip \
      --profile rocm \
      --binary /path/to/bigcherry-b10227-hip/llama-server \
      --git-ref b10227 \
      --git-sha <bigcherry-commit> \
      --tag bigcherry-mmq

`--binary` may point directly to `llama-server` or to a directory containing `llama-server`.

Use `--replace` to update an existing registry entry with the same ID.

Restart Toolchest after importing so the running service reloads `builds.json`.

## Provenance convention

Until upstream Toolchest gains first-class arbitrary-source metadata, use the native build fields consistently:

- `ID`: stable lane/build identity, e.g. `bigcherry-b10227-hip-mmq`
- `Profile`: Toolchest backend/profile, normally `rocm` or `vulkan`
- `GitSHA`: exact BigCherry commit used to produce the binary
- `GitRef`: upstream/base ref or BigCherry branch used for the run
- `Tag`: short variant label, e.g. `mmq-rccl`, `cublas-rocwmma`, `baseline`

These fields are already copied into benchmark `BuildSnapshot`, so historical benchmark results retain the build identity even if the build registry later changes.

## A/B job shape

Toolchest benchmark jobs already expand `models x builds x presets x sweeps`. Register each upstream/BigCherry binary as a distinct build ID, then select them in one job.

Typical matrix:

- upstream HIP baseline
- BigCherry HIP baseline
- BigCherry HIP MMQ
- upstream Vulkan
- BigCherry Vulkan

Then sweep model parameters such as tensor split, GPU assignment, KV type, batch/ubatch, and `draft-mtp` depth through Toolchest's existing job system.

## Deletion safety

Deleting an imported build from Toolchest unregisters the build. Toolchest's delete path removes only its own `dataDir/builds/<id>` directory; the registered external executable is not removed or modified.

## Upstream maintenance

Keep BigCherry-specific changes small. Generic functionality suitable for Toolchest should be proposed upstream. This fork includes a scheduled sync workflow that merges `tmac1973/llama-toolchest:main` into a sync branch and opens/updates a PR against this fork's `main`; it never auto-merges upstream changes.

## Next integration slices

1. Add a Web UI/API form for external build registration.
2. Promote arbitrary source repository/upstream-base/lane fields to first-class `BuildResult`/`BuildSnapshot` metadata and propose that generic model upstream.
3. Add SPEED-Bench as a benchmark adapter for MTP acceptance/speculative decoding characterization.
4. Optionally add GuideLLM/k6 adapters for sustained server load and saturation tests.
