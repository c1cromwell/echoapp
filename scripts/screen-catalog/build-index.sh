#!/usr/bin/env bash
# Build index.html from docs/screen_catalog/manifest.jsonl
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CATALOG="${SCREEN_CATALOG_ROOT:-$ROOT/docs/screen_catalog}"
MANIFEST="$CATALOG/manifest.jsonl"
OUT="$CATALOG/index.html"
GIT_SHA="$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_DATE="$(date -u +"%Y-%m-%d %H:%M UTC")"

if [[ ! -f "$MANIFEST" ]]; then
  echo "error: missing $MANIFEST — run scripts/screen-catalog/generate.sh first"
  exit 1
fi

python3 - "$MANIFEST" "$OUT" "$GIT_SHA" "$BUILD_DATE" <<'PY'
import json
import sys
from collections import defaultdict
from html import escape
from pathlib import Path

manifest_path, out_path, git_sha, build_date = sys.argv[1:5]
journeys = defaultdict(list)
titles = {
    "onboarding": "Onboarding",
    "auth": "Login & lock",
    "messaging": "Messaging (Week A)",
    "privacy": "Privacy & advanced",
    "contacts": "Contacts (Week B)",
}

with open(manifest_path, encoding="utf-8") as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        entry = json.loads(line)
        journeys[entry["journey"]].append(entry)

order = ["onboarding", "auth", "messaging", "contacts", "privacy"]
tab_ids = [j for j in order if j in journeys] + [j for j in journeys if j not in order]

def panel(jid, active):
    rows = journeys[jid]
    cards = []
    for row in rows:
        cards.append(
            f"""
            <article class="card">
              <header>
                <h3>{escape(row['title'])}</h3>
                <p class="e2e">{escape(row['e2eRef'])}</p>
              </header>
              <a href="{escape(row['relativePath'])}" target="_blank">
                <img src="{escape(row['relativePath'])}" alt="{escape(row['title'])}" loading="lazy" />
              </a>
              <p class="meta">{escape(row['relativePath'])}</p>
            </article>
            """
        )
    display = "block" if active else "none"
    label = titles.get(jid, jid.replace("-", " ").title())
    return f'<section class="panel" id="panel-{jid}" style="display:{display}"><h2>{escape(label)}</h2><div class="grid">{"".join(cards)}</div></section>'

tabs = []
panels = []
for i, jid in enumerate(tab_ids):
    label = titles.get(jid, jid.replace("-", " ").title())
    active = "active" if i == 0 else ""
    tabs.append(
        f'<button type="button" class="tab {active}" data-tab="{jid}" onclick="showTab(\'{jid}\')">{escape(label)}</button>'
    )
    panels.append(panel(jid, i == 0))

html = f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>ECHO Screen Catalog</title>
  <style>
    :root {{
      --bg: #0f1419;
      --surface: #1a222c;
      --text: #e8eef4;
      --muted: #8b9aab;
      --accent: #3ecf8e;
      --border: #2a3544;
    }}
    * {{ box-sizing: border-box; }}
    body {{
      margin: 0;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: var(--bg);
      color: var(--text);
      line-height: 1.45;
    }}
    header.page {{
      padding: 1.5rem 2rem;
      border-bottom: 1px solid var(--border);
      background: var(--surface);
    }}
    header.page h1 {{ margin: 0 0 0.25rem; font-size: 1.5rem; }}
    header.page p {{ margin: 0; color: var(--muted); font-size: 0.9rem; }}
    nav.tabs {{
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
      padding: 1rem 2rem;
      border-bottom: 1px solid var(--border);
      background: var(--surface);
      position: sticky;
      top: 0;
      z-index: 10;
    }}
    .tab {{
      border: 1px solid var(--border);
      background: transparent;
      color: var(--text);
      padding: 0.5rem 1rem;
      border-radius: 999px;
      cursor: pointer;
      font-size: 0.9rem;
    }}
    .tab.active {{
      background: var(--accent);
      color: #0a0f12;
      border-color: var(--accent);
      font-weight: 600;
    }}
    main {{ padding: 1.5rem 2rem 3rem; max-width: 1400px; margin: 0 auto; }}
    .panel h2 {{ font-size: 1.1rem; color: var(--muted); font-weight: 600; margin-bottom: 1rem; }}
    .grid {{
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 1.5rem;
    }}
    .card {{
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 12px;
      overflow: hidden;
    }}
    .card header {{ padding: 1rem 1rem 0.5rem; }}
    .card h3 {{ margin: 0; font-size: 1rem; }}
    .e2e {{ margin: 0.25rem 0 0; font-size: 0.8rem; color: var(--muted); }}
    .card img {{
      width: 100%;
      display: block;
      background: #000;
      border-top: 1px solid var(--border);
      border-bottom: 1px solid var(--border);
    }}
    .meta {{ padding: 0.5rem 1rem 1rem; font-size: 0.75rem; color: var(--muted); margin: 0; font-family: ui-monospace, monospace; }}
    a {{ color: inherit; text-decoration: none; }}
    a:hover img {{ opacity: 0.95; }}
  </style>
</head>
<body>
  <header class="page">
    <h1>ECHO Screen Catalog</h1>
    <p>SwiftUI <code>ImageRenderer</code> exports · commit {escape(git_sha)} · {escape(build_date)}</p>
    <p>Regenerate: <code>make screen-catalog</code> from repo root. Maps to <code>E2E_LAUNCH_AND_TESTING.md</code> / TestFlight scripts.</p>
  </header>
  <nav class="tabs" aria-label="Journeys">
    {"".join(tabs)}
  </nav>
  <main>
    {"".join(panels)}
  </main>
  <script>
    function showTab(id) {{
      document.querySelectorAll('.panel').forEach(p => p.style.display = 'none');
      document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
      const panel = document.getElementById('panel-' + id);
      if (panel) panel.style.display = 'block';
      document.querySelectorAll('.tab').forEach(t => {{
        if (t.dataset.tab === id) t.classList.add('active');
      }});
    }}
  </script>
</body>
</html>
"""

Path(out_path).write_text(html, encoding="utf-8")
print(f"wrote {out_path}")
PY
