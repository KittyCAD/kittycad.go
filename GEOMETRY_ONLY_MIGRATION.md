# Geometry-only modeling sessions (proposed)

The generated `Modeling.CommandsWs` method gains a `geometryOnly bool`
argument immediately before `pr`. Existing callers must pass `false` there
to retain rendering intent. Geometry-only callers pass `true` and set the
existing `webrtc` argument to `false`. Treat this positional signature change
as a breaking-release gate; do not silently publish it as a compatible patch.

The WebSocket template previously discarded all method query arguments.
This scoped change forwards `geometry_only` and `webrtc` explicitly, including
false. Other previously ignored WebSocket arguments remain outside this fix.
It does not synthesize SSAO or video dimensions.

Server support must be deployed before enabling geometry-only callers.
CPU allocation also depends on server routing and rollout configuration;
this option specifies capability, not guaranteed hardware. Keep image,
snapshot and visual-test workloads on the rendered path. Require paired
geometry/export parity and verified CPU routing before migrating CI broadly.
