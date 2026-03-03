# rule_based_checker.py
"""Rule-based checker plugin for prototype validation.

Implements Phase 1 requirement: loads YAML rules and matches on CPG.
"""

import sys
import os
# Add the plugin-manager directory to the path to import the interfaces
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'plugin-manager'))

from base_plugin import BasePlugin, Vulnerability
from interfaces import Severity
from typing import Dict, Any, List
import yaml


class RuleBasedChecker(BasePlugin):
    """Rule-based checker that loads YAML rules and matches them on CPG."""

    def __init__(self):
        super().__init__("rule-based-checker", "0.1.0")
        # Explicit I/O contract per requirements
        self.info.input_types = ["cpg"]
        self.info.output_types = ["vulnerability"]
        self.rules = []

    async def initialize(self, config: Dict[str, Any]) -> bool:
        """Initialize the checker with rule files from config."""
        success = await super().initialize(config)
        if not success:
            return False
            
        # Load rules from YAML files specified in config
        rule_files = config.get('rule_files', [])
        for rule_file in rule_files:
            try:
                with open(rule_file, 'r') as f:
                    rule_data = yaml.safe_load(f)
                    self.rules.extend(rule_data.get('rules', []))
            except Exception as e:
                print(f"Failed to load rule file {rule_file}: {e}")
                return False
        
        return True

    async def _do_execute(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute rule-based static analysis on CPG.
        
        Returns:
            Dict with list of matched vulnerabilities.
        """
        # Extract CPG from task input
        cpg = task.get('input', {})
        
        findings = []
        # Apply rules to CPG nodes
        for rule in self.rules:
            rule_findings = self._apply_rule_to_cpg(cpg, rule)
            findings.extend(rule_findings)
        
        return {
            'findings': findings,
            'count': len(findings),
            'status': 'completed'
        }
    
    def _apply_rule_to_cpg(self, cpg: Dict[str, Any], rule: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Apply a single rule to the CPG and return matched vulnerabilities."""
        findings = []
        
        # For Phase 1, implement a simple node matching
        # In future phases, implement more complex pattern matching
        nodes = cpg.get('nodes', [])
        
        for node in nodes:
            # Check if node matches rule criteria
            if self._node_matches_rule(node, rule):
                finding = {
                    "id": f"RULE-{rule.get('id', 'UNKNOWN')}",
                    "severity": rule.get('severity', 'medium'),
                    "title": rule.get('title', 'Rule Match'),
                    "description": rule.get('description', 'Pattern matched by rule'),
                    "location": {
                        "file": cpg.get('metadata', {}).get('source_file', 'unknown'),
                        "line": node.get('line', 0)
                    },
                    "evidence": {
                        "node_type": node.get('type'),
                        "node_name": node.get('name', ''),
                        "pattern_matched": rule.get('pattern', '')
                    },
                    "status": "reported"
                }
                findings.append(finding)
        
        return findings
    
    def _node_matches_rule(self, node: Dict[str, Any], rule: Dict[str, Any]) -> bool:
        """Check if a node matches the rule criteria."""
        # Simple pattern matching for Phase 1
        pattern = rule.get('pattern', {})
        
        # Check if node type matches
        if 'type' in pattern:
            if node.get('type') != pattern['type']:
                return False
        
        # Check if node name matches
        if 'name' in pattern:
            if pattern['name'] in str(node.get('name', '')):
                return True
        
        # Additional simple matching can be added here
        return False