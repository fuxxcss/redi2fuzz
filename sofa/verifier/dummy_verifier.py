# dummy_verifier.py
"""Dummy verifier plugin for prototype validation.

Implements minimal IPlugin contract for verification tasks:
- Input: vulnerability object (from Checker)
- Output: verified vulnerability with confirmation flag
"""

from .plugin_manager.base_plugin import BasePlugin
from typing import Dict, Any


class DummyVerifierPlugin(BasePlugin):
    """Dummy verifier that always confirms vulnerabilities (stub)."""

    def __init__(self):
        super().__init__("dummy-verifier", "0.1.0")
        # Explicit I/O contract per requirements
        self.info.input_types = ["vulnerability"]
        self.info.output_types = ["verified_vulnerability"]

    async def _do_execute(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute verification (stub implementation).

        Returns:
            Dict with verified vulnerability status.
        """
        # In real implementation: run PoC, fuzz, or debug
        return {
            "verified": True,
            "confidence": 0.8,
            "proof_of_concept": None,
            "reason": "stub-verification",
            "status": "confirmed"
        }