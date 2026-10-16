#!/usr/bin/env python3
"""Keep Linux Wails test builds on the WebKit ABI installed by CI."""
from pathlib import Path
import re
import shlex
import unittest

ROOT = Path(__file__).resolve().parents[1]


def missing_webkit_tag(workflow: str) -> list[str]:
    match = re.search(r"(?ms)^  full:\n(.*?)(?=^  [\w-]+:|\Z)", workflow)
    if match is None:
        raise ValueError("missing full Linux test job")
    violations = []
    for command in re.findall(r"(?m)^\s+run: (go test .+)$", match.group(1)):
        arguments = shlex.split(command)
        if not any(arg in ("./...", ".", "./internal/app", "./internal/app/") for arg in arguments):
            continue
        tags = []
        for index, argument in enumerate(arguments):
            if argument == "-tags" and index + 1 < len(arguments):
                tags.extend(arguments[index + 1].replace(",", " ").split())
            elif argument.startswith("-tags="):
                tags.extend(argument[6:].replace(",", " ").split())
        if "webkit2_41" not in tags:
            violations.append(command)
    return violations


class BackendWorkflowTest(unittest.TestCase):
    def test_app_and_root_builds_select_installed_webkit(self):
        workflow = (ROOT / ".github/workflows/backend-tests.yml").read_text()
        self.assertEqual(missing_webkit_tag(workflow), [], "Linux app/root tests must select webkit2_41")

    def test_detects_driver_command_even_with_narrow_test_filter(self):
        workflow = "  full:\n    run: go test -tags gonavi_duckdb_driver ./internal/app/ -run DuckDB\n"
        self.assertEqual(len(missing_webkit_tag(workflow)), 1)
        self.assertEqual(missing_webkit_tag(workflow.replace("gonavi_duckdb_driver", "webkit2_41,gonavi_duckdb_driver")), [])

    def test_other_platform_jobs_are_outside_linux_policy(self):
        workflow = "  full:\n    run: go test -tags=webkit2_41 ./...\n  platform-sensitive:\n    run: go test ./internal/app/\n"
        self.assertEqual(missing_webkit_tag(workflow), [])


if __name__ == "__main__":
    unittest.main()
