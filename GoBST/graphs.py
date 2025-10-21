#!/usr/bin/env python3
import csv
import os
import argparse
import math
import matplotlib.pyplot as plt
from collections import defaultdict

# -----------------------
# Helpers
# -----------------------
def log(msg): print(msg)

def read_csv(path):
    rows = []
    with open(path, newline='') as f:
        r = csv.DictReader(f)
        for row in r:
            rows.append(row)
    return rows

def to_int(x, default=None):
    try: return int(x)
    except: return default

def to_float(x, default=None):
    try: return float(x)
    except: return default

def ensure_log2_xticks(ax, xs):
    # nice base-2 ticks if xs are powers of two
    pows = sorted(set(xs))
    ax.set_xscale('log', base=2)
    ax.set_xticks(pows)
    ax.get_xaxis().set_major_formatter(plt.ScalarFormatter())
    ax.get_xaxis().set_minor_formatter(plt.NullFormatter())

# -----------------------
# STEP 2: HashGroup_Time speedups
# -----------------------
def plot_step2(step2_csv, outdir):
    if not os.path.exists(step2_csv):
        log(f"[skip] {step2_csv} not found.")
        return []

    rows = read_csv(step2_csv)
    # Normalize types
    data = []
    for r in rows:
        data.append({
            'file': r['file'],
            'strategy': r['strategy'],
            'threads': to_int(r['threads']),
            'avg': to_float(r.get('avg_HashGroup_Time_s')),
        })

    # Group per file
    files = sorted(set(r['file'] for r in data))
    made = []
    for fname in files:
        sub = [r for r in data if r['file'] == fname and r['avg'] is not None and r['threads'] is not None]

        # Baseline: Sequential @ threads=1
        base_rows = [r for r in sub if r['strategy'].lower() == 'sequential' and r['threads'] == 1]
        if not base_rows:
            log(f"[warn:step2] Missing baseline Sequential@1 for {fname}, skipping.")
            continue
        baseline = base_rows[0]['avg']
        if not baseline or baseline <= 0:
            log(f"[warn:step2] Bad baseline for {fname}, skipping.")
            continue

        # Build curves: one per strategy (include Sequential flat line)
        by_strategy = defaultdict(list)
        for r in sub:
            by_strategy[r['strategy']].append((r['threads'], r['avg']))

        plt.figure()
        # Sequential flat line across all seen thread counts
        all_threads = sorted(set(t for (_, lst) in by_strategy.items() for (t, _) in lst))
        plt.plot(all_threads, [1.0]*len(all_threads), marker='o', label='Sequential')

        for strat, pts in sorted(by_strategy.items()):
            if strat.lower() == 'sequential':
                continue  # already plotted as flat=1
            pts = sorted((t, v) for (t, v) in pts if t and v and v > 0)
            if not pts: continue
            xs = [t for (t, _) in pts]
            ys = [baseline/v for (_, v) in pts]
            plt.plot(xs, ys, marker='o', label=strat)

        ax = plt.gca()
        ensure_log2_xticks(ax, all_threads)
        plt.xlabel('Number of Hash/Data Workers')
        plt.ylabel('Speedup (vs Sequential)')
        title = f"Step 2: Hash/Group Speedup — {fname}"
        plt.title(title)
        plt.grid(True, which='both', linestyle='--', alpha=0.5)
        plt.legend()
        os.makedirs(outdir, exist_ok=True)
        outpath = os.path.join(outdir, f"step2_speedup_{os.path.splitext(fname)[0]}.png")
        plt.tight_layout(); plt.savefig(outpath, dpi=200)
        plt.close()
        log(f"[ok] wrote {outpath}")
        made.append(outpath)
    return made

# -----------------------
# STEP 3: Compare_Time speedups
# -----------------------
def choose_step3_baseline(rows_for_file):
    """
    Prefer PerComparison@1 if present; otherwise any strategy with the minimum threads==1.
    Returns (baseline_time, baseline_label, baseline_threads) or (None, None, None).
    """
    per_1 = [r for r in rows_for_file if r['strategy'].lower() == 'percomparison' and r['threads'] == 1]
    if per_1 and per_1[0]['avg'] and per_1[0]['avg'] > 0:
        return per_1[0]['avg'], 'PerComparison', 1
    # fallback: any strategy at threads=1
    cands = [r for r in rows_for_file if r['threads'] == 1 and r['avg'] and r['avg'] > 0]
    if cands:
        cands.sort(key=lambda r: r['avg'])
        r0 = cands[0]
        return r0['avg'], r0['strategy'], 1
    return None, None, None

def plot_step3(step3_csv, outdir):
    if not os.path.exists(step3_csv):
        log(f"[skip] {step3_csv} not found.")
        return []

    rows = read_csv(step3_csv)
    data = []
    for r in rows:
        data.append({
            'file': r['file'],
            'strategy': r['strategy'],
            'threads': to_int(r['threads']),
            'avg': to_float(r.get('avg_Compare_Time_s')),
        })

    files = sorted(set(r['file'] for r in data))
    made = []
    for fname in files:
        sub = [r for r in data if r['file'] == fname and r['avg'] is not None and r['threads'] is not None]
        if not sub:
            continue

        baseline, blabel, bthreads = choose_step3_baseline(sub)
        if not baseline:
            log(f"[warn:step3] No usable baseline for {fname}, skipping.")
            continue

        # Organize by strategy
        by_strategy = defaultdict(list)
        for r in sub:
            by_strategy[r['strategy']].append((r['threads'], r['avg']))

        plt.figure()
        # Sequential flat line across all seen thread counts
        all_threads = sorted(set(t for (_, lst) in by_strategy.items() for (t, _) in lst))
        plt.plot(all_threads, [1.0]*len(all_threads), marker='o', label='Sequential')

        # Plot a point/line for each strategy
        for strat, pts in sorted(by_strategy.items()):
            pts = sorted((t, v) for (t, v) in pts if t and v and v > 0)
            if not pts: continue
            xs = [t for (t, _) in pts]
            ys = [baseline/v for (_, v) in pts]
            # If a strategy only has a single t=1 point (PerComparison), it will show as a point at speedup=1
            plt.plot(xs, ys, marker='o', label=strat)

        ax = plt.gca()
        ensure_log2_xticks(ax, all_threads)
        plt.xlabel('Number of Comparison Workers')
        plt.ylabel('Speedup (vs Sequential)')
        plt.title(f"Step 3: Comparison Speedup — {fname}")
        plt.grid(True, which='both', linestyle='--', alpha=0.5)
        plt.legend()
        os.makedirs(outdir, exist_ok=True)
        outpath = os.path.join(outdir, f"step3_speedup_{os.path.splitext(fname)[0]}.png")
        plt.tight_layout(); plt.savefig(outpath, dpi=200)
        plt.close()
        log(f"[ok] wrote {outpath}")
        made.append(outpath)
    return made

# -----------------------
# Main
# -----------------------
def main():
    ap = argparse.ArgumentParser(description="Plot GoBST speedup graphs from CSVs.")
    ap.add_argument('--step2', default='step2_hashgroup_averages.csv', help='CSV with Step 2 averages')
    ap.add_argument('--step3', default='step3_compare_averages.csv', help='CSV with Step 3 averages')
    ap.add_argument('--outdir', default='.', help='Output directory for PNGs')
    args = ap.parse_args()

    made = []
    made += plot_step2(args.step2, args.outdir)
    made += plot_step3(args.step3, args.outdir)

    if not made:
        log("[warn] No graphs produced. Check CSV paths and contents.")
    else:
        log(f"[done] Generated {len(made)} graphs.")

if __name__ == '__main__':
    main()
