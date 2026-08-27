# Echo knowledge graph

Living traceability from **[Software Factory (factory.8090.ai)](https://factory.8090.ai)** work orders through **blueprints / requirements** to **code** and **2-device E2E**.

## Rebuild

```bash
make knowledge-graph
```

Writes:

| File | Role |
|------|------|
| [`graph.json`](graph.json) | Full graph (WO nodes include Factory URLs) |
| [`TRACEABILITY.md`](TRACEABILITY.md) | Human tables |

A Cursor hook (`.cursor/hooks/refresh-knowledge-graph.sh`) rebuilds after edits to phase WO docs, the Factory CSV, gap audits, or this folder.

## Source of truth

1. **Software Factory** at factory.8090.ai — live WO status (MCP `software-factory-echo`)
2. **`docs/phase-*-work-orders.md`** — last exported status sections
3. **`docs/echo-work-orders-*.csv`** — WO UUIDs + Factory URLs + blueprint titles
4. **FEATURES overlay** in `scripts/knowledge-graph/build.py` — code reality for testers

Requirement Titles in the CSV are empty. In Echo, each **blueprint title is the requirements section of the same name**.

## After reconnecting Factory MCP

1. `list_phases` / `list_work_orders` / `list_requirements`
2. Update phase markdown headers (`echo-work-order-sync`)
3. Re-export CSV from Factory if URLs/IDs drifted
4. `make knowledge-graph`

Agents: skill **`echo-knowledge-graph`**.
