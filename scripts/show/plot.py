# Copyright (c) 2025 pehmc. MIT License.
# See LICENSE file in the project root for full license information.

import argparse
import os
import sys
import subprocess
import shutil
import time
import tempfile
from pathlib import Path
from datetime import datetime


def get_abs_path(path: str) -> str:
    """Resolve absolute path, similar to get_abs_path() in the shell script."""
    return str(Path(path).resolve())


def check_gnuplot() -> str:
    """Check if gnuplot is available in PATH. Return path or raise FileNotFoundError."""
    gnuplot = shutil.which("gnuplot")
    if not gnuplot:
        raise FileNotFoundError("can't find 'gnuplot' in your $PATH")
    return gnuplot


def check_afl_plot_ui() -> str:
    """Check if afl-plot-ui is available and executable. Return path or raise FileNotFoundError."""
    ui = shutil.which("afl-plot-ui")
    if not ui:
        raise FileNotFoundError("afl-plot-ui not found in PATH")
    # Test by running -h
    result = subprocess.run([ui, "-h"], capture_output=True)
    if result.returncode != 0:
        raise RuntimeError("afl-plot-ui is not functioning correctly")
    return ui


def read_banner(stats_path: Path) -> str:
    """Extract afl_banner from fuzzer_stats file."""
    banner = "(none)"
    try:
        with open(stats_path) as f:
            for line in f:
                if line.startswith("afl_banner "):
                    parts = line.split(":", 1)
                    if len(parts) == 2:
                        banner = parts[1].strip()
                        break
    except (IOError, OSError):
        pass
    return banner


def generate_normal_plots(gnuplot: str, input_dir: Path, output_dir: Path):
    """Generate PNG plots using a single gnuplot session."""
    # Build the full gnuplot script
    script = f"""
{PLOT_HF}
{PLOT_LF}
{PLOT_ES}
{PLOT_EG}
"""
    # Substitute paths and options
    script = script.replace("$outputdir", str(output_dir))
    script = script.replace("$inputdir", str(input_dir))

    proc = subprocess.Popen(
        [gnuplot],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    _, stderr = proc.communicate(script)
    if proc.returncode != 0:
        print("Note: if you see errors concerning 'unknown or ambiguous terminal type'",
              "then you need to use a gnuplot that has png support compiled in.",
              file=sys.stderr)
        if stderr:
            print("gnuplot stderr:", stderr, file=sys.stderr)


def generate_graphical_plots(gnuplot: str, ui: str, input_dir: Path, output_dir: Path):
    """Generate plots in graphical windows using afl-plot-ui."""
    # Create temporary directory and FIFO for window IDs
    tmp_dir = output_dir / ".tmp"
    tmp_dir.mkdir(exist_ok=True)
    fifo_path = tmp_dir / "win_ids"

    try:
        os.mkfifo(fifo_path)
    except OSError as e:
        print(f"[-] Failed to create FIFO: {e}", file=sys.stderr)
        sys.exit(1)

    # Launch afl-plot-ui, redirect its output to the FIFO
    with open(fifo_path, "r") as fifo:
        ui_proc = subprocess.Popen(
            [ui],
            stdout=fifo,
            stderr=subprocess.PIPE,
            text=True
        )
        # Read the window IDs (four lines)
        win_ids = [fifo.readline().strip() for _ in range(4)]

    # Clean up
    shutil.rmtree(tmp_dir)

    if len(win_ids) < 4 or not all(win_ids):
        print("[-] Failed to obtain window IDs from afl-plot-ui", file=sys.stderr)
        ui_proc.terminate()
        sys.exit(1)

    w_id1, w_id2, w_id3, w_id4 = win_ids

    print("[*] Generating plots...")

    # Function to launch gnuplot with a specific plot script and window ID
    def launch_plot(plot_text, window_id):
        script = f"""
{plot_text}
set term x11 window "{window_id}"
set output
replot
pause mouse close
"""
        script = script.replace("$outputdir", str(output_dir))
        script = script.replace("$inputdir", str(input_dir))
        subprocess.Popen(
            [gnuplot],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True
        ).communicate(script)

    # Launch four separate processes
    launch_plot(PLOT_HF, w_id3)   # high_freq
    launch_plot(PLOT_LF, w_id4)   # low_freq
    launch_plot(PLOT_ES, w_id2)   # exec_speed
    launch_plot(PLOT_EG, w_id1)   # edges

    # Give them a moment to open windows
    time.sleep(1)


def create_index_html(output_dir: Path, input_dir: Path, banner: str):
    """Generate index.html in the output directory."""
    now = datetime.now().strftime("%a %b %d %H:%M:%S %Y")
    content = f"""<table style="font-family: 'Trebuchet MS', 'Tahoma', 'Arial', 'Helvetica'">
<tr><td style="width: 18ex"><b>Banner:</b></td><td>{banner}</td></tr>
<tr><td><b>Directory:</b></td><td>{input_dir}</td></tr>
<tr><td><b>Generated on:</b></td><td>{now}</td></tr>
</table>
<p>
<img src="edges.png" width=1000 height=300>
<img src="high_freq.png" width=1000 height=300><p>
<img src="low_freq.png" width=1000 height=200><p>
<img src="exec_speed.png" width=1000 height=200>
"""
    index_path = output_dir / "index.html"
    index_path.write_text(content)


def set_permissions(output_dir: Path):
    """Set permissions: 755 on dir, 644 on generated files."""
    # Directory
    output_dir.chmod(0o755)
    # PNG files
    for f in ["high_freq.png", "low_freq.png", "exec_speed.png", "edges.png"]:
        p = output_dir / f
        if p.exists():
            p.chmod(0o644)
    # index.html
    idx = output_dir / "index.html"
    if idx.exists():
        idx.chmod(0o644)


# Gnuplot scripts (as in the original, but with placeholders for $inputdir, $outputdir)
GNUPLOT_SETUP = """
#set xdata time
#set timefmt '%s'
#set format x \"%b %d\\n%H:%M\"
set tics font 'small'
unset mxtics
unset mytics

set grid xtics linetype 0 linecolor rgb '#e0e0e0'
set grid ytics linetype 0 linecolor rgb '#e0e0e0'
set border linecolor rgb '#50c0f0'
set tics textcolor rgb '#000000'
set key outside

set autoscale xfixmin
set autoscale xfixmax

set xlabel \"relative time in seconds\" font \"small\"
"""

PLOT_HF = """
set terminal png truecolor enhanced size 1000,300 butt
set output '$outputdir/high_freq.png'

$GNUPLOT_SETUP

plot '$inputdir/plot_data' using 1:4 with filledcurve x1 title 'corpus count' linecolor rgb '#000000' fillstyle transparent solid 0.2 noborder, \\
     '' using 1:3 with filledcurve x1 title 'current item' linecolor rgb '#f0f0f0' fillstyle transparent solid 0.5 noborder, \\
     '' using 1:5 with lines title 'pending items' linecolor rgb '#0090ff' linewidth 3, \\
     '' using 1:6 with lines title 'pending favs' linecolor rgb '#c00080' linewidth 3, \\
     '' using 1:2 with lines title 'cycles done' linecolor rgb '#c000f0' linewidth 3
"""

PLOT_LF = """
set terminal png truecolor enhanced size 1000,200 butt
set output '$outputdir/low_freq.png'

$GNUPLOT_SETUP

plot '$inputdir/plot_data' using 1:8 with filledcurve x1 title '' linecolor rgb '#c00080' fillstyle transparent solid 0.2 noborder, \\
     '' using 1:8 with lines title ' uniq crashes' linecolor rgb '#c00080' linewidth 3, \\
     '' using 1:9 with lines title 'uniq hangs' linecolor rgb '#c000f0' linewidth 3, \\
     '' using 1:10 with lines title 'levels' linecolor rgb '#0090ff' linewidth 3
"""

PLOT_ES = """
set terminal png truecolor enhanced size 1000,200 butt
set output '$outputdir/exec_speed.png'

$GNUPLOT_SETUP

plot '$inputdir/plot_data' using 1:11 with filledcurve x1 title '' linecolor rgb '#0090ff' fillstyle transparent solid 0.2 noborder, \\
     '$inputdir/plot_data' using 1:11 with lines title '    execs/sec' linecolor rgb '#0090ff' linewidth 3 smooth bezier;
"""

PLOT_EG = """
set terminal png truecolor enhanced size 1000,300 butt
set output '$outputdir/edges.png'

$GNUPLOT_SETUP

plot '$inputdir/plot_data' using 1:13 with lines title '        edges' linecolor rgb '#0090ff' linewidth 3
"""

# Insert setup into each plot script
PLOT_HF = PLOT_HF.replace("$GNUPLOT_SETUP", GNUPLOT_SETUP)
PLOT_LF = PLOT_LF.replace("$GNUPLOT_SETUP", GNUPLOT_SETUP)
PLOT_ES = PLOT_ES.replace("$GNUPLOT_SETUP", GNUPLOT_SETUP)
PLOT_EG = PLOT_EG.replace("$GNUPLOT_SETUP", GNUPLOT_SETUP)


def main():
    parser = argparse.ArgumentParser(
        description="Generate gnuplot images from afl-fuzz output data."
    )
    parser.add_argument(
        "-g", "--graphical",
        action="store_true",
        help="display the plots in a graphical window (requires afl-plot-ui)"
    )
    parser.add_argument(
        "afl_state_dir",
        help="existing state directory for an afl-fuzz instance"
    )
    parser.add_argument(
        "graph_output_dir",
        help="empty directory where plots will be written"
    )
    args = parser.parse_args()

    print("progress plotting utility for afl-fuzz by Michal Zalewski")
    print()

    # Resolve absolute paths
    try:
        input_dir = Path(get_abs_path(args.afl_state_dir))
        output_dir = Path(get_abs_path(args.graph_output_dir))
    except Exception as e:
        print(f"[-] Error resolving paths: {e}", file=sys.stderr)
        sys.exit(1)

    # Check for plot_data
    plot_data = input_dir / "plot_data"
    if not plot_data.exists():
        # Special case: maybe they meant input_dir/default/plot_data?
        alt = input_dir / "default" / "plot_data"
        if alt.exists():
            print(f"[-] Error: input directory is not valid (missing 'plot_data'), "
                  f"likely you mean {input_dir}/default?", file=sys.stderr)
        else:
            print("[-] Error: input directory is not valid (missing 'plot_data').", file=sys.stderr)
        sys.exit(1)

    # Check plot_data has at least 3 lines
    with open(plot_data) as f:
        line_count = sum(1 for _ in f)
    if line_count < 3:
        print("[-] Error: plot_data carries too little data, let it run longer.", file=sys.stderr)
        sys.exit(1)

    # Read banner
    stats_path = input_dir / "fuzzer_stats"
    banner = read_banner(stats_path)

    # Check for gnuplot
    try:
        gnuplot = check_gnuplot()
    except FileNotFoundError as e:
        print(f"[-] Error: {e}", file=sys.stderr)
        sys.exit(1)

    # Create output directory
    try:
        output_dir.mkdir(exist_ok=True)
    except OSError as e:
        print(f"[-] Error: unable to create output directory: {e}", file=sys.stderr)
        sys.exit(1)

    # Clean previous files in output dir
    for f in output_dir.glob("high_freq.png"):
        f.unlink()
    for f in output_dir.glob("low_freq.png"):
        f.unlink()
    for f in output_dir.glob("exec_speed.png"):
        f.unlink()
    for f in output_dir.glob("edges.png"):
        f.unlink()
    # Backup index.html
    index_orig = output_dir / "index.html"
    if index_orig.exists():
        index_orig.rename(output_dir / "index.html.orig")

    # Generate plots
    if args.graphical:
        try:
            ui = check_afl_plot_ui()
        except (FileNotFoundError, RuntimeError) as e:
            print(
                "[-] You do not seem to have the afl-plot-ui utility installed or it is not functional.\n"
                "    If you have installed afl-plot-ui, make sure the executable is in your PATH.\n"
                "    If you are still facing problems, please open an issue at\n"
                "    https://github.com/AFLplusplus/AFLplusplus/issues.\n\n"
                "    No plots have been generated. Please rerun without the -g flag to generate PNGs.",
                file=sys.stderr
            )
            sys.exit(1)
        generate_graphical_plots(gnuplot, ui, input_dir, output_dir)
    else:
        print("[*] Generating plots...")
        generate_normal_plots(gnuplot, input_dir, output_dir)
        print("[?] You can also use -g flag to view the plots in a GUI window, and interact with the plots (if you have built afl-plot-ui). Run \"afl-plot -h\" to know more.")

    # Verify that exec_speed.png was created and non-empty
    exec_speed = output_dir / "exec_speed.png"
    if not exec_speed.exists() or exec_speed.stat().st_size == 0:
        print("[-] Error: something went wrong! Perhaps you have an ancient version of gnuplot?", file=sys.stderr)
        sys.exit(1)

    # Generate index.html
    print("[*] Generating index.html...")
    create_index_html(output_dir, input_dir, banner)

    # Set permissions
    set_permissions(output_dir)

    print("[+] All done - enjoy your charts!")
    sys.exit(0)


if __name__ == "__main__":
    main()