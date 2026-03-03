# Copyright (c) 2025 pehmc. MIT License.
# See LICENSE file in the project root for full license information.

import argparse
import os
import sys
import time
import tempfile
import subprocess
import shutil
from pathlib import Path
import re
from datetime import timedelta
import signal


# ----------------------------------------------------------------------
#  ANSI color codes (used if NO_COLOR is not set and stdout is a tty)
# ----------------------------------------------------------------------
COLOR_RED = "\033[91m"
COLOR_GREEN = "\033[92m"
COLOR_YELLOW = "\033[93m"
COLOR_BLUE = "\033[94m"
COLOR_NC = "\033[0m"
COLOR_RESET = COLOR_NC


def fmt_duration(seconds: int) -> str:
    """Convert seconds to human readable string (days/hours/minutes/seconds)."""
    if seconds <= 0:
        return "0 seconds"
    if seconds == 1:
        return "1 second"
    delta = timedelta(seconds=seconds)
    days = delta.days
    hours = delta.seconds // 3600
    minutes = (delta.seconds % 3600) // 60
    secs = delta.seconds % 60

    if days > 0:
        return f"{days} days, {hours} hours"
    if hours > 0:
        return f"{hours} hours, {minutes} minutes"
    if minutes > 0:
        return f"{minutes} minutes, {secs} seconds"
    return f"{secs} seconds"


def get_current_time() -> int:
    """Return current time in seconds (epoch)."""
    return int(time.time())


def read_fuzzer_stats(stats_path: Path) -> dict:
    """
    Parse a fuzzer_stats file and return a dictionary of key/value pairs.
    Similar to the sed sourcing in the original shell script.
    """
    stats = {}
    if not stats_path.exists():
        return stats

    with open(stats_path, "r") as f:
        for line in f:
            line = line.rstrip("\n")
            if not line or line.startswith("command_line"):
                continue
            if ":" not in line:
                continue
            key, value = line.split(":", 1)
            key = key.strip()
            value = value.strip()
            stats[key] = value
    return stats


def is_process_alive(pid: int) -> bool:
    """Check if a process with the given PID exists and is running."""
    try:
        os.kill(pid, 0)
        return True
    except (ProcessLookupError, PermissionError, OSError):
        return False


def get_cpu_mem_usage(pid: int) -> tuple:
    """
    Return (cpu_percent, mem_percent) for the given PID.
    Uses `ps` command (similar to original script).
    """
    cpu = "0.0"
    mem = "0.0"
    try:
        output = subprocess.check_output(
            ["ps", "-p", str(pid), "-o", "pcpu=,pmem="],
            stderr=subprocess.DEVNULL,
            universal_newlines=True,
        ).strip()
        if output:
            parts = output.split()
            if len(parts) >= 2:
                cpu = parts[0]
                mem = parts[1]
    except (subprocess.SubprocessError, IndexError):
        pass
    return cpu, mem


def has_fuser() -> bool:
    """Check if the `fuser` command is available."""
    return shutil.which("fuser") is not None


def get_fuser_pids(directory: Path) -> set:
    """
    Run `fuser -v` on a directory and return a set of PIDs of afl-fuzz processes.
    If fuser is not available, return empty set.
    """
    pids = set()
    if not has_fuser():
        return pids
    try:
        output = subprocess.check_output(
            ["fuser", "-v", str(directory)],
            stderr=subprocess.DEVNULL,
            universal_newlines=True,
        )
        for line in output.splitlines():
            if "afl-fuzz" in line:
                # format: user     PID ...; we just need PID
                parts = line.split()
                for part in parts:
                    if part.isdigit():
                        pids.add(int(part))
                        break
    except subprocess.SubprocessError:
        pass
    return pids


def main():
    parser = argparse.ArgumentParser(
        description="Status check tool for afl-fuzz",
        add_help=False,  # we handle -h/-hh ourselves to match original
    )
    parser.add_argument("-d", dest="process_dead", action="store_true",
                        help="include dead fuzzer stats")
    parser.add_argument("-m", dest="minimal_only", action="store_true",
                        help="just show minimal stats")
    parser.add_argument("-n", dest="no_color", action="store_true",
                        help="no color output")
    parser.add_argument("-s", dest="summary_only", action="store_true",
                        help="skip details and output summary results only")
    parser.add_argument("directory", nargs="?", help="AFL output directory")
    args, unknown = parser.parse_known_args()

    # Handle -h / -hh / --help specially (original script prints custom help)
    if unknown or args.directory in ("-h", "-hh", "--help") or not args.directory:
        print(f"{sys.argv[0]} status check tool for afl-fuzz by Michal Zalewski\n")
        print("Usage: {} [-d] [-m] [-n] [-s] afl_output_directory\n".format(sys.argv[0]))
        print("Options:")
        print("  -d  -  include dead fuzzer stats")
        print("  -m  -  just show minimal stats")
        print("  -n  -  no color output")
        print("  -s  -  skip details and output summary results only")
        sys.exit(1)

    # ------------------------------------------------------------------
    # Environment & colors
    # ------------------------------------------------------------------
    if not args.no_color and sys.stdout.isatty():
        RED = COLOR_RED
        GREEN = COLOR_GREEN
        YELLOW = COLOR_YELLOW
        BLUE = COLOR_BLUE
        NC = COLOR_NC
        RESET = COLOR_RESET
    else:
        RED = GREEN = YELLOW = BLUE = NC = RESET = ""

    # ------------------------------------------------------------------
    # Directory validation
    # ------------------------------------------------------------------
    sync_dir = Path(args.directory).resolve()
    if not sync_dir.is_dir():
        print(f"[-] Error: directory '{sync_dir}' does not exist.", file=sys.stderr)
        sys.exit(1)

    # Check if it's an individual output directory (contains 'queue')
    if (sync_dir / "queue").is_dir():
        print("[-] Error: parameter is an individual output directory, not a sync dir.", file=sys.stderr)
        sys.exit(1)

    # ------------------------------------------------------------------
    # Initialization
    # ------------------------------------------------------------------
    if not args.minimal_only:
        print(f"{sys.argv[0]} status check tool for afl-fuzz by Michal Zalewski\n")

    current_time = get_current_time()

    # Temporary file is not needed in Python; we'll just keep data in memory.

    # Counters
    alive_cnt = 0
    dead_cnt = 0
    start_cnt = 0
    total_time = 0
    total_execs = 0
    total_eps = 0          # sum of execs/sec (lifetime)
    total_eplm = 0         # sum of execs per last minute
    total_crashes = 0
    total_hangs = 0
    total_pfav = 0
    total_pending = 0
    total_coverage = None
    total_wcop = []        # list of cycles_wo_finds (for slash-separated output)
    total_last_find = 0

    # Collect individual fuzzer data for summary
    fuzzers = []

    # ------------------------------------------------------------------
    # Iterate over all fuzzer instances
    # ------------------------------------------------------------------
    # Find all directories that contain fuzzer_setup (case-insensitive)
    # Using glob: **/fuzzer_setup (maxdepth 2)
    for setup_path in sync_dir.glob("*/fuzzer_setup"):
        fuzzer_dir = setup_path.parent
        stats_path = fuzzer_dir / "fuzzer_stats"

        if stats_path.is_file():
            stats = read_fuzzer_stats(stats_path)

            # Required fields
            afl_banner = stats.get("afl_banner", "unknown")
            fuzzer_pid_str = stats.get("fuzzer_pid", "")
            try:
                fuzzer_pid = int(fuzzer_pid_str)
            except (ValueError, TypeError):
                fuzzer_pid = None

            run_time = int(stats.get("run_time", 0))
            execs_done = int(stats.get("execs_done", 0))
            corpus_count = int(stats.get("corpus_count", 0))
            cur_item = int(stats.get("cur_item", 0))
            cycles_done = int(stats.get("cycles_done", 0))
            cycles_wo_finds = stats.get("cycles_wo_finds", None)
            last_find = int(stats.get("last_find", 0))
            last_crash = int(stats.get("last_crash", 0))
            last_hang = int(stats.get("last_hang", 0))
            saved_crashes = int(stats.get("saved_crashes", 0))
            saved_hangs = int(stats.get("saved_hangs", 0))
            pending_total = int(stats.get("pending_total", 0))
            pending_favs = int(stats.get("pending_favs", 0))
            execs_ps_last_min = stats.get("execs_ps_last_min", "0")
            bitmap_cvg = stats.get("bitmap_cvg", "0%")
            coverage = int(bitmap_cvg.strip("%"))
            exec_timeout = int(stats.get("exec_timeout", 0))

            # Coverage: keep maximum
            if total_coverage is None or coverage > total_coverage:
                total_coverage = coverage

            # cycles_wo_finds for summary
            if cycles_wo_finds is not None:
                total_wcop.append(cycles_wo_finds)

            # Determine if alive, dead, or starting
            is_alive = False
            is_dead = False
            is_starting = False

            if fuzzer_pid is not None and is_process_alive(fuzzer_pid):
                is_alive = True
            else:
                # Process is not running. Check if maybe it's starting up.
                # fuzzer_setup newer than fuzzer_stats?
                if setup_path.stat().st_mtime > stats_path.stat().st_mtime:
                    # Try to detect if afl-fuzz is still using the directory (via fuser)
                    pids = get_fuser_pids(fuzzer_dir)
                    if pids:
                        # Found afl-fuzz using the directory -> starting
                        is_starting = True
                        if args.process_dead:
                            # include it as alive? The script: when starting, it increments start_cnt
                            # and if PROCESS_DEAD is set, it continues (doesn't skip) but counts it as alive.
                            # Actually: if PROCESS_DEAD is empty, it 'continue's. So starting instances
                            # are only processed if PROCESS_DEAD is set. We'll handle in counters later.
                            is_starting = True
                    else:
                        is_dead = True
                else:
                    is_dead = True

            # Counters
            if is_alive:
                alive_cnt += 1
            elif is_starting:
                start_cnt += 1
                if not args.process_dead:
                    # Skip this instance entirely if we don't process dead/starting
                    continue
            elif is_dead:
                dead_cnt += 1
                if not args.process_dead:
                    continue

            # If we are here, the instance is either alive, or we are including dead/starting with -d
            # For dead/starting, we still count them in totals but skip some calculations (execs per last min)
            if not is_alive and not is_starting:
                # dead: no execs_ps_last_min, etc.
                pass

            # Update totals (do for all, including dead/starting)
            total_time += run_time
            total_execs += execs_done
            total_crashes += saved_crashes
            total_hangs += saved_hangs
            total_pending += pending_total
            total_pfav += pending_favs

            # Execs per second (lifetime) - avoid division by zero
            if run_time > 0:
                exec_sec = execs_done // run_time
            else:
                exec_sec = 0
            total_eps += exec_sec

            # Execs per last minute (only for alive; for dead script sets to 0)
            if is_alive:
                try:
                    eplm = int(float(execs_ps_last_min))
                except ValueError:
                    eplm = 0
            else:
                eplm = 0
            total_eplm += eplm

            # Track most recent find time (all instances)
            if last_find > total_last_find:
                total_last_find = last_find

            # ------------------------------------------------------------------
            # Individual fuzzer output (unless summary_only)
            # ------------------------------------------------------------------
            if not args.summary_only:
                run_days = run_time // 86400
                run_hours = (run_time % 86400) // 3600
                dir_name = fuzzer_dir.name

                print(f">>> {afl_banner} instance: {dir_name} ({run_days} days, {run_hours} hrs) "
                      f"fuzzer PID: {fuzzer_pid_str} <<<\n")

                if is_dead:
                    print("  Instance is dead or running remotely, skipping.\n")
                    continue
                if is_starting:
                    print("  Instance is still starting up, skipping.\n")
                    continue

                # Warnings
                timeout_perc = (exec_timeout * 100) // execs_done if execs_done > 0 else 0
                if timeout_perc >= 10:
                    print(f"  {RED}timeout_ratio {timeout_perc}%{NC}")

                if exec_sec == 0:
                    print(f"  {YELLOW}no data yet, 0 execs/sec{NC}")
                elif exec_sec < 100:
                    print(f"  {RED}slow execution, {exec_sec} execs/sec{NC}")

                # Time since last find/crash/hang
                if last_find > 0:
                    find_fmt = fmt_duration(current_time - last_find)
                else:
                    find_fmt = f"{RED}none seen yet{NC}"
                if last_crash > 0:
                    crash_fmt = fmt_duration(current_time - last_crash)
                else:
                    crash_fmt = "none seen yet"
                if last_hang > 0:
                    hang_fmt = fmt_duration(current_time - last_hang)
                else:
                    hang_fmt = "none seen yet"

                # cycles_wo_finds with color
                cwop = cycles_wo_finds if cycles_wo_finds is not None else "not available"
                if cycles_wo_finds is not None:
                    try:
                        cw_val = int(cycles_wo_finds)
                        if cw_val > 50:
                            cwop = f"{RED}{cw_val}{NC}"
                        elif cw_val > 10:
                            cwop = f"{YELLOW}{cw_val}{NC}"
                        else:
                            cwop = str(cw_val)
                    except ValueError:
                        pass

                print(f"  last_find       : {find_fmt}")
                print(f"  last_crash      : {crash_fmt}")
                if not args.minimal_only:
                    print(f"  last_hang       : {hang_fmt}")
                    print(f"  cycles_wo_finds : {cwop}")
                print(f"  coverage        : {coverage}%")

                if not args.minimal_only:
                    cpu, mem = get_cpu_mem_usage(fuzzer_pid)
                    print(f"  cpu usage {cpu}%, memory usage {mem}%")

                path_perc = (cur_item * 100) // corpus_count if corpus_count > 0 else 0
                print(f"  cycles {cycles_done + 1}, lifetime speed {exec_sec} execs/sec, "
                      f"items {cur_item}/{corpus_count} ({path_perc}%)")

                if saved_crashes == 0:
                    print(f"  pending {pending_favs}/{pending_total}, coverage {bitmap_cvg}, no crashes yet")
                else:
                    print(f"  pending {pending_favs}/{pending_total}, coverage {bitmap_cvg}, "
                          f"crashes saved {saved_crashes} (!)")

                print()
        else:
            # fuzzer_stats not found, but fuzzer_setup exists: likely starting up
            if setup_path.exists() and args.process_dead:
                start_cnt += 1
                # include in alive count? Original script increments ALIVE_CNT for starting if PROCESS_DEAD is set.
                # Actually they do: ALIVE_CNT=$((ALIVE_CNT+1)) and then later subtract? Let's mimic: they add to ALIVE_CNT
                # and also to START_CNT. Then in summary, if PROCESS_DEAD is set, they adjust ALIVE_CNT = ALIVE_CNT - DEAD_CNT - START_CNT.
                # So we just add to start_cnt and also to alive_cnt? We'll handle in summary.
                # For now, we just increment start_cnt and we'll adjust alive_cnt later.
                pass

    # ------------------------------------------------------------------
    # Summary stats
    # ------------------------------------------------------------------
    if not args.summary_only or not args.minimal_only:
        print("Summary stats")
        print("=============")
        if not args.summary_only and not args.minimal_only:
            print()

    # Adjust alive_cnt if we are counting dead/starting: original script subtracts dead and start from alive.
    if args.process_dead:
        alive_cnt = alive_cnt - dead_cnt - start_cnt

    print(f"        Fuzzers alive : {alive_cnt}")

    if start_cnt != 0:
        txt = "excluded from stats" if not args.process_dead else "included in stats"
        print(f"          Starting up : {start_cnt} ({txt})")
    if dead_cnt != 0:
        txt = "excluded from stats" if not args.process_dead else "included in stats"
        print(f"       Dead or remote : {dead_cnt} ({txt})")

    # Total run time
    total_run_seconds = current_time - total_time if total_time > 0 else 0
    print(f"       Total run time : {fmt_duration(total_run_seconds)}")

    if not args.minimal_only:
        # Total execs in millions/thousands
        execs_million = total_execs // 1000000
        execs_thousand = (total_execs % 1000000) // 1000
        if execs_million > 9:
            fmt_execs = f"{execs_million} millions"
        elif execs_million > 0:
            fmt_execs = f"{execs_million} millions, {execs_thousand} thousands"
        else:
            fmt_execs = f"{execs_thousand} thousands"
        print(f"          Total execs : {fmt_execs}")
        print(f"     Cumulative speed : {total_eps} execs/sec")
        if alive_cnt > 0:
            print(f"  Total average speed : {total_eps // alive_cnt if alive_cnt else 0} execs/sec")

    if alive_cnt > 0:
        print(f"Current average speed : {total_eplm // alive_cnt if alive_cnt else 0} execs/sec")

    if not args.minimal_only:
        print(f"        Pending items : {total_pfav} faves, {total_pending} total")

    if alive_cnt > 1 or args.minimal_only:
        if alive_cnt > 0:
            avg_pfav = total_pfav // alive_cnt if alive_cnt else 0
            avg_pending = total_pending // alive_cnt if alive_cnt else 0
            print(f"   Pending per fuzzer : {avg_pfav} faves, {avg_pending} total (on average)")

    print(f"     Coverage reached : {total_coverage if total_coverage is not None else '0'}%")
    print(f"        Crashes saved : {total_crashes}")
    if not args.minimal_only:
        print(f"          Hangs saved : {total_hangs}")
        if total_wcop:
            total_wcop_str = "/".join(str(w) for w in total_wcop)
        else:
            total_wcop_str = "not available"
        print(f" Cycles without finds : {total_wcop_str}")
    print(f"   Time without finds : {fmt_duration(total_last_find if total_last_find > 0 else 0)}")
    print()

    sys.exit(0)


if __name__ == "__main__":
    main()