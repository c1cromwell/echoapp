---
name: echo-knowledge-graph
description: >-
  Rebuilds and consults Echo's living traceability graph from Software Factory
  (factory.8090.ai) work orders through blueprints/requirements to code and
  2-device E2E. Use after WO sync, Factory CSV export, gap-audit edits, or when
  the user asks about 8090, requirements mapping, or knowledge graph updates.
---

# Echo knowledge graph (→ factory.8090.ai)

## Outputs

| Path | Use |
|------|-----|
| [`docs/knowledge-graph/graph.json`](../../docs/knowledge-graph/graph.json) | Full WO/blueprint/feature graph + Factory URLs |
| [`docs/knowledge-graph/TRACEABILITY.md`](../../docs/knowledge-graph/TRACEABILITY.md) | Human tables |
| [`docs/E2E_TWO_DEVICE.md`](../../docs/E2E_TWO_DEVICE.md) | Two-device test playbook |

## Rebuild (always after WO/doc changes)

```bash
make knowledge-graph
```

Script: `scripts/knowledge-graph/build.py`.

## Source hierarchy

1. Live Factory MCP `software-factory-echo` (`list_work_orders`, `list_requirements`, `list_blueprints`) when authenticated
2. `docs/phase-*-work-orders.md` status sections
3. `docs/echo-work-orders-*.csv` (WO UUID + URL)
4. `FEATURES` overlay in `build.py` for **code reality**

Requirement Titles in the CSV are empty — treat **blueprint title = requirements section**.

## When Factory MCP is available

1. `list_phases` + paginated `list_work_orders` (`page_size` 100)
2. Reconcile status via `echo-work-order-sync` (do not mark `in_review` unless asked)
3. If CSV is stale vs live IDs, tell the user to re-export from Factory
4. `make knowledge-graph`
5. Optionally `create_resource` in Factory Knowledge Base with a short TRACEABILITY summary — only if the user asked to push **to** 8090

## Do not

- Load `Echo_Combined_Requirements.md` by default
- Invent Factory statuses
- Put message plaintext, phrases, or APNs bodies in the graph
