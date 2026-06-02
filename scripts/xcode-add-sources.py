#!/usr/bin/env python3
"""Add Swift source files to a classic (non-synchronized) Xcode project.pbxproj.

Inserts the 4 entries each file needs: PBXBuildFile, PBXFileReference, the owning
PBXGroup child, and the PBXSourcesBuildPhase entry. IDs are fresh 24-hex strings.

Usage:
    python3 scripts/xcode-add-sources.py <project.pbxproj> GROUPID:Basename.swift [GROUPID:Other.swift ...]

The project's PBXGroup `path` is already the folder, so the file reference `path`
is just the basename (matches how sibling sources are referenced).
Idempotent: a basename already present in the file is skipped.
Always writes a .bak backup first.
"""
import re
import secrets
import sys


def new_id() -> str:
    return secrets.token_hex(12).upper()  # 24 hex chars


def insert_after(text: str, anchor: str, payload: str) -> str:
    idx = text.index(anchor) + len(anchor)
    return text[:idx] + payload + text[idx:]


def add_group_child(text: str, group_id: str, line: str) -> str:
    pat = re.compile(
        re.escape(group_id) + r" /\* [^*]+ \*/ = \{\s*isa = PBXGroup;\s*children = \(\n"
    )
    m = pat.search(text)
    if not m:
        raise SystemExit(f"Group {group_id} not found")
    return text[: m.end()] + line + text[m.end():]


def add_source_phase(text: str, line: str) -> str:
    pat = re.compile(r"isa = PBXSourcesBuildPhase;.*?files = \(\n", re.DOTALL)
    m = pat.search(text)
    if not m:
        raise SystemExit("PBXSourcesBuildPhase files list not found")
    return text[: m.end()] + line + text[m.end():]


def main() -> None:
    if len(sys.argv) < 3:
        raise SystemExit(__doc__)
    pbx = sys.argv[1]
    specs = sys.argv[2:]

    with open(pbx) as f:
        text = f.read()
    with open(pbx + ".bak", "w") as f:
        f.write(text)

    added = []
    for spec in specs:
        group_id, path_attr = spec.split(":", 1)
        # Display name is the basename; the file-ref `path` may be a full relative
        # path (mirrors the project's existing full-path references) or a basename.
        name = path_attr.rsplit("/", 1)[-1]
        if f"path = {path_attr};" in text:
            print(f"skip (already present): {path_attr}")
            continue
        build_id, ref_id = new_id(), new_id()
        build_line = (
            f"\t\t{build_id} /* {path_attr} in Sources */ = "
            f"{{isa = PBXBuildFile; fileRef = {ref_id} /* {path_attr} */; }};\n"
        )
        ref_line = (
            f"\t\t{ref_id} /* {path_attr} */ = {{isa = PBXFileReference; "
            f"lastKnownFileType = sourcecode.swift; path = {path_attr}; sourceTree = \"<group>\"; }};\n"
        )
        child_line = f"\t\t\t\t{ref_id} /* {path_attr} */,\n"
        source_line = f"\t\t\t\t{build_id} /* {path_attr} in Sources */,\n"

        text = insert_after(text, "/* Begin PBXBuildFile section */\n", build_line)
        text = insert_after(text, "/* Begin PBXFileReference section */\n", ref_line)
        text = add_group_child(text, group_id, child_line)
        text = add_source_phase(text, source_line)
        added.append(name)
        print(f"added: {name}  (build={build_id} ref={ref_id} group={group_id})")

    with open(pbx, "w") as f:
        f.write(text)
    print(f"\n{len(added)} file(s) added. Backup at {pbx}.bak")


if __name__ == "__main__":
    main()
