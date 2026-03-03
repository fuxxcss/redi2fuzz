# dummy_checker.py
"""Dummy checker plugin for prototype validation.

Implements minimal IPlugin contract for static analysis:
- Input: CPG graph (from Parser)
- Output: list of vulnerability findings (stub)
"""

import sys
import os
# Add the plugin-manager directory to the path to import the interfaces
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'plugin-manager'))

from base_plugin import BasePlugin, Vulnerability
from interfaces import Severity
from typing import Dict, Any, List


class DummyCheckerPlugin(BasePlugin):
    """Dummy checker that returns fixed vulnerabilities (stub)."""

    def __init__(self):
        super().__init__("dummy-checker", "0.1.0")
        # Explicit I/O contract per requirements
        self.info.input_types = ["cpg"]
        self.info.output_types = ["vulnerability"]

    async def _do_execute(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute static analysis (stub implementation).

        Returns:
            Dict with list of dummy vulnerabilities.
        """
        # In real implementation: run dataflow/taint analysis on CPG
        findings: List[Dict] = [
            {
                "id": "CVE-2025-001",
                "severity": Severity.LOW.value,
                "title": "Stub vulnerability",
                "description": "This is a placeholder vulnerability from dummy checker.",
                "location": {"file": "unknown.c", "line": 42},
                "evidence": {"snippet": "int *p = NULL; *p = 1;"},
                "status": "reported"
            }
        ]

        return {
            "findings": findings,
            "count": len(findings),
            "status": "completed"
        }