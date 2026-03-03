# Copyright (c) 2025 pehmc. MIT License.
# See LICENSE file in the project root for full license information.

import os
import sys
import shutil
import argparse
import logging
from pathlib import Path
from typing import List, Optional, Set

# 设置日志
logging.basicConfig(level=logging.INFO, format='%(message)s')
logger = logging.getLogger(__name__)


class SeedPool:
    """管理AFL++种子文件的类"""
    
    def __init__(self, output_dir: str):
        """
        初始化AFL种子管理器
        
        Args:
            output_dir: AFL++输出目录
        """
        self.output_dir = Path(output_dir)
        self.seed_dir = self.output_dir / "addseeds" / "queue"
        self._validate_output_dir()
        
    def _validate_output_dir(self) -> None:
        """验证输出目录是否为有效的AFL++输出目录"""
        if not self.output_dir.exists():
            raise ValueError(f"错误: 目录 '{self.output_dir}' 不存在")
        
        if not self.output_dir.is_dir():
            raise ValueError(f"错误: '{self.output_dir}' 不是目录")
        
        # 检查是否存在fuzzer_stats文件
        fuzzer_stats_files = list(self.output_dir.glob("*/fuzzer_stats"))
        if not fuzzer_stats_files:
            raise ValueError(
                f"错误: '{self.output_dir}' 不是有效的AFL++输出目录 "
                f"(未找到fuzzer_stats文件)"
            )
        
        # 创建种子目录
        self.seed_dir.mkdir(parents=True, exist_ok=True)
        if not self.seed_dir.is_dir():
            raise ValueError(f"错误: 无法创建目录 '{self.seed_dir}'")
    
    def _get_next_available_id(self) -> int:
        """
        获取下一个可用的种子ID
        
        Returns:
            可用的种子ID
        """
        existing_ids = set()
        
        # 收集已存在的ID
        for file in self.seed_dir.glob("id:*"):
            filename = file.name
            # 提取ID部分 (格式: id:000001,...)
            if filename.startswith("id:"):
                id_part = filename.split(",")[0]  # id:000001
                id_str = id_part[3:]  # 000001
                try:
                    existing_ids.add(int(id_str))
                except ValueError:
                    continue
        
        # 找到下一个可用的ID
        next_id = 0
        while next_id in existing_ids:
            next_id += 1
            
        return next_id
    
    def _generate_afl_filename(self, original_file: Path, seed_id: int) -> str:
        """
        生成AFL格式的文件名
        
        Args:
            original_file: 原始文件路径
            seed_id: 种子ID
            
        Returns:
            AFL格式的文件名
        """
        # 格式: id:{六位数},time:0,execs:0,orig:{原始文件名}
        return f"id:{seed_id:06d},time:0,execs:0,orig:{original_file.name}"
    
    def add_seed_file(self, seed_file: Path) -> str:
        """
        添加单个种子文件
        
        Args:
            seed_file: 种子文件路径
            
        Returns:
            新文件的路径
            
        Raises:
            ValueError: 如果文件不存在或不是文件
        """
        if not seed_file.exists():
            raise ValueError(f"错误: 文件 '{seed_file}' 不存在")
        
        if not seed_file.is_file():
            raise ValueError(f"错误: '{seed_file}' 不是文件")
        
        # 获取下一个可用ID
        seed_id = self._get_next_available_id()
        
        # 生成目标文件名和路径
        target_filename = self._generate_afl_filename(seed_file, seed_id)
        target_path = self.seed_dir / target_filename
        
        # 复制文件
        shutil.copy2(seed_file, target_path)
        logger.info(f"已添加: {seed_file} -> {target_path.name}")
        
        return str(target_path)
    
    def add_seed_directory(self, seed_dir: Path) -> List[str]:
        """
        添加目录中的所有种子文件
        
        Args:
            seed_dir: 种子目录路径
            
        Returns:
            新添加的文件路径列表
            
        Raises:
            ValueError: 如果目录不存在
        """
        if not seed_dir.exists():
            raise ValueError(f"错误: 目录 '{seed_dir}' 不存在")
        
        if not seed_dir.is_dir():
            raise ValueError(f"错误: '{seed_dir}' 不是目录")
        
        added_files = []
        
        # 递归查找所有文件
        for file_path in seed_dir.rglob("*"):
            if file_path.is_file():
                try:
                    new_file = self.add_seed_file(file_path)
                    added_files.append(new_file)
                except Exception as e:
                    logger.warning(f"跳过文件 '{file_path}': {e}")
        
        return added_files
    
    def add_seeds(self, seed_paths: List[str], recursive: bool = False) -> List[str]:
        """
        添加种子文件或目录
        
        Args:
            seed_paths: 种子文件或目录路径列表
            recursive: 是否递归处理目录
            
        Returns:
            新添加的文件路径列表
        """
        added_files = []
        
        for seed_path_str in seed_paths:
            seed_path = Path(seed_path_str)
            
            if not seed_path.exists():
                logger.warning(f"警告: 路径 '{seed_path}' 不存在，已跳过")
                continue
            
            if seed_path.is_file():
                try:
                    new_file = self.add_seed_file(seed_path)
                    added_files.append(new_file)
                except Exception as e:
                    logger.error(f"错误: 无法添加文件 '{seed_path}': {e}")
            
            elif seed_path.is_dir():
                try:
                    if recursive:
                        dir_files = self.add_seed_directory(seed_path)
                    else:
                        # 仅处理目录中的直接文件
                        dir_files = []
                        for file_path in seed_path.iterdir():
                            if file_path.is_file():
                                try:
                                    new_file = self.add_seed_file(file_path)
                                    dir_files.append(new_file)
                                except Exception as e:
                                    logger.error(f"错误: 无法添加文件 '{file_path}': {e}")
                    
                    added_files.extend(dir_files)
                except Exception as e:
                    logger.error(f"错误: 无法处理目录 '{seed_path}': {e}")
        
        return added_files
    
    def get_seed_count(self) -> int:
        """获取当前种子目录中的种子文件数量"""
        count = 0
        for _ in self.seed_dir.glob("id:*"):
            count += 1
        return count
    
    def list_seeds(self) -> List[str]:
        """列出所有种子文件"""
        return sorted([str(f.name) for f in self.seed_dir.glob("id:*")])


