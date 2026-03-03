# Copyright (c) 2025. MIT License.
# This is the sole scheduler in the system. Orchestrator is deprecated.
"""
Scheduler implementation module.
Manages task scheduling, execution, and state tracking.
"""

import asyncio
import logging
from collections import deque
from datetime import datetime
from typing import Any, Dict, List, Optional, Callable
import heapq

from .interfaces import (
    IScheduler, IPlugin, Task, TaskStatus,
    TaskNotFoundException, SchedulerException,
    IMonitor
)

logger = logging.getLogger(__name__)


class PriorityTask:
    """带优先级的任务包装类
    
    用于优先级队列，支持比较操作
    """
    
    def __init__(self, task: Task, plugin: IPlugin):
        self.task = task
        self.plugin = plugin
        self.priority = task.priority
        self.created_at = task.created_at
        
    def __lt__(self, other):
        # 优先级数字越小优先级越高
        if self.priority != other.priority:
            return self.priority < other.priority
        # 优先级相同，先创建的优先
        return self.created_at < other.created_at
    
    def __eq__(self, other):
        return self.task.id == other.task.id


class Scheduler(IScheduler):
    """任务调度器实现类
    
    提供并发任务调度、优先级管理、状态跟踪功能
    """
    
    def __init__(self, max_workers: int = 5, queue_size: int = 1000, monitor: Optional[IMonitor] = None):
        """初始化调度器
        
        Args:
            max_workers: 最大并发工作线程数
            queue_size: 任务队列最大容量
            monitor: 监控器实例
        """
        self._max_workers = max_workers
        self._queue_size = queue_size
        self._monitor = monitor
        
        # 任务存储
        self._tasks: Dict[str, Task] = {}
        self._task_plugins: Dict[str, IPlugin] = {}
        
        # 优先级队列 (使用堆实现)
        self._queue: List[PriorityTask] = []
        self._queue_lock = asyncio.Lock()
        
        # 运行状态
        self._running = False
        self._workers: List[asyncio.Task] = []
        self._semaphore = asyncio.Semaphore(max_workers)
        
        # 运行中的任务跟踪
        self._executing_tasks: Dict[str, asyncio.Task] = {}
        
        # 事件通知
        self._task_completed_event = asyncio.Event()
        self._callbacks: Dict[str, List[Callable]] = {
            'task_created': [],
            'task_started': [],
            'task_completed': [],
            'task_failed': [],
            'task_cancelled': []
        }
        
        # 统计信息
        self._stats = {
            'submitted': 0,
            'completed': 0,
            'failed': 0,
            'cancelled': 0,
            'timeout': 0
        }
        
        logger.info(f"Scheduler initialized (max_workers={max_workers}, queue_size={queue_size})")
    
    async def start(self):
        """启动调度器"""
        if self._running:
            logger.warning("Scheduler is already running")
            return
        
        self._running = True
        logger.info(f"Scheduler started with {self._max_workers} workers")
        
        # 启动工作协程
        self._workers = [
            asyncio.create_task(self._worker_loop(i))
            for i in range(self._max_workers)
        ]
    
    async def stop(self):
        """停止调度器"""
        if not self._running:
            return
        
        self._running = False
        logger.info("Stopping scheduler...")
        
        # 取消所有工作协程
        for worker in self._workers:
            worker.cancel()
        
        # 等待所有工作协程完成
        if self._workers:
            await asyncio.gather(*self._workers, return_exceptions=True)
        
        # 取消队列中所有等待的任务
        async with self._queue_lock:
            for priority_task in self._queue:
                task = priority_task.task
                if task.status == TaskStatus.PENDING:
                    task.status = TaskStatus.CANCELLED
                    self._stats['cancelled'] += 1
                    self._trigger_callback('task_cancelled', task)
        
        self._queue.clear()
        logger.info("Scheduler stopped")
    
    async def submit_task(self, task: Task, plugin: IPlugin) -> str:
        """提交任务到调度队列
        
        Args:
            task: 任务对象
            plugin: 执行任务的插件
            
        Returns:
            str: 任务ID
        """
        if not self._running:
            raise SchedulerException("Scheduler is not running")
        
        # 检查队列容量
        async with self._queue_lock:
            if len(self._queue) >= self._queue_size:
                raise SchedulerException(f"Task queue is full (max={self._queue_size})")
        
        # 存储任务
        self._tasks[task.id] = task
        self._task_plugins[task.id] = plugin
        
        # 创建优先级任务并加入队列
        priority_task = PriorityTask(task, plugin)
        
        async with self._queue_lock:
            heapq.heappush(self._queue, priority_task)
        
        self._stats['submitted'] += 1
        logger.info(f"Task submitted: {task.id} (priority={task.priority}, plugin={plugin.get_info().name})")
        
        self._trigger_callback('task_created', task)
        
        # 如果有监控器，记录任务创建
        if self._monitor:
            self._monitor.record_task_created(task)
        
        return task.id
    
    async def submit_batch(self, plugin: IPlugin, inputs: List[Any]) -> List[str]:
        """批量提交任务
        
        Args:
            plugin: 执行任务的插件
            inputs: 输入列表
            
        Returns:
            List[str]: 任务ID列表
        """
        if not self._running:
            raise SchedulerException("Scheduler is not running")
        
        task_ids = []
        for input_data in inputs:
            # 创建任务
            task = Task()
            task.plugin_name = plugin.get_info().name
            task.target = str(input_data)[:100]  # 限制目标描述长度
            task.metadata['input_data'] = input_data
            
            # 提交任务
            task_id = await self.submit_task(task, plugin)
            task_ids.append(task_id)
        
        logger.info(f"Batch submitted {len(task_ids)} tasks for plugin {plugin.get_info().name}")
        return task_ids

    async def cancel_task(self, task_id: str) -> bool:
        """取消指定任务
        
        Args:
            task_id: 任务ID
            
        Returns:
            bool: 是否成功取消
        """
        task = self._tasks.get(task_id)
        if not task:
            return False
        
        # 如果任务还在等待队列中，从队列移除
        if task.status == TaskStatus.PENDING:
            async with self._queue_lock:
                # 从堆中移除指定任务
                self._queue = [pt for pt in self._queue if pt.task.id != task_id]
                heapq.heapify(self._queue)
            
            task.status = TaskStatus.CANCELLED
            self._stats['cancelled'] += 1
            self._trigger_callback('task_cancelled', task)
            logger.info(f"Task cancelled: {task_id}")
            
            # 如果有监控器，记录任务取消
            if self._monitor:
                self._monitor.record_task_failed(task, "Task was cancelled")
            
            return True
        
        # 如果任务正在运行，尝试取消执行中的任务
        if task.status == TaskStatus.RUNNING:
            if task_id in self._executing_tasks:
                executing_task = self._executing_tasks[task_id]
                executing_task.cancel()
                
                try:
                    # 等待任务真正取消
                    await executing_task
                except asyncio.CancelledError:
                    pass  # 期望的结果
                
                # 更新任务状态
                task.status = TaskStatus.CANCELLED
                task.completed_at = datetime.now()
                self._stats['cancelled'] += 1
                self._trigger_callback('task_cancelled', task)
                
                logger.info(f"Running task cancelled: {task_id}")
                
                # 如果有监控器，记录任务取消
                if self._monitor:
                    self._monitor.record_task_failed(task, "Running task was cancelled")
                
                return True
            
        return False
    
    def get_task(self, task_id: str) -> Optional[Task]:
        """获取任务状态
        
        Args:
            task_id: 任务ID
            
        Returns:
            Task: 任务对象，不存在返回None
        """
        return self._tasks.get(task_id)
    
    def list_tasks(self, status: Optional[TaskStatus] = None) -> List[Task]:
        """列出任务
        
        Args:
            status: 按状态筛选，None表示所有状态
            
        Returns:
            List[Task]: 任务列表
        """
        tasks = list(self._tasks.values())
        if status:
            tasks = [t for t in tasks if t.status == status]
        return tasks
    
    def get_stats(self) -> Dict[str, Any]:
        """获取调度器统计信息"""
        return {
            'running': self._running,
            'max_workers': self._max_workers,
            'queue_size': len(self._queue),
            'total_tasks': len(self._tasks),
            'stats': self._stats.copy(),
            'active_tasks': len([t for t in self._tasks.values() if t.status == TaskStatus.RUNNING])
        }
    
    async def _worker_loop(self, worker_id: int):
        """工作协程主循环
        
        Args:
            worker_id: 工作线程ID
        """
        logger.debug(f"Worker {worker_id} started")
        
        while self._running:
            try:
                # 从队列获取任务
                priority_task = await self._get_task_from_queue()
                
                if priority_task:
                    # 使用信号量控制并发
                    async with self._semaphore:
                        # 创建执行任务的协程
                        executing_task = asyncio.create_task(self._execute_task(priority_task))
                        # 记录执行中的任务，以便可以取消
                        self._executing_tasks[priority_task.task.id] = executing_task
                        
                        try:
                            await executing_task
                        finally:
                            # 移除执行完成的任务记录
                            if priority_task.task.id in self._executing_tasks:
                                del self._executing_tasks[priority_task.task.id]
                else:
                    # 队列为空，短暂等待
                    await asyncio.sleep(0.1)
                    
            except asyncio.CancelledError:
                logger.debug(f"Worker {worker_id} cancelled")
                break
            except Exception as e:
                logger.error(f"Worker {worker_id} error: {e}")
        
        logger.debug(f"Worker {worker_id} stopped")
    
    async def _get_task_from_queue(self) -> Optional[PriorityTask]:
        """从队列获取一个任务"""
        async with self._queue_lock:
            if self._queue:
                return heapq.heappop(self._queue)
        return None
    
    async def _execute_task(self, priority_task: PriorityTask):
        """执行单个任务
        
        Args:
            priority_task: 优先级任务包装对象
        """
        task = priority_task.task
        plugin = priority_task.plugin
        
        # 更新任务状态
        task.status = TaskStatus.RUNNING
        task.started_at = datetime.now()
        
        logger.info(f"Task started: {task.id} (plugin={plugin.get_info().name})")
        self._trigger_callback('task_started', task)
        
        # 如果有监控器，记录任务开始
        if self._monitor:
            self._monitor.record_task_started(task)
        
        try:
            # 设置超时
            result = await asyncio.wait_for(
                plugin.execute(task),
                timeout=task.timeout
            )
            
            # 更新任务结果
            task.result = result
            task.status = TaskStatus.COMPLETED
            task.completed_at = datetime.now()
            
            self._stats['completed'] += 1
            self._trigger_callback('task_completed', task)
            
            logger.info(f"Task completed: {task.id}")
            
            # 如果有监控器，记录任务完成
            if self._monitor:
                self._monitor.record_task_completed(task)
            
        except asyncio.TimeoutError:
            task.status = TaskStatus.TIMEOUT
            task.error = f"Task timed out after {task.timeout} seconds"
            task.completed_at = datetime.now()
            
            self._stats['timeout'] += 1
            self._stats['failed'] += 1
            self._trigger_callback('task_failed', task)
            
            logger.warning(f"Task timeout: {task.id}")
            
            # 如果有监控器，记录任务失败
            if self._monitor:
                self._monitor.record_task_failed(task, task.error)
            
        except asyncio.CancelledError:
            # 任务被取消
            task.status = TaskStatus.CANCELLED
            task.completed_at = datetime.now()
            
            self._stats['cancelled'] += 1
            self._trigger_callback('task_cancelled', task)
            
            logger.info(f"Task cancelled: {task.id}")
            
            # 如果有监控器，记录任务取消
            if self._monitor:
                self._monitor.record_task_failed(task, "Task was cancelled")
            
        except Exception as e:
            task.status = TaskStatus.FAILED
            task.error = str(e)
            task.completed_at = datetime.now()
            
            self._stats['failed'] += 1
            self._trigger_callback('task_failed', task)
            
            logger.error(f"Task failed: {task.id}, error={e}")
            
            # 如果有监控器，记录任务失败
            if self._monitor:
                self._monitor.record_task_failed(task, task.error)
        
        finally:
            self._task_completed_event.set()
    
    def register_callback(self, event: str, callback: Callable):
        """注册事件回调
        
        Args:
            event: 事件类型 ('task_created', 'task_started', 'task_completed', 'task_failed', 'task_cancelled')
            callback: 回调函数
        """
        if event in self._callbacks:
            self._callbacks[event].append(callback)
    
    def unregister_callback(self, event: str, callback: Callable):
        """注销事件回调"""
        if event in self._callbacks and callback in self._callbacks[event]:
            self._callbacks[event].remove(callback)
    
    def _trigger_callback(self, event: str, task: Task):
        """触发事件回调"""
        for callback in self._callbacks.get(event, []):
            try:
                if asyncio.iscoroutinefunction(callback):
                    asyncio.create_task(callback(task))
                else:
                    callback(task)
            except Exception as e:
                logger.error(f"Callback error for event {event}: {e}")
    
    async def wait_for_completion(self, task_ids: Optional[List[str]] = None, 
                                   timeout: Optional[float] = None) -> bool:
        """等待任务完成
        
        Args:
            task_ids: 要等待的任务ID列表，None表示所有任务
            timeout: 超时时间(秒)
            
        Returns:
            bool: 是否在超时前完成
        """
        start_time = datetime.now()
        
        while True:
            # 检查指定任务是否都已完成
            if task_ids:
                all_completed = all(
                    self._tasks.get(tid) and 
                    self._tasks[tid].status in [TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED, TaskStatus.TIMEOUT]
                    for tid in task_ids
                )
            else:
                # 检查所有任务是否都已完成
                all_completed = all(
                    task.status in [TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED, TaskStatus.TIMEOUT]
                    for task in self._tasks.values()
                )
            
            if all_completed:
                return True
            
            # 检查超时
            if timeout:
                elapsed = (datetime.now() - start_time).total_seconds()
                if elapsed >= timeout:
                    return False
            
            # 等待任务完成事件
            try:
                await asyncio.wait_for(
                    self._task_completed_event.wait(),
                    timeout=0.5
                )
                self._task_completed_event.clear()
            except asyncio.TimeoutError:
                pass
    
    def clear_completed_tasks(self):
        """清理已完成的任务"""
        completed_statuses = [TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED, TaskStatus.TIMEOUT]
        
        to_remove = [
            tid for tid, task in self._tasks.items()
            if task.status in completed_statuses
        ]
        
        for tid in to_remove:
            del self._tasks[tid]
            if tid in self._task_plugins:
                del self._task_plugins[tid]
        
        logger.info(f"Cleared {len(to_remove)} completed tasks")
        return len(to_remove)