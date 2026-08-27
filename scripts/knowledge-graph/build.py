#!/usr/bin/env python3
"""Build Echo traceability graph: factory.8090.ai WOs → blueprints/requirements → code.

Sources (priority for status):
  1. docs/phase-*-work-orders.md section headers (latest SF export in git)
  2. docs/echo-work-orders-*.csv (Factory URLs, blueprint titles, WO ids)
  3. FEATURES overlay in this file (code reality as of last review)

Outputs:
  docs/knowledge-graph/graph.json
  docs/knowledge-graph/TRACEABILITY.md
"""

from __future__ import annotations

import csv
import json
import re
import sys
from collections import Counter, defaultdict
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DOCS = ROOT / "docs"
OUT = DOCS / "knowledge-graph"
FACTORY_HOST = "https://factory.8090.ai"
PROJECT_ID = "0f477a55-4e26-4500-98f3-a69e4f3381b3"

SECTION_STATUS = {
    "completed": "completed",
    "backlog": "backlog",
    "in progress": "in_progress",
    "blocked": "blocked",
    "ready": "ready",
}

# Code-reality overlay. Status is what a 2-device tester will hit today,
# not the May CSV. Blueprint names must match Factory titles when possible.
FEATURES = [
    {
        "id": "identity-didkey",
        "name": "did:key identity + passkey REST",
        "status": "shipped",
        "blueprints": ["Decentralized Identity and Authentication", "Secure Enclave Key Management"],
        "code": ["pkg/didkey/", "internal/api/passkey_auth.go", "ios/Echo/Sources/Services/PasskeySigningInterceptor"],
        "e2e": "Both devices complete first-run; REST calls carry X-Sender-DID + X-Signature",
        "sept1": True,
    },
    {
        "id": "onboarding-frozen",
        "name": "Frozen onboarding + Glacial login",
        "status": "shipped",
        "blueprints": ["Streamlined Onboarding with Verifiable Credentials and Passkeys", "Universal Onboarding and Identity Creation"],
        "code": [
            "ios/Echo/Sources/Features/Onboarding/FirstRun/FirstRunCoordinator.swift",
            "ios/Echo/Sources/Features/Auth/Views/GlacialLoginScreen.swift",
        ],
        "e2e": "Do not redesign. Validate both devices authenticate. Login is frozen.",
        "sept1": True,
    },
    {
        "id": "ws-relay",
        "name": "Content-blind WebSocket relay",
        "status": "shipped",
        "blueprints": ["End-to-End Message Encryption and Commitment", "Backend"],
        "code": ["internal/api/ws.go", "ios/Echo/Sources/Services/MessageRelaySession.swift", "ios/Echo/Sources/Networking/WebSocketClient.swift"],
        "e2e": "D2: live DM while both apps foregrounded",
        "sept1": True,
    },
    {
        "id": "offline-delivery",
        "name": "Durable offline queue + APNs wake",
        "status": "partial",
        "blueprints": ["End-to-End Message Encryption and Commitment", "Backend"],
        "code": [
            "internal/api/ws.go",
            "internal/infra/apns.go",
            "ios/Echo/Sources/Services/PushRegistrationService.swift",
            "ios/Echo/Sources/Services/MessageRelaySession.swift",
        ],
        "e2e": "D3: B force-quit → A sends → queue + optional APNs → B opens → flush",
        "sept1": True,
        "notes": "Needs DATABASE_URL for durable queue; APNS_* for lock-screen wake. Without APNs, B still gets messages on open.",
    },
    {
        "id": "phase3-signals",
        "name": "Typing, read receipts, reactions",
        "status": "shipped",
        "blueprints": ["Message Reactions, Polls, and Interactive Elements"],
        "code": [
            "internal/api/ws.go",
            "ios/Echo/Sources/Services/ConversationSignalService.swift",
            "ios/Echo/Sources/Features/Messaging/TypingIndicatorView.swift",
        ],
        "e2e": "D4: typing / receipts / reactions on both devices",
        "sept1": True,
    },
    {
        "id": "username-invite",
        "name": "@username public invite",
        "status": "shipped",
        "blueprints": ["Privacy-Preserving Contact Discovery"],
        "code": [
            "ios/Echo/Sources/Features/Contacts/EchoDeepLink.swift",
            "ios/Echo/Sources/Domain/UseCases/Contacts/ContactUseCases.swift",
            "internal/api/",
        ],
        "e2e": "D1: A shares echo://invite?u=handle → B accepts → both in contacts",
        "sept1": True,
    },
    {
        "id": "qr-contacts",
        "name": "QR identity exchange",
        "status": "shipped",
        "blueprints": ["Privacy-Preserving Contact Discovery"],
        "code": ["ios/Echo/Sources/Features/Contacts/QRContactAddCoordinator.swift"],
        "e2e": "D1b: Profile QR → other device scans → POST /v3/contacts/add",
        "sept1": True,
    },
    {
        "id": "backup-restore",
        "name": "Phrase-encrypted backup + restore-did",
        "status": "partial",
        "blueprints": ["Data Sovereignty Layer", "Universal Onboarding and Identity Creation"],
        "code": [
            "internal/api/enrollment_handlers.go",
            "internal/api/backup_handlers.go",
            "ios/Echo/Sources/Services/RecoveryService.swift",
            "ios/Echo/Sources/Services/MessageBackupService.swift",
        ],
        "e2e": "D5: A backups → B restores phrase → chats return. Wallet must be linked.",
        "sept1": True,
        "notes": "Cloud backup E2E still human/Xcode. WO-CA3 full multi-device history is Wave S1.",
    },
    {
        "id": "groups",
        "name": "Encrypted groups",
        "status": "shipped",
        "blueprints": ["Public and Private Groups with Verified Status Display"],
        "code": ["internal/services/groups/", "ios/Echo/Sources/Features/Groups/"],
        "e2e": "D6: create group, add B, bidirectional messages, offline fan-out",
        "sept1": True,
    },
    {
        "id": "channels",
        "name": "Broadcast channels",
        "status": "partial",
        "blueprints": ["Broadcast Channels and Community Features"],
        "code": ["internal/api/", "ios/Echo/Sources/Features/Messaging/ChannelsListView.swift"],
        "e2e": "Optional: create/subscribe/post. Do not market as durable without Postgres soak.",
        "sept1": False,
    },
    {
        "id": "telegram-parity",
        "name": "Telegram-class messaging UX",
        "status": "partial",
        "blueprints": ["Telegram-class Messaging UX", "Broadcast Channels and Community Features"],
        "code": [
            "docs/TELEGRAM_ECHO_PARITY.md",
            "ios/Echo/Sources/Features/Messaging/SavedMessagesStore.swift",
            "ios/Echo/Sources/Features/Messaging/ComposerDraftStore.swift",
            "ios/Echo/Sources/Features/Messaging/ForwardMessageSheet.swift",
            "internal/api/scheduled_handlers.go",
            "internal/api/v3_handlers.go",
        ],
        "e2e": "T1 two-device: Saved Messages, drafts, forward, search. T0 channels soak. T2 folder + schedule HTTP.",
        "sept1": False,
        "notes": "Factory WO-335 epic (T0 WO-336, T1 WO-337, T2 WO-338, T3 WO-339). Skip Stories/Stars/Nearby/Secret Chat clone.",
    },
    {
        "id": "calls",
        "name": "1:1 voice/video calls",
        "status": "partial",
        "blueprints": ["Voice and Video Calls with Screen Sharing"],
        "code": ["internal/api/ws.go", "ios/Echo/Sources/Features/Calling/"],
        "e2e": "Optional if WebRTC linked. Missed-call push needs APNs.",
        "sept1": False,
    },
    {
        "id": "rewards",
        "name": "Rewards hub (earn, not buy)",
        "status": "partial",
        "blueprints": ["ECHO Token Reward System and Incentive Economy", "User Rewards Tracker on Profile"],
        "code": ["internal/api/tokenomics_handlers.go", "ios/Echo/Sources/Presentation/Screens/Rewards/"],
        "e2e": "Smoke: Rewards tab loads; no withdraw/VIP purchase CTA",
        "sept1": True,
    },
    {
        "id": "hidden-folders",
        "name": "Hidden folders (night palette)",
        "status": "partial",
        "blueprints": ["Hidden Folders with Biometric Protection"],
        "code": ["ios/Echo/Sources/Features/HiddenFolders/"],
        "e2e": "Per-device vault. Night tokens only here.",
        "sept1": False,
    },
    {
        "id": "identity-l1",
        "name": "Identity L1 username + trust anchors",
        "status": "shipped",
        "blueprints": ["Decentralized Identity and Authentication", "Dynamic Trust Network and Social Verification"],
        "code": ["metagraph/modules/identity_l1/", "internal/metagraph/"],
        "e2e": "Needed for validate-phase1 step 3 and username anchors. make start-identity.",
        "sept1": True,
    },
]


def _split_titles(raw: str) -> list[str]:
    if not raw:
        return []
    parts = []
    for chunk in raw.replace(";", ",").split(","):
        t = chunk.strip()
        if t:
            parts.append(t)
    return parts


def _norm_status(raw: str) -> str:
    s = (raw or "").strip().lower().replace(" ", "_")
    aliases = {
        "in_progress": "in_progress",
        "inprogress": "in_progress",
        "done": "completed",
        "complete": "completed",
        "completed": "completed",
        "backlog": "backlog",
        "blocked": "blocked",
        "ready": "ready",
        "in_review": "in_review",
    }
    return aliases.get(s, s or "unknown")


def latest_csv() -> Path:
    files = sorted(DOCS.glob("echo-work-orders-*.csv"))
    if not files:
        raise SystemExit("No docs/echo-work-orders-*.csv Factory export found")
    return files[-1]


def parse_csv(path: Path) -> dict[int, dict]:
    wos: dict[int, dict] = {}
    with path.open(newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            num_s = (row.get("Work Order Number") or "").replace("WO-", "").strip()
            if not num_s.isdigit():
                continue
            n = int(num_s)
            reqs = _split_titles(row.get("Requirement Titles") or "")
            bps = _split_titles(row.get("Blueprint Titles") or "")
            # Factory export often leaves Requirement Titles empty; blueprints
            # share names with the requirements document sections.
            if not reqs:
                reqs = list(bps)
            wos[n] = {
                "number": n,
                "id": row.get("Work Order ID") or "",
                "title": (row.get("Title") or "").strip(),
                "status_csv": _norm_status(row.get("Status") or ""),
                "priority": (row.get("Priority") or "").strip(),
                "type": (row.get("Type") or "").strip(),
                "phase": (row.get("Phase") or "").strip(),
                "requirements": reqs,
                "blueprints": bps,
                "url": (row.get("URL") or "").strip()
                or f"{FACTORY_HOST}/project/{PROJECT_ID}/work-orders?workOrderId={row.get('Work Order ID') or ''}",
                "parent": (row.get("Parent Work Order") or "").strip(),
                "blocked_by": (row.get("Blocked By") or "").strip(),
                "blocking": (row.get("Blocking") or "").strip(),
            }
    return wos


def parse_phase_markdown() -> tuple[dict[int, dict], dict[str, dict]]:
    """Return (wo overlays, phase headers)."""
    overlays: dict[int, dict] = {}
    phases: dict[str, dict] = {}
    wo_re = re.compile(r"^###\s+WO-(\d+)\s*:\s*(.+)$")
    section_re = re.compile(r"^##\s+(.+?)(?:\s+\(\d+\))?\s*$")
    header_total = re.compile(r"\*\*Total Work Orders:\*\*\s*(\d+)")
    header_sync = re.compile(r"\*\*Last synced with Software Factory:\*\*\s*(.+)$")
    header_summary = re.compile(r"\*\*Status Summary:\*\*\s*(.+)$")

    for path in sorted(DOCS.glob("phase-*-work-orders.md")):
        m = re.search(r"phase-(\d+)", path.name)
        if not m:
            continue
        phase_num = m.group(1)
        current_section = None
        meta: dict = {"file": str(path.relative_to(ROOT))}
        for line in path.read_text(encoding="utf-8").splitlines():
            if ht := header_total.search(line):
                meta["total"] = int(ht.group(1))
            if hs := header_sync.search(line):
                meta["last_synced"] = hs.group(1).strip()
            if sm := header_summary.search(line):
                meta["status_summary"] = sm.group(1).strip()
            sec = section_re.match(line)
            if sec:
                label = sec.group(1).strip().lower()
                if label.startswith("⚠"):
                    label = label.lstrip("⚠ ").strip()
                matched = False
                for key, status in SECTION_STATUS.items():
                    if label.startswith(key):
                        current_section = status
                        matched = True
                        break
                if matched:
                    continue
                # Inner headings (## Summary, ## In Scope) keep the bucket.
            wo = wo_re.match(line)
            if wo and current_section:
                n = int(wo.group(1))
                overlays[n] = {
                    "status_md": current_section,
                    "title_md": wo.group(2).strip(),
                    "phase_md": phase_num,
                    "source": path.name,
                }
        phases[phase_num] = meta
    return overlays, phases


def load_factory_live() -> dict:
    path = OUT / "factory-live.json"
    if not path.is_file():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}


def live_status_overrides(live: dict) -> dict[int, str]:
    out: dict[int, str] = {}
    for n in live.get("blocked_ids") or []:
        out[int(n)] = "blocked"
    for row in live.get("in_progress") or []:
        out[int(row["id"])] = "in_progress"
    for row in live.get("ready") or []:
        out[int(row["id"])] = "ready"
    return out


def merge(csv_wos: dict[int, dict], md_wos: dict[int, dict], live: dict | None = None) -> list[dict]:
    numbers = sorted(set(csv_wos) | set(md_wos))
    out = []
    for n in numbers:
        base = csv_wos.get(n, {
            "number": n,
            "id": "",
            "title": md_wos.get(n, {}).get("title_md", f"WO-{n}"),
            "status_csv": "",
            "priority": "",
            "type": "",
            "phase": md_wos.get(n, {}).get("phase_md", ""),
            "requirements": [],
            "blueprints": [],
            "url": f"{FACTORY_HOST}/project/{PROJECT_ID}/work-orders",
            "parent": "",
            "blocked_by": "",
            "blocking": "",
        })
        md = md_wos.get(n, {})
        title = md.get("title_md") or base.get("title")
        phase = md.get("phase_md") or base.get("phase") or ""
        if md.get("status_md"):
            status = md["status_md"]
            status_source = "phase-md"
        elif n not in md_wos and base.get("id"):
            # Remaining-work markdown omits shipped WOs; CSV still lists them.
            status = "completed"
            status_source = "inferred-absent-from-phase-md"
        else:
            status = base.get("status_csv") or "unknown"
            status_source = "csv"
        node = {
            **base,
            "title": title,
            "status": status,
            "phase": str(phase),
            "status_source": status_source,
        }
        out.append(node)
    overrides = live_status_overrides(live or {})
    by_num = {w["number"]: w for w in out}
    for n, st in overrides.items():
        if n in by_num:
            by_num[n]["status"] = st
            by_num[n]["status_source"] = "factory-live"
            continue
        title = f"WO-{n}"
        phase = ""
        for row in (live or {}).get("in_progress") or []:
            if int(row["id"]) == n:
                title = row.get("title") or title
                phase = str(row.get("phase") or "")
        for row in (live or {}).get("ready") or []:
            if int(row["id"]) == n:
                title = row.get("title") or title
                phase = str(row.get("phase") or "")
        out.append({
            "number": n,
            "id": "",
            "title": title,
            "status_csv": "",
            "priority": "",
            "type": "",
            "phase": phase,
            "requirements": [],
            "blueprints": [],
            "url": f"{FACTORY_HOST}/project/{PROJECT_ID}/work-orders",
            "parent": "",
            "blocked_by": "",
            "blocking": "",
            "status": st,
            "status_source": "factory-live",
        })
    return sorted(out, key=lambda w: w["number"])


def blueprint_nodes(wos: list[dict]) -> list[dict]:
    counts: dict[str, Counter] = defaultdict(Counter)
    wo_ids: dict[str, list[int]] = defaultdict(list)
    for wo in wos:
        names = wo["blueprints"] or ["(unlinked)"]
        for bp in names:
            counts[bp][wo["status"]] += 1
            wo_ids[bp].append(wo["number"])
    nodes = []
    for name, st in sorted(counts.items(), key=lambda kv: (-sum(kv[1].values()), kv[0])):
        nodes.append({
            "id": f"bp:{name}",
            "type": "blueprint",
            "name": name,
            "requirement_name": name,  # 1:1 with requirements TOC in this project
            "wo_count": sum(st.values()),
            "status_counts": dict(st),
            "work_orders": sorted(wo_ids[name]),
            "factory_search": f"{FACTORY_HOST}/project/{PROJECT_ID}",
        })
    return nodes


def attach_features(wos: list[dict]) -> list[dict]:
    features = []
    by_bp = defaultdict(list)
    for wo in wos:
        for bp in wo["blueprints"]:
            by_bp[bp].append(wo["number"])
    for feat in FEATURES:
        related = []
        for bp in feat["blueprints"]:
            related.extend(by_bp.get(bp, []))
        features.append({
            **feat,
            "related_work_orders": sorted(set(related)),
        })
    return features


def write_traceability(graph: dict) -> None:
    lines = [
        "# Echo traceability graph",
        "",
        f"Generated `{graph['generated_at']}` from Factory CSV `{graph['sources']['csv']}` "
        f"+ `docs/phase-*-work-orders.md` (last SF sync in headers: "
        f"{graph['sources']['phase_md_sync']}).",
        "",
        "Canonical Factory: [factory.8090.ai](https://factory.8090.ai) · "
        f"project `{PROJECT_ID}`.",
        "",
        f"Live MCP pull: `{graph['sources'].get('factory_live') or 'none'}` "
        f"({graph.get('factory_live', {}).get('total_work_orders', '—')} WOs).",
        "",
        "Rebuild: `make knowledge-graph`. Agents: skill `echo-knowledge-graph`.",
        "",
        "## Chain",
        "",
        "```text",
        "Requirements section  ==  Blueprint title  (same name in Echo)",
        "        ↓",
        "Work order  (factory.8090.ai URL on every WO node)",
        "        ↓",
        "Code path  (FEATURES overlay + gap audits)",
        "        ↓",
        "2-device E2E scenario  (docs/E2E_TWO_DEVICE.md)",
        "```",
        "",
        "## Phase status (markdown export)",
        "",
        "| Phase | Total (header) | Last synced | Status summary |",
        "|-------|----------------|-------------|----------------|",
    ]
    for num, meta in sorted(graph["phases"].items(), key=lambda kv: int(kv[0])):
        lines.append(
            f"| {num} | {meta.get('total', '—')} | {meta.get('last_synced', '—')} | "
            f"{meta.get('status_summary', '—')} |"
        )
    live_st = (graph.get("factory_live") or {}).get("status") or {}
    st = Counter(w["status"] for w in graph["work_orders"])
    lines += [
        "",
        "## Work order status (live Factory)",
        "",
        "| Status | Count |",
        "|--------|-------|",
    ]
    if live_st:
        for k in ("completed", "in_progress", "ready", "backlog", "blocked"):
            if k in live_st:
                lines.append(f"| {k} | {live_st[k]} |")
        lines.append(
            f"| **total** | **{(graph.get('factory_live') or {}).get('total_work_orders', '')}** |"
        )
    else:
        for k in ("completed", "in_progress", "ready", "backlog", "blocked", "in_review", "unknown"):
            if st[k]:
                lines.append(f"| {k} | {st[k]} |")
    lines += [
        "",
        f"**Nodes in graph.json:** {len(graph['work_orders'])} (CSV + phase md + live IP/ready/blocked).",
        "",
        "Per-WO status in graph.json overlays factory-live for in_progress / ready / blocked. "
        "Completed vs backlog counts above are the live Factory totals.",
        "",
        "### In progress now",
        "",
        "| WO | Title | Phase |",
        "|----|-------|-------|",
    ]
    for row in (graph.get("factory_live_lists") or {}).get("in_progress") or []:
        lines.append(f"| WO-{row['id']} | {row['title']} | {row.get('phase', '')} |")
    lines += [
        "",
        "### Ready (go-live / parity queue)",
        "",
        "| WO | Title | Phase |",
        "|----|-------|-------|",
    ]
    for row in (graph.get("factory_live_lists") or {}).get("ready") or []:
        lines.append(f"| WO-{row['id']} | {row['title']} | {row.get('phase', '')} |")
    lines += [
        "",
        "## Blueprints → requirements → WOs",
        "",
        "Requirement Titles in the CSV export are empty; each blueprint title is the "
        "requirements-document section of the same name.",
        "",
        "| Blueprint / requirement | WOs | completed | in_progress | backlog | blocked |",
        "|-------------------------|-----|-----------|-------------|---------|---------|",
    ]
    for bp in graph["blueprints"]:
        sc = bp["status_counts"]
        lines.append(
            f"| {bp['name']} | {bp['wo_count']} | {sc.get('completed', 0)} | "
            f"{sc.get('in_progress', 0)} | {sc.get('backlog', 0)} | {sc.get('blocked', 0)} |"
        )
    lines += [
        "",
        "## Product features (code reality)",
        "",
        "| Feature | Status | Sept 1 | E2E | Factory blueprints |",
        "|---------|--------|--------|-----|--------------------|",
    ]
    for feat in graph["features"]:
        bps = ", ".join(feat["blueprints"])
        lines.append(
            f"| {feat['name']} | {feat['status']} | {'yes' if feat.get('sept1') else 'no'} | "
            f"{feat['e2e']} | {bps} |"
        )
    lines += [
        "",
        "## Sept 1 critical WOs (open in Factory)",
        "",
        "Use the URL on each node in `graph.json`. Highest-signal numbers:",
        "",
        "| WO | Why |",
        "|----|-----|",
        "| [WO-230](https://factory.8090.ai/project/"
        + PROJECT_ID
        + ") | Phase 1 go/no-go (`make validate-phase1`) |",
        "| WO-4 / messaging core | Encryption + delivery |",
        "| WO-221 / WO-222 | PSI (optional) + invite deep links |",
        "| WO-288 | Link new device (Week B B6) |",
        "| WO-233 | App Store go-live |",
        "",
        "Full node list: [`graph.json`](graph.json).",
        "",
    ]
    (OUT / "TRACEABILITY.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> int:
    csv_path = latest_csv()
    csv_wos = parse_csv(csv_path)
    md_wos, phases = parse_phase_markdown()
    live = load_factory_live()
    if live.get("phases"):
        for num, meta in live["phases"].items():
            phases.setdefault(str(num), {})
            phases[str(num)]["total"] = meta.get("total", phases[str(num)].get("total"))
            phases[str(num)]["status_summary"] = meta.get(
                "summary", phases[str(num)].get("status_summary")
            )
            phases[str(num)]["last_synced"] = "2026-08-27 (factory-live.json)"
    wos = merge(csv_wos, md_wos, live)
    bps = blueprint_nodes(wos)
    features = attach_features(wos)
    sync_dates = sorted({p.get("last_synced", "") for p in phases.values() if p.get("last_synced")})
    graph = {
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "factory": {"host": FACTORY_HOST, "project_id": PROJECT_ID},
        "sources": {
            "csv": str(csv_path.relative_to(ROOT)),
            "phase_markdown": [p["file"] for p in phases.values()],
            "phase_md_sync": ", ".join(sync_dates) or "unknown",
            "features_overlay": "scripts/knowledge-graph/build.py FEATURES",
            "factory_live": "docs/knowledge-graph/factory-live.json" if live else None,
        },
        "factory_live": {
            "total_work_orders": live.get("total_work_orders"),
            "status": live.get("status"),
            "pulled_at": live.get("pulled_at"),
        } if live else {},
        "factory_live_lists": {
            "in_progress": live.get("in_progress") or [],
            "ready": live.get("ready") or [],
            "blocked_ids": live.get("blocked_ids") or [],
        },
        "counts": {
            "work_orders": len(wos),
            "blueprints": len(bps),
            "features": len(features),
            "status": dict(Counter(w["status"] for w in wos)),
        },
        "phases": phases,
        "blueprints": bps,
        "features": features,
        "work_orders": wos,
        "edges": {
            "requirement_eq_blueprint": True,
            "blueprint_to_wo": [
                {"from": bp["id"], "to": f"WO-{n}"}
                for bp in bps
                for n in bp["work_orders"]
            ],
            "feature_to_blueprint": [
                {"from": f"feat:{feat['id']}", "to": f"bp:{bp}"}
                for feat in features
                for bp in feat["blueprints"]
            ],
        },
    }
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / "graph.json").write_text(json.dumps(graph, indent=2) + "\n", encoding="utf-8")
    write_traceability(graph)
    print(f"Wrote {OUT / 'graph.json'} ({len(wos)} WOs, {len(bps)} blueprints)")
    print(f"Wrote {OUT / 'TRACEABILITY.md'}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
