# SUT Core - 漏洞赏金挖掘框架核心架构

基于README.md中定义的架构实现的原型系统核心组件。

## 架构概览

```
SUT Core
├── interfaces.py          # 核心接口定义
│   ├── IPlugin           # 插件接口基类
│   ├── IPluginManager    # 插件管理器接口
│   ├── IScheduler        # 调度器接口
│   ├── IMonitor          # 监控器接口
│   └── IOrchestrator     # 协调器接口
├── plugin_manager.py      # 插件管理器实现
├── scheduler.py           # 任务调度器实现
├── monitor.py             # 监控器实现
├── orchestrator.py        # 协调器实现
├── base_plugin.py         # 基础插件类
└── tests/                 # 架构测试
```

## 核心组件说明

### 1. Plugin Manager (插件管理器)

**职责**: 负责插件的加载、注册、配置和生命周期管理

**主要功能**:
- 动态加载插件（从文件或模块）
- 插件注册和注销
- 批量加载插件目录
- 插件配置管理
- 插件生命周期管理（初始化、清理）
- 插件启用/禁用控制

**关键方法**:
```python
async def load_plugin(self, plugin_path: str) -> Optional[IPlugin]
def register_plugin(self, plugin: IPlugin) -> bool
def get_plugin(self, name: str) -> Optional[IPlugin]
async def initialize_all(self, configs: Dict[str, Dict])
async def cleanup_all(self)
```

### 2. Scheduler (调度器)

**职责**: 负责任务的调度、执行和状态管理

**主要功能**:
- 优先级任务队列（基于堆实现）
- 并发任务执行（可配置工作线程数）
- 任务状态跟踪
- 任务超时处理
- 任务取消
- 事件回调机制
- 统计信息收集

**关键方法**:
```python
async def submit_task(self, task: Task, plugin: IPlugin) -> str
async def cancel_task(self, task_id: str) -> bool
def get_task(self, task_id: str) -> Optional[Task]
async def wait_for_completion(self, task_ids: List[str], timeout: float) -> bool
```

### 3. Monitor (监控器)

**职责**: 负责系统状态监控、指标收集和事件记录

**主要功能**:
- 任务生命周期事件记录
- 漏洞收集和分类
- 实时监控指标
- 报告生成（JSON格式）
- 时间线记录
- 事件回调通知

**关键方法**:
```python
def record_task_created(self, task: Task)
def record_vulnerability(self, vuln: Vulnerability)
def get_metrics(self) -> Dict[str, Any]
def generate_report(self, output_file: Optional[str]) -> str
```

### 4. Orchestrator (协调器)

**职责**: 统一管理系统组件，提供高层API

**主要功能**:
- 组件初始化和启动
- 扫描任务管理
- 组件间协调
- 状态查询
- 报告生成
- 单例模式支持

**关键方法**:
```python
async def initialize(self)
async def start(self)
async def stop(self)
async def scan(self, config: ScanConfig) -> List[str]
async def scan_and_wait(self, config: ScanConfig, timeout: float) -> Dict
```

## 数据模型

### Task (任务)
```python
@dataclass
class Task:
    id: str                    # 任务ID
    name: str                  # 任务名称
    plugin_name: str          # 执行插件
    target: str               # 检测目标
    status: TaskStatus        # 任务状态
    priority: int             # 优先级(1-10)
    created_at: datetime      # 创建时间
    started_at: datetime      # 开始时间
    completed_at: datetime    # 完成时间
    timeout: int              # 超时时间
    result: Dict              # 执行结果
    error: str                # 错误信息
    metadata: Dict            # 元数据
```

### Vulnerability (漏洞)
```python
@dataclass
class Vulnerability:
    id: str                   # 漏洞ID
    task_id: str             # 关联任务ID
    severity: Severity       # 严重程度
    title: str               # 标题
    description: str         # 描述
    target: str              # 目标
    plugin_name: str         # 发现插件
    evidence: Dict           # 证据
    remediation: str         # 修复建议
    references: List[str]    # 参考链接
    timestamp: datetime      # 发现时间
    verified: bool           # 是否已验证
```

### PluginInfo (插件信息)
```python
@dataclass
class PluginInfo:
    name: str                # 插件名称
    version: str             # 版本
    description: str         # 描述
    author: str              # 作者
    plugin_type: str         # 类型
    dependencies: List[str]  # 依赖
    config_schema: Dict      # 配置模式
    enabled: bool            # 是否启用
    status: PluginStatus     # 状态
```

## 使用示例

### 基础用法

```python
import asyncio
from sofa_core import (
    Orchestrator, get_orchestrator, reset_orchestrator,
    BasePlugin, Task, Severity, ScanConfig
)

# 定义自定义插件
class MyPlugin(BasePlugin):
    def __init__(self):
        super().__init__("my_plugin", "1.0.0")
        self.set_plugin_info(
            description="My custom plugin",
            plugin_type="custom"
        )
    
    async def _do_execute(self, task):
        # 实现检测逻辑
        vuln = self.create_vulnerability(
            task=task,
            severity=Severity.HIGH,
            title="Critical Issue",
            description="Found critical vulnerability"
        )
        return {
            'findings': [vuln],
            'findings_count': 1
        }

async def main():
    # 重置单例（首次使用可省略）
    reset_orchestrator()
    
    # 获取协调器
    config = {
        "scheduler": {"max_workers": 5},
        "plugins": {}
    }
    orch = get_orchestrator(config)
    
    # 注册插件
    orch.register_plugin(MyPlugin())
    
    # 初始化并启动
    await orch.initialize()
    await orch.start()
    
    # 执行扫描
    scan_config = ScanConfig(
        target="https://example.com",
        plugins=["my_plugin"],
        max_workers=5
    )
    
    result = await orch.scan_and_wait(scan_config, timeout=60)
    print(f"Completed: {result['tasks_completed']}")
    print(f"Vulnerabilities: {result['vulnerabilities']}")
    
    # 生成报告
    report = orch.generate_report("report.json")
    
    # 停止系统
    await orch.stop()

if __name__ == "__main__":
    asyncio.run(main())
```

### 直接使用组件

```python
from sofa_core import (
    PluginManager, Scheduler, Monitor,
    BasePlugin, Task
)

# 独立使用插件管理器
pm = PluginManager()
plugin = MyPlugin()
pm.register_plugin(plugin)
await pm.initialize_all({"my_plugin": {}})

# 独立使用调度器
scheduler = Scheduler(max_workers=3)
await scheduler.start()
task_id = await scheduler.submit_task(task, plugin)
await scheduler.wait_for_completion([task_id])
await scheduler.stop()

# 独立使用监控器
monitor = Monitor()
monitor.record_task_created(task)
monitor.record_vulnerability(vuln)
metrics = monitor.get_metrics()
```

## 扩展开发

### 创建自定义插件

继承`BasePlugin`类并实现`_do_execute`方法：

```python
from sofa_core import BasePlugin, Task, Severity

class CustomPlugin(BasePlugin):
    def __init__(self):
        super().__init__("custom", "1.0.0")
        self.set_plugin_info(
            description="Custom detection plugin",
            plugin_type="custom"
        )
    
    async def _do_execute(self, task: Task) -> Dict:
        # 实现检测逻辑
        findings = []
        
        # 检测代码...
        if self._detect_issue(task.target):
            vuln = self.create_vulnerability(
                task=task,
                severity=Severity.HIGH,
                title="Issue Found",
                description="Detailed description",
                evidence={"key": "value"}
            )
            findings.append(vuln)
        
        return {
            'findings': findings,
            'findings_count': len(findings)
        }
    
    def _detect_issue(self, target: str) -> bool:
        # 具体检测逻辑
        pass
```

## 架构特点

1. **接口驱动**: 所有核心组件都基于接口定义，便于替换实现
2. **模块化**: 组件间松耦合，可独立使用
3. **异步支持**: 全面支持async/await，高效处理并发
4. **可扩展**: 插件架构支持动态扩展检测能力
5. **类型安全**: 完整的类型注解支持
6. **事件驱动**: 支持事件回调和监控
7. **配置灵活**: 支持多种配置方式

## 运行测试

```bash
python sofa_core/tests/test_architecture.py
```

## 与原始架构对比

本实现严格遵循README.md中的架构设计：

- ✅ Orchestrator - 统一协调器
- ✅ Plugin Manager - 插件管理器
- ✅ Scheduler - 任务调度器
- ✅ Monitor - 系统监控
- ✅ IPlugin接口 - 插件基类
- ✅ 支持多种插件类型（通过继承实现）

架构设计强调**先实现核心管理组件，再实现具体插件**，确保系统的灵活性和可扩展性。