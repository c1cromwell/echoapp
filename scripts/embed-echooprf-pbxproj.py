#!/usr/bin/env python3
"""WO-221: Add EchoOPRF.xcframework to EchoApp.xcodeproj when built locally."""
from __future__ import annotations

import pathlib
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[1]
FRAMEWORK = ROOT / "ios/Echo/Libraries/EchoOPRF.xcframework"
PBXPROJ = ROOT / "ios/Echo/EchoApp.xcodeproj/project.pbxproj"


def uid() -> str:
    return uuid.uuid4().hex[:24].upper()


def main() -> int:
    if not FRAMEWORK.is_dir():
        print(f"Missing {FRAMEWORK} — run: make echooprf-ios", file=sys.stderr)
        return 1
    text = PBXPROJ.read_text()
    if "EchoOPRF.xcframework" in text:
        print("EchoOPRF.xcframework already in project.pbxproj")
        return 0

    file_ref = uid()
    build_file = uid()
    rel_path = "Libraries/EchoOPRF.xcframework"

    file_ref_line = (
        f"\t\t{file_ref} /* EchoOPRF.xcframework */ = "
        f"{{isa = PBXFileReference; lastKnownFileType = wrapper.xcframework; "
        f'name = EchoOPRF.xcframework; path = {rel_path}; sourceTree = "<group>"; }};\n'
    )
    build_file_line = (
        f"\t\t{build_file} /* EchoOPRF.xcframework in Frameworks */ = "
        f"{{isa = PBXBuildFile; fileRef = {file_ref} /* EchoOPRF.xcframework */; }};\n"
    )

    marker = "/* End PBXBuildFile section */"
    if marker not in text:
        print("Could not find PBXBuildFile section", file=sys.stderr)
        return 1
    text = text.replace(marker, build_file_line + marker, 1)

    marker = "/* End PBXFileReference section */"
    if marker not in text:
        print("Could not find PBXFileReference section", file=sys.stderr)
        return 1
    text = text.replace(marker, file_ref_line + marker, 1)

    # Embed in EchoApp Frameworks build phase (first PBXFrameworksBuildPhase files list).
    embed = f"\t\t\t\t{build_file} /* EchoOPRF.xcframework in Frameworks */,\n"
    needle = "/* Frameworks */ = {\n\t\t\tisa = PBXFrameworksBuildPhase;\n\t\t\trunOnlyForDeploymentPostprocessing = 0;\n\t\t\tfiles = (\n"
    if needle not in text:
        print("Could not find Frameworks build phase", file=sys.stderr)
        return 1
    text = text.replace(needle, needle + embed, 1)

    # Add to Libraries group if present.
    libraries_needle = "path = Libraries;\n\t\t\tsourceTree = \"<group>\";\n\t\t};\n"
    if libraries_needle in text:
        group_child = f"\t\t\t\t{file_ref} /* EchoOPRF.xcframework */,\n"
        text = text.replace(libraries_needle, group_child + libraries_needle, 1)

    PBXPROJ.write_text(text)
    print(f"Embedded EchoOPRF.xcframework in {PBXPROJ}")
    print("Open Xcode → EchoApp target → General → set Embed & Sign on EchoOPRF.xcframework if needed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
