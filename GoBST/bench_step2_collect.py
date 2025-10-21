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

THREADS = [1, 2, 4, 8, 16, 32, 64, 128, 256]  # hash/data worker counts
TRIALS  = 5
SLEEP_BETWEEN = 0.03
CSV_OUT = 'step2_hashgroup_averages.csv'

def avg(xs):
    return sum(xs) / len(xs) if xs else float('nan')

# Arg builders for Step 2 strategies.
# Tweak these to match your binary’s flags/names.
def args_sequential(t):  # Baseline: single thread everywhere
    return ['-hash-workers=1', '-data-workers=1', '-comp-workers=1']

def args_channel(t):     # Hash workers use channel to single inserter
    return [f'-hash-workers={t}', '-data-workers=1', '-comp-workers=1']

def args_mutex(t):       # Hash/data workers write to shared map under lock
    return [f'-hash-workers={t}', f'-data-workers={t}', '-comp-workers=1']

STRATEGIES = [
    ('Sequential', args_sequential),   # recorded at threads=1
    ('Channel',    args_channel),      # recorded at threads in THREADS (>=2 typical)
    ('Mutex',      args_mutex),
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
    """Return dict with HashGroup_Time if found."""
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
    rows = []  # one row per (file, strategy, threads)
    for fname in INPUTS:
        ipath = os.path.join(INPUT_DIR, fname)

        for strat_name, arg_fn in STRATEGIES:
            for t in THREADS:
                # Only record Sequential at t=1 (baseline)
                if strat_name == 'Sequential' and t != 1:
                    continue
                # For non-sequential at t=1, you can keep or skip; keeping is fine.
                args = arg_fn(t)
                cmd = build_cmd(ipath, args)

                samples = []
                for _ in range(TRIALS):
                    out = run_once(cmd)
                    met = parse_last_summary(out)
                    if 'HashGroup_Time' in met:
                        samples.append(met['HashGroup_Time'])
                    time.sleep(SLEEP_BETWEEN)

                if not samples:
                    continue

                rows.append({
                    'file': fname,
                    'strategy': strat_name,
                    'threads': t,
                    'trials': TRIALS,
                    'avg_HashGroup_Time_s': avg(samples),
                })
                print(f"{fname:10s} | {strat_name:10s} | t={t:3d} | avg HashGroup_Time = {avg(samples):.6f}s")

    # Write a compact CSV (both inputs together)
    with open(CSV_OUT, 'w', newline='') as f:
        w = csv.DictWriter(f, fieldnames=['file','strategy','threads','trials','avg_HashGroup_Time_s'])
        w.writeheader()
        w.writerows(rows)

    print(f"\n[OK] Wrote {CSV_OUT} with {len(rows)} rows.")

if __name__ == '__main__':
    main()
