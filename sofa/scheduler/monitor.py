# Copyright (c) 2025. MIT License.
"""
Monitor 实现模块
负责系统状态监控、指标收集和事件记录
"""

import json
import logging
from collections import defaultdict
from datetime import datetime
from typing import Any, Dict, List, Optional, Callable
from pathlib import Path

from .interfaces import (
    IMonitor, Task, Vulnerability, Severity, TaskStatus,
    PluginStatus
)

logger = logging.getLogger(__name__)


class Monitor(IMonitor):
    """监控器实现类
    
    提供任务监控、漏洞收集、指标统计和报告生成功能
    """
    
    def __init__(self, log_dir: Optional[str] = None):
        """初始化监控器
        
        Args:
            log_dir: 日志目录路径
        """
        # 漏洞存储
        self._vulnerabilities: List[Vulnerability] = []
        self._vuln_by_severity: Dict[Severity, List[Vulnerability]] = defaultdict(list)
        self._vuln_by_plugin: Dict[str, List[Vulnerability]] = defaultdict(list)
        
        # 任务统计
        self._task_stats = {
            'created': 0,
            'started': 0,
            'completed': 0,
            'failed': 0,
            'cancelled': 0,
            'timeout': 0
        }
        
        # 插件事件
        self._plugin_events: List[Dict[str, Any]] = []
        
        # 时间线记录
        self._timeline: List[Dict[str, Any]] = []
        
        # 回调函数
        self._callbacks: List[Callable] = []
        
        # 日志目录
        self._log_dir = Path(log_dir) if log_dir else None
        if self._log_dir:
            self._log_dir.mkdir(parents=True, exist_ok=True)
        
        logger.info("Monitor initialized")
    
    def record_task_created(self, task: Task):
        """记录任务创建"""
        self._task_stats['created'] += 1
        
        event = {
            'type': 'task_created',
            'timestamp': datetime.now().isoformat(),
            'task_id': task.id,
            'task_name': task.name,
            'plugin': task.plugin_name,
            'target': task.target,
            'priority': task.priority
        }
        
        self._timeline.append(event)
        self._notify(event)
        
        logger.debug(f"Task created recorded: {task.id}")
    
    def record_task_started(self, task: Task):
        """记录任务开始"""
        self._task_stats['started'] += 1
        
        event = {
            'type': 'task_started',
            'timestamp': datetime.now().isoformat(),
            'task_id': task.id,
            'task_name': task.name,
            'plugin': task.plugin_name
        }
        
        self._timeline.append(event)
        self._notify(event)
        
        logger.info(f"Task started recorded: {task.id}")
    
    def record_task_completed(self, task: Task):
        """记录任务完成"""
        self._task_stats['completed'] += 1
        
        duration = None
        if task.started_at and task.completed_at:
            duration = (task.completed_at - task.started_at).total_seconds()
        
        event = {
            'type': 'task_completed',
            'timestamp': datetime.now().isoformat(),
            'task_id': task.id,
            'task_name': task.name,
            'plugin': task.plugin_name,
            'duration': duration,
            'result_summary': self._summarize_result(task.result)
        }
        
        self._timeline.append(event)
        self._notify(event)
        
        logger.info(f"Task completed recorded: {task.id}")
    
    def record_task_failed(self, task: Task, error: str):
        """记录任务失败"""
        if task.status == TaskStatus.TIMEOUT:
            self._task_stats['timeout'] += 1
        else:
            self._task_stats['failed'] += 1
        
        event = {
            'type': 'task_failed',
            'timestamp': datetime.now().isoformat(),
            'task_id': task.id,
            'task_name': task.name,
            'plugin': task.plugin_name,
            'status': task.status.name,
            'error': error
        }
        
        self._timeline.append(event)
        self._notify(event)
        
        logger.warning(f"Task failed recorded: {task.id}, status={task.status.name}")
    
    def record_vulnerability(self, vuln: Vulnerability):
        """记录漏洞发现"""
        self._vulnerabilities.append(vuln)
        self._vuln_by_severity[vuln.severity].append(vuln)
        self._vuln_by_plugin[vuln.plugin_name].append(vuln)
        
        event = {
            'type': 'vulnerability_found',
            'timestamp': datetime.now().isoformat(),
            'vuln_id': vuln.id,
            'severity': vuln.severity.value,
            'title': vuln.title,
            'plugin': vuln.plugin_name,
            'target': vuln.target
        }
        
        self._timeline.append(event)
        self._notify(event)
        
        logger.warning(
            f"Vulnerability found: [{vuln.severity.value.upper()}] {vuln.title} "
            f"on {vuln.target} by {vuln.plugin_name}"
        )
    
    def record_plugin_event(self, plugin_name: str, event: str, data: Any):
        """记录插件事件"""
        plugin_event = {
            'timestamp': datetime.now().isoformat(),
            'plugin': plugin_name,
            'event': event,
            'data': data
        }
        
        self._plugin_events.append(plugin_event)
        
        logger.debug(f"Plugin event recorded: {plugin_name} - {event}")
    
    def get_metrics(self) -> Dict[str, Any]:
        """获取监控指标"""
        return {
            'tasks': self._task_stats.copy(),
            'vulnerabilities': {
                'total': len(self._vulnerabilities),
                'by_severity': {
                    sev.value: len(vulns)
                    for sev, vulns in self._vuln_by_severity.items()
                },
                'by_plugin': {
                    plugin: len(vulns)
                    for plugin, vulns in self._vuln_by_plugin.items()
                }
            },
            'timeline_events': len(self._timeline),
            'plugin_events': len(self._plugin_events)
        }
    
    def get_vulnerabilities(
        self,
        severity: Optional[Severity] = None,
        plugin_name: Optional[str] = None,
        verified_only: bool = False
    ) -> List[Vulnerability]:
        """获取漏洞列表
        
        Args:
            severity: 按严重程度筛选
            plugin_name: 按插件名称筛选
            verified_only: 是否只返回已验证的漏洞
            
        Returns:
            List[Vulnerability]: 漏洞列表
        """
        vulns = self._vulnerabilities
        
        if severity:
            vulns = [v for v in vulns if v.severity == severity]
        
        if plugin_name:
            vulns = [v for v in vulns if v.plugin_name == plugin_name]
        
        if verified_only:
            vulns = [v for v in vulns if v.verified]
        
        return vulns
    
    def get_vulnerability_summary(self) -> Dict[str, Any]:
        """获取漏洞摘要"""
        return {
            'total': len(self._vulnerabilities),
            'by_severity': {
                sev.value: {
                    'count': len(vulns),
                    'vulnerabilities': [
                        {
                            'id': v.id,
                            'title': v.title,
                            'target': v.target,
                            'plugin': v.plugin_name,
                            'timestamp': v.timestamp.isoformat()
                        }
                        for v in vulns
                    ]
                }
                for sev, vulns in sorted(
                    self._vuln_by_severity.items(),
                    key=lambda x: ['critical', 'high', 'medium', 'low', 'info'].index(x[0].value)
                )
            }
        }
    
    def generate_report(self, output_file: Optional[str] = None) -> str:
        """生成扫描报告
        
        Args:
            output_file: 输出文件路径，None则返回JSON字符串
            
        Returns:
            str: 报告JSON字符串
        """
        report = {
            'generated_at': datetime.now().isoformat(),
            'summary': {
                'tasks': self._task_stats,
                'vulnerabilities': {
                    'total': len(self._vulnerabilities),
                    'by_severity': {
                        sev.value: len(vulns)
                        for sev, vulns in self._vuln_by_severity.items()
                    }
                }
            },
            'vulnerabilities': [
                {
                    'id': v.id,
                    'severity': v.severity.value,
                    'title': v.title,
                    'description': v.description,
                    'target': v.target,
                    'plugin_name': v.plugin_name,
                    'evidence': v.evidence,
                    'remediation': v.remediation,
                    'references': v.references,
                    'timestamp': v.timestamp.isoformat(),
                    'verified': v.verified
                }
                for v in sorted(
                    self._vulnerabilities,
                    key=lambda x: (
                        ['critical', 'high', 'medium', 'low', 'info'].index(x.severity.value),
                        x.timestamp
                    ),
                    reverse=True
                )
            ],
            'timeline': self._timeline,
            'plugin_events': self._plugin_events[-100:]  # 最近100个事件
        }
        
        json_str = json.dumps(report, indent=2, ensure_ascii=False)
        
        if output_file:
            output_path = Path(output_file)
            output_path.write_text(json_str, encoding='utf-8')
            logger.info(f"Report saved to: {output_file}")
        
        return json_str
    
    def export_timeline(self, output_file: str):
        """导出时间线记录"""
        output_path = Path(output_file)
        output_path.write_text(
            json.dumps(self._timeline, indent=2, ensure_ascii=False),
            encoding='utf-8'
        )
        logger.info(f"Timeline exported to: {output_file}")
    
    def register_callback(self, callback: Callable):
        """注册监控回调函数
        
        Args:
            callback: 回调函数，接收事件字典参数
        """
        self._callbacks.append(callback)
    
    def unregister_callback(self, callback: Callable):
        """注销监控回调函数"""
        if callback in self._callbacks:
            self._callbacks.remove(callback)
    
    def _notify(self, event: Dict[str, Any]):
        """通知所有回调"""
        for callback in self._callbacks:
            try:
                if callable(callback):
                    callback(event)
            except Exception as e:
                logger.error(f"Monitor callback error: {e}")
    
    def _summarize_result(self, result: Optional[Dict[str, Any]]) -> Dict[str, Any]:
        """摘要任务结果"""
        if not result:
            return {}
        
        # 只保留关键信息，避免报告过大
        summary = {}
        
        if 'status' in result:
            summary['status'] = result['status']
        
        if 'findings_count' in result:
            summary['findings_count'] = result['findings_count']
        
        if 'error' in result:
            summary['error'] = result['error']
        
        return summary
    
    def clear(self):
        """清空所有记录"""
        self._vulnerabilities.clear()
        self._vuln_by_severity.clear()
        self._vuln_by_plugin.clear()
        
        for key in self._task_stats:
            self._task_stats[key] = 0
        
        self._plugin_events.clear()
        self._timeline.clear()
        
        logger.info("Monitor records cleared")
    
    def get_statistics(self) -> Dict[str, Any]:
        """获取详细统计信息"""
        # 计算平均任务执行时间
        durations = []
        for event in self._timeline:
            if event['type'] == 'task_completed' and event.get('duration'):
                durations.append(event['duration'])
        
        avg_duration = sum(durations) / len(durations) if durations else 0
        
        return {
            'task_statistics': {
                **self._task_stats,
                'success_rate': (
                    self._task_stats['completed'] / max(self._task_stats['started'], 1)
                ),
                'average_duration': avg_duration,
                'total_duration': sum(durations) if durations else 0
            },
            'vulnerability_statistics': {
                'total': len(self._vulnerabilities),
                'by_severity': {
                    sev.value: len(vulns)
                    for sev, vulns in self._vuln_by_severity.items()
                },
                'by_plugin': {
                    plugin: len(vulns)
                    for plugin, vulns in self._vuln_by_plugin.items()
                },
                'verified_count': sum(1 for v in self._vulnerabilities if v.verified)
            },
            'time_range': {
                'start': self._timeline[0]['timestamp'] if self._timeline else None,
                'end': self._timeline[-1]['timestamp'] if self._timeline else None,
                'event_count': len(self._timeline)
            }
        }