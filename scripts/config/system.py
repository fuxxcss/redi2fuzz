# Copyright (c) 2025 pehmc. MIT License.
# See LICENSE file in the project root for full license information.

import os
import sys
import time
import platform
import subprocess
import shutil
import re
from pathlib import Path


def print_help():
    print("afl-system-config by Marc Heuse <mh@mh-sec.de>\n")
    print(sys.argv[0])
    print()
    print("afl-system-config has no command line options")
    print()
    print("afl-system-config reconfigures the system to a high performance fuzzing state.")
    print("WARNING: this reduces the security of the system!")
    print()
    print("Note that there is also afl-persistent-config which sets additional permanent")
    print("configuration options.")
    sys.exit(0)


def main():
    # Argument parsing
    args = sys.argv[1:]
    if any(arg in args for arg in ("-h", "-hh", "--help")):
        print_help()
    if args:
        print(f"ERROR: Unknown option(s): {' '.join(args)}", file=sys.stderr)
        sys.exit(1)

    # Banner
    print("This reconfigures the system to have a better fuzzing performance.")
    print("WARNING: this reduces the security of the system!\n")

    # Root check (warning only)
    if os.geteuid() != 0:
        print("Warning: you need to be root to run this!", file=sys.stderr)
        time.sleep(1)

    time.sleep(1)

    # Platform detection
    system = platform.system()
    done = False

    if system == "Linux":
        done = configure_linux()
    elif system == "FreeBSD":
        done = configure_freebsd()
    elif system == "OpenBSD":
        done = configure_openbsd()
    elif system == "DragonFly":
        done = configure_dragonfly()
    elif system == "NetBSD":
        done = configure_netbsd()
    elif system == "Darwin":
        done = configure_darwin()
    elif system == "Haiku":
        done = configure_haiku()
    else:
        print(f"Error: Unknown platform: {system}", file=sys.stderr)
        sys.exit(1)

    if not done:
        print(f"Error: Configuration for {system} failed or incomplete.", file=sys.stderr)
        sys.exit(1)

    sys.exit(0)


def run_cmd(cmd, ignore_errors=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL):
    """Run a shell command (list) with suppressed output by default."""
    try:
        subprocess.run(cmd, stdout=stdout, stderr=stderr, check=not ignore_errors)
    except subprocess.SubprocessError as e:
        if not ignore_errors:
            raise
        # Silently ignore


def sysctl_set(key, value, ignore=True):
    """Run sysctl -w key=value with output suppressed."""
    run_cmd(["sysctl", "-w", f"{key}={value}"], ignore_errors=ignore)


def write_sysfs(path, content):
    """Write a string to a sysfs file; silently ignore errors."""
    try:
        Path(path).write_text(content)
    except (OSError, IOError):
        pass


def configure_linux():
    print("Configuring Linux for fuzzing...")

    # kernel core pattern settings
    sysctl_set("kernel.core_uses_pid", 0)
    # Arch Linux special: core_pattern must be empty
    if Path("/etc/arch-release").exists():
        sysctl_set("kernel.core_pattern", "")
    else:
        sysctl_set("kernel.core_pattern", "core")

    # ASLR
    sysctl_set("kernel.randomize_va_space", 0)

    # Scheduler tweaks
    sysctl_set("kernel.sched_child_runs_first", 1)
    sysctl_set("kernel.sched_autogroup_enabled", 1)
    sysctl_set("kernel.sched_migration_cost_ns", 50000000, ignore=True)
    sysctl_set("kernel.sched_latency_ns", 250000000, ignore=True)

    # VM
    sysctl_set("vm.swappiness", 10, ignore=True)

    # Transparent hugepages
    write_sysfs("/sys/kernel/mm/transparent_hugepage/enabled", "never")

    # CPU frequency scaling governor
    for path in [
        "/sys/devices/system/cpu/cpufreq/scaling_governor",
        "/sys/devices/system/cpu/cpufreq/policy0/scaling_governor",
        "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor",
    ]:
        if Path(path).exists():
            # For the pattern path, we need to write to all matching files
            if "policy0" in path:
                # Write to all policy*/scaling_governor
                for p in Path("/sys/devices/system/cpu/cpufreq").glob("policy*/scaling_governor"):
                    write_sysfs(str(p), "performance")
            elif "cpu0" in path:
                for p in Path("/sys/devices/system/cpu").glob("cpu*/cpufreq/scaling_governor"):
                    write_sysfs(str(p), "performance")
            else:
                write_sysfs(path, "performance")

    # Turbo/boost settings
    write_sysfs("/sys/devices/system/cpu/intel_pstate/no_turbo", "0")
    write_sysfs("/sys/devices/system/cpu/cpufreq/boost", "1")
    write_sysfs("/sys/devices/system/cpu/intel_pstate/max_perf_pct", "100")

    # Disable audit
    if shutil.which("auditctl"):
        run_cmd(["auditctl", "-a", "never,task"], ignore_errors=True)

    print("Settings applied.\n")

    # Check for kernel command line security mitigations
    try:
        # Check dmesg for mitigation disabling options
        dmesg = subprocess.run(["dmesg"], capture_output=True, text=True, check=True).stdout
        if not re.search(r'noibrs|pcid|nopti', dmesg):
            print("It is recommended to boot the kernel with lots of security off - if you are running a machine that is in a secured network - so set this:")
            print('  /etc/default/grub:GRUB_CMDLINE_LINUX_DEFAULT="ibpb=off ibrs=off kpti=off l1tf=off spec_rstack_overflow=off mds=off nokaslr no_stf_barrier noibpb noibrs pcid nopti nospec_store_bypass_disable nospectre_v1 nospectre_v2 pcid=on pti=off spec_store_bypass_disable=off spectre_v2=off stf_barrier=off srbds=off noexec=off noexec32=off tsx=on tsx=on tsx_async_abort=off mitigations=off audit=0 hardened_usercopy=off ssbd=force-off"')
            print()
    except (subprocess.SubprocessError, FileNotFoundError):
        pass

    print('If you run fuzzing instances in docker, run them with "--security-opt seccomp=unconfined" for more speed.\n')
    return True


def configure_freebsd():
    print("Configuring FreeBSD for fuzzing...")
    sysctl_set("kern.elf32.aslr.enable", 0)
    sysctl_set("kern.elf64.aslr.enable", 0)
    print("Settings applied.\n")
    print("In order to suppress core file generation during fuzzing it is recommended to set")
    print("me:\\")
    print("\t:coredumpsize=0:")
    print("in the ~/.login_conf file for the user used for fuzzing.\n")
    print("It is recommended to boot the kernel with lots of security off - if you are running a machine that is in a secured network - so set this:")
    print("  sysctl hw.ibrs_disable=1")
    print("Setting 'kern.pmap.pg_ps_enabled=0' into /boot/loader.conf might be helpful too.\n")
    return True


def configure_openbsd():
    print("Configuring OpenBSD for fuzzing...")
    # doas sysctl vm.malloc_conf=
    run_cmd(["doas", "sysctl", "vm.malloc_conf="], ignore_errors=True)
    print("Freecheck on allocation in particular can be detrimental to performance.")
    print("Also we might not want necessarily to abort at any allocation failure.")
    print("System security features cannot be disabled on OpenBSD.\n")
    return True


def configure_dragonfly():
    print("Configuring DragonFly BSD for fuzzing...")
    print("In order to suppress core file generation during fuzzing it is recommended to set")
    print("me:\\")
    print("\t:coredumpsize=0:")
    print("in the ~/.login_conf file for the user used for fuzzing.\n")
    return True


def configure_netbsd():
    print("Configuring NetBSD for fuzzing...")
    sysctl_set("security.models.extensions.user_set_cpu_affinity", 1)
    print("Settings applied.\n")
    return True


def configure_darwin():
    print("Configuring macOS (Darwin) for fuzzing...")
    sysctl_set("kern.sysv.shmmax", 524288000)
    sysctl_set("kern.sysv.shmmin", 1)
    sysctl_set("kern.sysv.shmseg", 48)
    sysctl_set("kern.sysv.shmall", 131072000)
    print("Settings applied.\n")

    # Crash reporter unloading
    # Check if we have SUDO_USER environment variable
    sudo_user = os.environ.get("SUDO_USER")
    if sudo_user:
        try:
            # User launch agent
            subprocess.run(
                ["sudo", "-u", sudo_user, "launchctl", "unload", "-w",
                 "/System/Library/LaunchAgents/com.apple.ReportCrash.plist"],
                stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
            )
            # System daemon
            subprocess.run(
                ["launchctl", "unload", "-w",
                 "/System/Library/LaunchDaemons/com.apple.ReportCrash.Root.plist"],
                stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
            )
            print("Unloaded the default crash reporter.\n")
        except Exception:
            pass

    print("It is recommended to disable System Integrity Protection for increased performance.")
    print("See: https://developer.apple.com/documentation/security/disabling_and_enabling_system_integrity_protection\n")
    return True


def configure_haiku():
    print("Configuring Haiku for fuzzing...")
    # ~/config/settings/system/debug_server/settings
    home = Path.home()
    debug_server_dir = home / "config" / "settings" / "system" / "debug_server"
    settings_file = debug_server_dir / "settings"

    if not debug_server_dir.exists():
        debug_server_dir.mkdir(parents=True)

    # Check if already configured
    need_change = True
    if settings_file.exists():
        content = settings_file.read_text()
        if re.search(r"default_action\s+kill", content):
            need_change = False
            print("Nothing to do")
        else:
            # Replace 'default_action user' with 'default_action kill'
            new_content = re.sub(
                r"default_action\s+user",
                "default_action kill",
                content
            )
            settings_file.write_text(new_content)
            print("Settings applied.")
    else:
        settings_file.write_text("default_action kill\n")
        print("Settings applied.")

    if need_change:
        print()
    return True


if __name__ == "__main__":
    main()