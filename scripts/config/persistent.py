# Copyright (c) 2025 pehmc. MIT License.
# See LICENSE file in the project root for full license information.

#!/usr/bin/env python3
"""
afl-persistent-config - 永久配置系统至高性能模糊测试状态
警告：此操作会降低系统安全性！
"""

import os
import sys
import platform
import argparse
import subprocess
import re
import shutil
from pathlib import Path
from typing import NoReturn


class AFLPersistentConfig:
    """AFL++ 高性能模糊测试系统永久配置工具"""

    # Linux sysctl 配置内容
    LINUX_SYSCTL_CONFIG = """\
kernel.core_uses_pid=0
kernel.core_pattern=core
kernel.randomize_va_space=0
kernel.sched_child_runs_first=1
kernel.sched_autogroup_enabled=1
kernel.sched_migration_cost_ns=50000000
kernel.sched_latency_ns=250000000
vm.swappiness=10
"""

    # macOS LaunchDaemon plist 内容
    MACOS_PLIST_CONTENT = """\
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>shmemsetup</string>
    <key>UserName</key>
    <string>root</string>
    <key>GroupName</key>
    <string>wheel</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/sbin/sysctl</string>
      <string>-w</string>
      <string>kern.sysv.shmmax=524288000</string>
      <string>kern.sysv.shmmin=1</string>
      <string>kern.sysv.shmmni=128</string>
      <string>kern.sysv.shmseg=48</string>
      <string>kern.sysv.shmall=131072000</string>
    </array>
    <key>KeepAlive</key>
    <false/>
    <key>RunAtLoad</key>
    <true/>
  </dict>
</plist>
"""

    # 用于 GRUB 的安全缓解措施禁用选项列表
    MITIGATION_OPTIONS = [
        "ibpb=off",
        "ibrs=off",
        "kpti=off",
        "l1tf=off",
        "spec_rstack_overflow=off",
        "mds=off",
        "nokaslr",
        "no_stf_barrier",
        "noibpb",
        "noibrs",
        "pcid",
        "nopti",
        "nospec_store_bypass_disable",
        "nospectre_v1",
        "nospectre_v2",
        "pcid=on",
        "pti=off",
        "spec_store_bypass_disable=off",
        "spectre_v2=off",
        "stf_barrier=off",
        "srbds=off",
        "noexec=off",
        "noexec32=off",
        "tsx=on",
        "tsx_async_abort=off",
        "mitigations=off",
        "audit=0",
        "hardened_usercopy=off",
        "ssbd=force-off"
    ]

    def __init__(self):
        self.platform = platform.system()
        self.arch = platform.machine()
        self.is_root = os.geteuid() == 0
        self.sysctl_d_dir = Path("/etc/sysctl.d")
        self.grub_default = Path("/etc/default/grub")
        self.macos_plist_path = Path("/Library/LaunchDaemons/shm_setup.plist")

    def print_warning(self) -> None:
        """打印安全警告并请求确认"""
        print("\n警告：此脚本将对系统进行永久性配置更改，以提升模糊测试性能。")
        print("      因此，系统对攻击的防护能力会降低！")
        print("      如果使用此脚本，请设置强防火墙规则，并仅开放 SSH 网络服务！\n")
        
        try:
            answer = input('输入 "YES" 以继续: ').strip()
        except KeyboardInterrupt:
            print("\n已取消")
            sys.exit(1)
        
        if answer != "YES":
            print("输入不是 YES，中止...")
            sys.exit(1)

    def die(self, message: str, code: int = 1) -> NoReturn:
        """输出错误信息并退出"""
        print(f"错误: {message}", file=sys.stderr)
        sys.exit(code)

    def check_requirements(self) -> None:
        """检查运行环境和权限"""
        # 检查平台支持
        if self.platform not in ["Darwin", "Linux"]:
            self.die(f"不支持的操作系统平台 \"{self.platform}\"，目前仅支持 Linux 和 macOS")

        # 检查 root 权限
        if not self.is_root:
            self.die("需要 root 权限，请使用 sudo 运行")

        print("权限检查通过。")

        # macOS 额外 SIP 检查
        if self.platform == "Darwin":
            self._check_macos_sip()

    def _check_macos_sip(self) -> None:
        """检查 macOS 系统完整性保护 (SIP) 状态"""
        try:
            result = subprocess.run(
                ["csrutil", "status"],
                capture_output=True,
                text=True,
                check=False
            )
            if "disabled" not in result.stdout.lower():
                self.die(
                    "SIP 需要被禁用。请重启，按住 Command-R 进入恢复模式，"
                    "打开终端并执行 'csrutil disable'"
                )
        except FileNotFoundError:
            self.die("无法执行 csrutil，可能不是 macOS 或工具缺失")

    def apply_macos(self) -> None:
        """应用 macOS 系统配置"""
        print("正在应用 macOS 高性能模糊测试配置...")

        # 1. 安装 LaunchDaemon plist
        print(f"安装 {self.macos_plist_path}")
        try:
            self.macos_plist_path.write_text(self.MACOS_PLIST_CONTENT)
            self.macos_plist_path.chmod(0o644)  # 默认权限
        except Exception as e:
            self.die(f"无法写入 plist 文件: {e}")

        # 2. 禁用系统级 ASLR (仅 Intel)
        if self.arch == "x86_64":
            print("禁用系统级 ASLR")
            try:
                subprocess.run(
                    ["nvram", "boot-args=no_aslr=1"],
                    check=True
                )
            except subprocess.CalledProcessError as e:
                self.die(f"设置 nvram 失败: {e}")
        else:
            print("注意: 当前架构为 ARM64，目前未知如何全局禁用 ASLR，如有方法请告知。")

        print("\n配置完成！请重启系统以启用所有设置。")

    def apply_linux(self) -> None:
        """应用 Linux 系统配置"""
        print("正在应用 Linux 高性能模糊测试配置...")

        # 1. 创建 sysctl.d 配置文件
        if not self.sysctl_d_dir.is_dir():
            print(f"错误: {self.sysctl_d_dir} 目录不存在，无法安装共享内存配置", file=sys.stderr)
        else:
            conf_file = self.sysctl_d_dir / "99-fuzzing.conf"
            if not conf_file.exists():
                print(f"安装 {conf_file}")
                try:
                    conf_file.write_text(self.LINUX_SYSCTL_CONFIG)
                except Exception as e:
                    print(f"警告: 无法写入 {conf_file}: {e}", file=sys.stderr)
            else:
                print(f"{conf_file} 已存在，跳过创建")

        # 2. 修改 GRUB 引导参数以禁用安全缓解措施
        if not self.grub_default.exists():
            print(f"错误: {self.grub_default} 不存在，无法设置引导参数", file=sys.stderr)
        else:
            self._modify_grub_config()

    def _modify_grub_config(self) -> None:
        """修改 /etc/default/grub，添加禁用安全缓解措施的内核参数"""
        # 读取当前配置
        try:
            with open(self.grub_default, 'r') as f:
                lines = f.readlines()
        except Exception as e:
            print(f"警告: 无法读取 {self.grub_default}: {e}", file=sys.stderr)
            return

        # 需要处理的变量列表
        targets = ["GRUB_CMDLINE_LINUX_DEFAULT", "GRUB_CMDLINE_LINUX"]
        modified = False

        # 生成需要添加的选项字符串
        new_options = " ".join(self.MITIGATION_OPTIONS)

        for target in targets:
            # 查找目标行
            pattern = re.compile(rf'^{target}=')
            for i, line in enumerate(lines):
                if pattern.match(line):
                    # 提取当前值（可能带引号）
                    match = re.search(rf'{target}=["\']?(.*?)["\']?\s*(?:#.*)?$', line)
                    if not match:
                        continue
                    current_value = match.group(1).strip()
                    # 检查是否已包含所有必要选项（简单判断 - 避免重复添加）
                    # 我们检查是否包含常见的禁用选项，如果已包含则跳过
                    if "noibrs" in current_value and "nopti" in current_value and "mitigations=off" in current_value:
                        print(f"{target} 已包含性能引导选项，跳过")
                        continue
                    
                    # 构造新值：原有选项 + 新选项
                    new_value = f"{current_value} {new_options}".strip()
                    # 更新行
                    lines[i] = f'{target}="{new_value}"\n'
                    modified = True
                    print(f"更新 {target} 引导选项")
                    break

        if modified:
            # 备份原始文件
            backup_path = self.grub_default.with_suffix('.conf.bak')
            try:
                shutil.copy2(self.grub_default, backup_path)
                print(f"已备份原始配置至 {backup_path}")
            except Exception as e:
                print(f"警告: 无法备份 {self.grub_default}: {e}", file=sys.stderr)

            # 写入修改
            try:
                with open(self.grub_default, 'w') as f:
                    f.writelines(lines)
                print("GRUB 配置已更新")
            except Exception as e:
                print(f"错误: 无法写入 {self.grub_default}: {e}", file=sys.stderr)
                # 尝试恢复备份
                if backup_path.exists():
                    shutil.copy2(backup_path, self.grub_default)
                    print("已恢复原始配置")
                return
        else:
            print("未找到需要更新的 GRUB 配置行")

    def run(self) -> None:
        """主执行流程"""
        # 显示警告并确认
        self.print_warning()

        # 检查环境
        self.check_requirements()

        # 根据平台执行配置
        if self.platform == "Darwin":
            self.apply_macos()
        elif self.platform == "Linux":
            self.apply_linux()
            print("\n配置完成！请重启系统以启用所有设置。")
        else:
            # 理论上不会走到这里
            self.die(f"未处理的平台: {self.platform}")


def parse_args():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(
        description='afl-persistent-config - 永久配置系统至高性能模糊测试状态',
        epilog='注意：此脚本会降低系统安全性，仅在专用模糊测试机上使用！',
        add_help=False  # 自定义帮助输出以匹配原脚本风格
    )
    parser.add_argument('-h', '--help', action='store_true', dest='help',
                        help='显示帮助信息')
    parser.add_argument('-hh', action='store_true', dest='hh',
                        help='(兼容) 显示帮助信息')
    # 不接受其他参数
    args, unknown = parser.parse_known_args()

    if unknown:
        print(f"错误: 未知选项: {' '.join(unknown)}", file=sys.stderr)
        sys.exit(1)

    if args.help or args.hh:
        print('afl-persistent-config\n')
        print(__file__)
        print('\nafl-persistent-config 没有命令行选项\n')
        print('afl-persistent-config 永久重新配置系统到高性能模糊测试状态。')
        print('警告：这会降低系统安全性！')
        print('\n注意：还有 afl-system-config 用于设置额外的运行时配置选项。')
        sys.exit(0)

    return args


def main():
    """主函数"""
    parse_args()
    config = AFLPersistentConfig()
    config.run()


if __name__ == '__main__':
    main()