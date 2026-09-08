# Generated API modernization (unreleased)

This change regenerates the SDK from the spec already checked into main at
61c389f. It does not introduce geometry-only session intent. That change belongs
in the stacked PR #453.

## Release compatibility

This is a breaking SDK update. Do not publish it as a compatible patch release.
Review the release/version policy and migrate consumers before publishing.
No version or tag is changed by this PR.

The current spec no longer contains these retired REST methods:

- `Ml.CreateTextToCad`
- `Ml.CreateTextToCadIteration`
- `Ml.CreateTextToCadMultiFileIteration`

Their generated types `TextToCad`, `TextToCadCreateBody`, `TextToCadIteration`,
`TextToCadIterationBody`, and `TextToCadMultiFileIteration` are also removed.
Move generation/iteration integrations to the Zookeeper Copilot WebSocket
protocol and its conversation lifecycle. This is not a drop-in return-type
replacement: callers must handle streamed messages, final files, and errors.
Do not substitute an empty result or retain a stub for a removed endpoint.

Two existing methods gain arguments:

- `Org.ListDatasets(limit, pageToken, lookupEnabled, sortBy)` adds
  `lookupEnabled` before `sortBy`. It sends an explicit boolean filter; choose
  the intended filter rather than mechanically inserting false.
- `Ml.CopilotWs(replay, conversationId, replayAttachmentMode, pr, body)` adds
  `replayAttachmentMode` before `pr`. The existing WebSocket generator does not
  forward these query arguments; this update does not claim to fix replay
  selection. Validate replay behavior separately before relying on the option.

## Stable Copilot messages and deprecations

The old generator selected names by property order and could reassign a public
name to another payload when optional properties were added. Copilot variants
now use the wire `type` discriminator, with the `MlCopilotMessage` prefix:
`Ping`, `ListModes`, `Headers`, `ProjectContext`, `User`, `FetchAttachments`,
`System`, and `AttachmentResponse`. Continue setting the `Type` field to the
corresponding protocol tag when sending a message.

Deprecated aliases preserve the meanings of every previously generated variant:

| Previous name | Replacement |
| --- | --- |
| `MlCopilotClientMessageCurrentFiles` | `MlCopilotMessageUser` |
| `MlCopilotClientMessageHeaders` | `MlCopilotMessageHeaders` |
| `MlCopilotClientMessageListModes` | `MlCopilotMessageListModes` |
| `MlCopilotClientMessageMlCopilotClientMessageHeaders` | `MlCopilotMessageProjectContext` |
| `MlCopilotClientMessagePing` | `MlCopilotMessagePing` |
| `MlCopilotClientMessageProjectContext` | `MlCopilotMessageAttachmentResponse` |
| `MlCopilotClientMessageProjectName` | `MlCopilotMessageSystem` |

`MlCopilotMode` becomes a deprecated alias for string, with the five existing
constants retained. Discover supported mode IDs with `list_modes`; a constant
does not guarantee that the server still offers that mode.

The aliases live outside generated output in `compat_copilot.go`, so regeneration
cannot silently turn a user message into a system message. Other union naming
behavior is unchanged.

## Other spec catch-up

The generated output adds user/organization Factory job lists and project
organization-assignment methods, project access/revision metadata, attachment
retrieval, Copilot access-denial and intermediate-checkpoint types, STEP target
representation, and updated feature enums. These additions follow the checked-in
spec; no remote service is changed or invoked by regeneration.

The generator template also preserves `NewClientFromEnv`'s existing invalid-host
error handling instead of dropping it during generation.

## Validation

Run `make generate` and verify a second run changes no output. Run `go test ./...`
and `go vet ./...`. The focused regressions exercise payload-field growth without
variant renaming, missing/duplicate tag rejection, retention of the input tag,
legacy user JSON serialization/deserialization, and old alias assignments.
Compile and migrate representative downstream REST callers before release;
these local tests do not exercise live Zookeeper, Factory, or Engine services.
