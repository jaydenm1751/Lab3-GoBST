#!/usr/bin/env python3
import os
import re
import csv
import time
import statistics as stats
from subprocess import run, PIPE, CalledProcessError
from typing import List

# ==================== CONFIG (edit if needed) ====================
EXEC = 'go run ./cmd/gobst'          # or './gobst'
INPUT_DIR = 'testdata'
INPUTS = ['coarse.txt', 'fine.txt']

THREADS = [1, 2, 4, 8, 16, 32, 64, 128, 256]  # comp-worker counts
TRIALS  = 5
SLEEP_BETWEEN = 0.03
CSV_OUT = 'step3_compare_averages.csv'

# Keep hashing fixed so we isolate comparison time differences.
FIXED_HASH_ARGS = ['-hash-workers=1', '-data-workers=1']

def avg(xs):
    return sum(xs) / len(xs) if xs else float('nan')

# Comparison strategies — adjust flags to match your binary.
# Example: per-comparison (unbounded) vs thread-pool (fixed comp-workers).
def args_percomparison(t: int) -> List[str]:
    # t ignored; use a flag your binary understands for per-comparison mode.
    return FIXED_HASH_ARGS + ['-comp-workers=0']

def args_threadpool(t: int) -> List[str]:
    return FIXED_HASH_ARGS + [f'-comp-workers={t}']

STRATEGIES = [
    ('PerComparison', args_percomparison),  # recorded at threads=1 (t ignored)
    ('ThreadPool',    args_threadpool),     # recorded at threads in THREADS
]
# ================================================================

# CSV summary line parsing (from your Go print)
HDR_RE = re.compile(r'^\s*Overall_Time,Total_Time,Build_Time,Hash_time,HashGroup_Time,Compare_Time\s*$')
NUM_RE = re.compile(r'^\s*([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?)\s*$')

def run_once(cmd: str) -> str:
    try:
        r = run(cmd, shell=True, stdout=PIPE, stderr=PIPE, universal_newlines=True, check=True)
        return r.stdout
    except CalledProcessError as e:
        print("\nERROR running:", cmd)
        print("--- STDOUT ---\n", e.stdout)
        print("--- STDERR ---\n", e.stderr)
        raise

def parse_last_summary(stdout: str):
    """Return dict with Compare_Time if found."""
    last = None
    saw_hdr = False
    for line in stdout.splitlines():
        if HDR_RE.match(line):
            saw_hdr = True
            continue
        if saw_hdr:
            m = NUM_RE.match(line.strip())
            if m:
                last = m
            saw_hdr = False
    if not last:
        return {}
    return {
        'Overall_Time':   float(last.group(1)),
        'Total_Time':     float(last.group(2)),
        'Build_Time':     float(last.group(3)),
        'Hash_time':      float(last.group(4)),
        'HashGroup_Time': float(last.group(5)),
        'Compare_Time':   float(last.group(6)),
    }

def build_cmd(input_path: str, args_list: List[str]) -> str:
    return ' '.join([EXEC, '-input', f'"{input_path}"'] + args_list)

def main():
    rows = []
    for fname in INPUTS:
        ipath = os.path.join(INPUT_DIR, fname)

        for strat_name, arg_fn in STRATEGIES:
            for t in THREADS:
                # Only record PerComparison once at "threads=1"
                if strat_name == 'PerComparison' and t != 1:
                    continue

                args = arg_fn(t)
                cmd = build_cmd(ipath, args)

                samples = []
                for _ in range(TRIALS):
                    out = run_once(cmd)
                    met = parse_last_summary(out)
                    if 'Compare_Time' in met:
                        samples.append(met['Compare_Time'])
                    time.sleep(SLEEP_BETWEEN)

                if not samples:
                    continue

                rows.append({
                    'file': fname,
                    'strategy': strat_name,
                    'threads': t,
                    'trials': TRIALS,
                    'avg_Compare_Time_s': avg(samples),
                })
                print(f"{fname:10s} | {strat_name:14s} | t={t:3d} | avg Compare_Time = {avg(samples):.6f}s")

    with open(CSV_OUT, 'w', newline='') as f:
        w = csv.DictWriter(f, fieldnames=['file','strategy','threads','trials','avg_Compare_Time_s'])
        w.writeheader()
        w.writerows(rows)

    print(f"\n[OK] Wrote {CSV_OUT} with {len(rows)} rows.")

if __name__ == '__main__':
    main()
