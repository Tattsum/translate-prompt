# Test fixtures

## `intake/`

Analyze integration tests (`application/intake/integration_test.go`).

| File | Purpose |
|------|---------|
| `heuristic_ambiguous.prompt.md` | Missing goal/scope/acceptance → `needs_input` |
| `heuristic_complete.prompt.md` | Complete prompt → `ready` |

## `optimize/`

Optimize / compress pipeline integration tests (`application/optimize/integration_test.go`).

| File | Purpose |
|------|---------|
| `llm_cursor_rules.prompt.md` | Full optimize path; format-stage `cursor-actionable` |
| `llm_claude_task.prompt.md` | Claude task with imperative residual (compress fixture) |
| `llm_common_examples_body.md` | Long examples body for `common-example-summarize` |
| `verbose_prompt.md` | Phase 1 profile structure regression (parent dir) |

LLM refiner rule tests use **post-format Section trees** plus `ExportBuildCompressPipeline` so compress-stage LLM rules are exercised without format-stage pattern replacement clearing triggers.
