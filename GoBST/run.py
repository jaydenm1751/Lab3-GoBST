#!/usr/bin/env python3
import os
import re
import csv
import time
import statistics as stats
from subprocess import run, PIPE, CalledProcessError

# -------- CONFIG --------
EXEC = 'go run ./cmd/gobst'      # or path to compiled binary, e.g. './bin/gobst'
INPUT_DIR = 'testdata'
INPUTS = ['coarse.txt', 'fine.txt']

HASH_WORKERS = [1, 2, 4, 8, 16]
TRIALS = 10
SLEEP_BETWEEN = 0.05             # seconds between trials
CSV_OUT = 'runs_agg.csv'

# Step 3 timing: set to 1 to get compareTreeTime, or 0 to skip comparisons
COMP_WORKERS = 1
# ------------------------

# Regexes for line-based timings
RE_HASH_TIME   = re.compile(r'^hashTime:\s*([0-9]+(?:\.[0-9]+)?)\s*$')
RE_GROUP_TIME  = re.compile(r'^hashGroupTime:\s*([0-9]+(?:\.[0-9]+)?)\s*$')
RE_COMPARE_TIME= re.compile(r'^compareTreeTime:\s*([0-9]+(?:\.[0-9]+)?)\s*$')

# Optional CSV summary line you print in Go:
# Header: Overall_Time,Total_Time,Build_Time,Hash_time,Compare_Time
RE_SUMMARY_LINE = re.compile(
    r'^\s*([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?),'
    r'([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?)\s*$'
)

def run_once(cmd: str) -> str:
    try:
        r = run(cmd, shell=True, stdout=PIPE, stderr=PIPE,
                universal_newlines=True, check=True)
        return r.stdout
    except CalledProcessError as e:
        print("\nERROR running:", cmd)
        print("--- STDOUT ---\n", e.stdout)
        print("--- STDERR ---\n", e.stderr)
        raise

def parse_metrics(stdout: str):
    """
    Returns a dict with any of:
      hashTime_s, hashGroupTime_s, compareTime_s,
      overall_s, total_s, build_s, hash_s, compare_s
    Only keys found will be present.
    """
    m = {}
    # First pass: named lines
    for line in stdout.splitlines():
        line = line.strip()
        m1 = RE_HASH_TIME.match(line)
        if m1:
            m['hashTime_s'] = float(m1.group(1))
            continue
        m2 = RE_GROUP_TIME.match(line)
        if m2:
            m['hashGroupTime_s'] = float(m2.group(1))
            continue
        m3 = RE_COMPARE_TIME.match(line)
        if m3:
            m['compareTime_s'] = float(m3.group(1))
            continue

    # Second pass: optional CSV summary line AFTER its header
    # Look for the last numeric CSV-like line with 5 fields
    # (Assumes you've printed the header once just before it)
    last_csv = None
    for line in stdout.splitlines():
        line = line.strip()
        msum = RE_SUMMARY_LINE.match(line)
        if msum:
            last_csv = msum
    if last_csv:
        m['overall_s']  = float(last_csv.group(1))
        m['total_s']    = float(last_csv.group(2))
        m['build_s']    = float(last_csv.group(3))
        m['hash_s']     = float(last_csv.group(4))
        m['compare_s']  = float(last_csv.group(5))
    return m

def build_cmd(input_path: str, hash_workers: int, data_workers: int, comp_workers: int) -> str:
    return (f'{EXEC} -input "{input_path}" '
            f'-hash-workers={hash_workers} -data-workers={data_workers} '
            f'-comp-workers={comp_workers}')

def aggregate(trials):
    """
    trials: list of dicts with float metrics.
    Returns a single dict with avg_ and min_ for each metric seen.
    """
    keys = set().union(*trials) if trials else set()
    out = {}
    for k in sorted(keys):
        vals = [t[k] for t in trials if k in t]
        if not vals:
            continue
        out[f'avg_{k}'] = stats.mean(vals)
        out[f'min_{k}'] = min(vals)
    return out

def main():
    rows = []
    for fname in INPUTS:
        ipath = os.path.join(INPUT_DIR, fname)

        for hw in HASH_WORKERS:
            # Mode A: seq/chan (data-workers=1)
            dw = 1
            mode = 'seq' if hw == 1 else 'chan'
            cmd = build_cmd(ipath, hw, dw, COMP_WORKERS)
            print(f'\n{fname} | {mode} | hash-workers={hw} data-workers={dw} comp-workers={COMP_WORKERS}\nCMD: {cmd}')
            trials = []
            for t in range(1, TRIALS + 1):
                out = run_once(cmd)
                met = parse_metrics(out)
                trials.append(met)
                ht  = met.get('hashTime_s')
                hgt = met.get('hashGroupTime_s')
                ct  = met.get('compareTime_s')
                print(f'  trial {t}/{TRIALS}: '
                      f'hash={ht:.9f}s  hashGroup={hgt:.9f}s'
                      + (f'  compare={ct:.9f}s' if ct is not None else ''))
                time.sleep(SLEEP_BETWEEN)
            agg = aggregate(trials)
            row = {
                'file': fname,
                'mode': mode,
                'hash_workers': hw,
                'data_workers': dw,
                'comp_workers': COMP_WORKERS,
                'trials': TRIALS,
            }
            row.update(agg)
            rows.append(row)

            # Mode B: mutex (only when hw > 1)
            if hw > 1:
                dw = hw
                mode = 'mutex'
                cmd = build_cmd(ipath, hw, dw, COMP_WORKERS)
                print(f'\n{fname} | {mode} | hash-workers={hw} data-workers={dw} comp-workers={COMP_WORKERS}\nCMD: {cmd}')
                trials = []
                for t in range(1, TRIALS + 1):
                    out = run_once(cmd)
                    met = parse_metrics(out)
                    trials.append(met)
                    ht  = met.get('hashTime_s')
                    hgt = met.get('hashGroupTime_s')
                    ct  = met.get('compareTime_s')
                    print(f'  trial {t}/{TRIALS}: '
                          f'hash={ht:.9f}s  hashGroup={hgt:.9f}s'
                          + (f'  compare={ct:.9f}s' if ct is not None else ''))
                    time.sleep(SLEEP_BETWEEN)
                agg = aggregate(trials)
                row = {
                    'file': fname,
                    'mode': mode,
                    'hash_workers': hw,
                    'data_workers': dw,
                    'comp_workers': COMP_WORKERS,
                    'trials': TRIALS,
                }
                row.update(agg)
                rows.append(row)

    # Union of all metric columns we saw
    all_keys = set()
    for r in rows:
        all_keys |= set(r.keys())

    # Order columns: metadata first, then metrics
    meta = ['file', 'mode', 'hash_workers', 'data_workers', 'comp_workers', 'trials']
    metric_cols = sorted(k for k in all_keys if k not in meta)
    fieldnames = meta + metric_cols

    with open(CSV_OUT, 'w', newline='') as f:
        w = csv.DictWriter(f, fieldnames=fieldnames)
        w.writeheader()
        w.writerows(rows)

    print(f'\nWrote {CSV_OUT} with {len(rows)} rows.')

if __name__ == '__main__':
    main()
