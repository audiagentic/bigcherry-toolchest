# BigCherry Toolchest integration

This fork keeps BigCherry responsible for producing formal llama.cpp comparison builds and uses Toolchest as the launcher, experiment manager, benchmark matrix runner, history store, and comparison UI.

## Live external build registration

An external build is registered into Toolchest's native build registry. Toolchest does not copy or own the supplied binary. Existing build selection, router lifecycle, benchmark jobs, and benchmark `BuildSnapshot` capture work without a separate BigCherry execution path.

The live endpoint is:

    POST /api/builds?external=1

JSON fields:

- `id`: stable build identity
- `profile`: Toolchest build profile (`rocm`, `vulkan`, etc.)
- `binary_path`: `llama-server` executable or directory containing it
- `git_ref`: source/upstream revision
- `git_sha`: producer revision
- `tag`: optional human/identity label
- `replace`: replace the same build ID when true

A successful request updates the same in-memory Builder used by router selection and benchmark jobs and persists the entry to `builds.json`. The build is immediately available; Toolchest does not need to restart.

The Builds page exposes the same operation under **Register Existing Build (BigCherry / external)**.

## Automatic BigCherry handoff

BigCherry can publish completed canonical builds directly by setting:

    BIGCHERRY_TOOLCHEST_URL=http://127.0.0.1:3000

and then running its normal build workflow, for example:

    bigcherry build --profile standard

With publication enabled, BigCherry produces `llama-server` plus `llama-bench` from the same content-addressed BuildPlan, publishes them into the same verified runtime bundle, and registers the immutable ArtifactStore `llama-server` path here. Toolchest then owns launch/benchmark state while BigCherry remains the artifact authority.

Toolchest and BigCherry must see the registered absolute path. If Toolchest runs in a container, bind-mount the BigCherry ArtifactStore/work root at the same absolute path inside the container.

## Offline importer fallback

The command-line importer remains useful when Toolchest is not running:

    go run ./cmd/toolchest-import-build \
      --data-dir /path/to/toolchest-data \
      --id bigcherry-b10227-hip \
      --profile rocm \
      --binary /path/to/bigcherry-b10227-hip/llama-server \
      --git-ref b10227 \
      --git-sha <bigcherry-commit> \
      --tag bigcherry-mmq

`--binary` may point directly to `llama-server` or to a directory containing `llama-server`. Use `--replace` to update an existing registry entry with the same ID. Because this fallback edits persisted registry state rather than the running Builder instance, restart Toolchest after using the offline importer.

## Provenance convention

For BigCherry-published builds:

- `ID`: `<source>-<build>-<platform>-<full BigCherry build_plan_id>`
- `Profile`: Toolchest backend/profile; BigCherry `hip` maps to `rocm`
- `GitRef`: exact resolved upstream llama.cpp revision
- `GitSHA`: exact BigCherry producer revision
- `Tag`: short effective-build/runtime-bundle identity (`eb-...-rb-...`)
- `BinaryPath`: immutable BigCherry ArtifactStore `llama-server`

These native fields are already copied into benchmark `BuildSnapshot`, so historical benchmark results retain the registered build identity even if the live registry later changes. Detailed source-slice, patch, input, toolchain and runtime-bundle provenance remains authoritative in BigCherry's immutable artifact store.

## Formal A/B rule

Build both stock upstream and patched candidates through BigCherry for formal comparisons. This keeps compiler/toolchain, backend stack, requested targets and build environment under one identity system; the code/patch selection is then the intended difference rather than an accidental second build pipeline.

Toolchest's native llama.cpp builder remains available for convenience and exploratory work, but it should not be used as the baseline side of formal BigCherry A/B results.

Toolchest benchmark jobs already expand `models x builds x presets x sweeps`. Register each upstream/BigCherry runtime bundle as a distinct Build ID, then select them in one job and sweep tensor split, GPU assignment, KV type, batch/ubatch and `draft-mtp` settings through the existing job system.

## Deletion safety

Deleting an imported build from Toolchest unregisters the build. Toolchest's delete path removes only its own `dataDir/builds/<id>` directory; the registered BigCherry ArtifactStore executable and runtime bundle are not removed or modified.

## Upstream maintenance

Keep BigCherry-specific changes small. Generic functionality suitable for Toolchest should be proposed upstream. This fork includes a scheduled sync workflow that merges `tmac1973/llama-toolchest:main` into a sync branch and opens/updates a PR against this fork's `main`; it never auto-merges upstream changes.

## Next benchmark integrations

The next benchmark adapter should be SPEED-Bench for MTP/speculative-decoding acceptance and workload characterization. GuideLLM/k6 remain optional additions for sustained server load and saturation testing.
