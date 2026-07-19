import React, { useState } from "react";
import type { Signal, Severity } from "../types";
import { SignalCard } from "./SignalCard";

interface Props {
  signals: Signal[];
  selectedFixtureId: number | null;
}

const SEVERITY_OPTIONS: (Severity | "ALL")[] = ["ALL", "HIGH", "MEDIUM", "LOW"];

export function SignalFeed({ signals, selectedFixtureId }: Props) {
  const [severityFilter, setSeverityFilter] = useState<Severity | "ALL">("ALL");
  const [directionFilter, setDirectionFilter] = useState<"ALL" | "SHORTENING" | "DRIFTING">("ALL");

  const filtered = signals
    .filter((s) => !selectedFixtureId || s.fixture_id === selectedFixtureId)
    .filter((s) => severityFilter === "ALL" || s.severity === severityFilter)
    .filter((s) => directionFilter === "ALL" || s.direction === directionFilter)
    .sort((a, b) => new Date(b.detected_at).getTime() - new Date(a.detected_at).getTime());

  const highCount = signals.filter((s) => s.severity === "HIGH").length;
  const medCount  = signals.filter((s) => s.severity === "MEDIUM").length;

  return (
    <main style={styles.main}>
      {/* Feed header */}
      <div style={styles.feedHeader}>
        <div style={styles.feedTitle}>
          <span style={styles.feedLabel}>SIGNAL FEED</span>
          <span style={styles.feedCount}>{filtered.length} signals</span>
          {highCount > 0 && (
            <span style={styles.highAlert}>
              ⚡ {highCount} HIGH
            </span>
          )}
        </div>

        <div style={styles.filters}>
          <div style={styles.filterGroup}>
            <span style={styles.filterLabel}>SEVERITY</span>
            <div style={styles.filterButtons}>
              {SEVERITY_OPTIONS.map((s) => (
                <button
                  key={s}
                  onClick={() => setSeverityFilter(s)}
                  style={{
                    ...styles.filterBtn,
                    ...(severityFilter === s ? styles.filterBtnActive : {}),
                  }}
                >
                  {s}
                </button>
              ))}
            </div>
          </div>

          <div style={styles.filterGroup}>
            <span style={styles.filterLabel}>DIRECTION</span>
            <div style={styles.filterButtons}>
              {(["ALL", "SHORTENING", "DRIFTING"] as const).map((d) => (
                <button
                  key={d}
                  onClick={() => setDirectionFilter(d)}
                  style={{
                    ...styles.filterBtn,
                    ...(directionFilter === d ? styles.filterBtnActive : {}),
                  }}
                >
                  {d === "ALL" ? "ALL" : d === "SHORTENING" ? "▲ SHORT" : "▼ DRIFT"}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Stats bar */}
      <div style={styles.statsBar}>
        <StatChip label="TOTAL" value={signals.length} />
        <StatChip label="HIGH" value={highCount} color="var(--danger)" />
        <StatChip label="MEDIUM" value={medCount} color="var(--warn)" />
        <StatChip
          label="RESOLVED"
          value={signals.filter((s) => s.resolved).length}
          color="var(--accent)"
        />
        <StatChip
          label="CORRECT"
          value={signals.filter((s) => s.prediction_correct).length}
          color="var(--accent)"
        />
        <div style={styles.statsDivider} />
        <div style={styles.accuracyChip}>
          <span style={styles.accuracyLabel}>ACCURACY</span>
          <span style={styles.accuracyValue}>
            {signals.filter((s) => s.resolved).length > 0
              ? `${Math.round(
                  (signals.filter((s) => s.prediction_correct).length /
                    signals.filter((s) => s.resolved).length) *
                    100
                )}%`
              : "—"}
          </span>
        </div>
      </div>

      {/* Signal list */}
      <div style={styles.list}>
        {filtered.length === 0 ? (
          <EmptyState />
        ) : (
          filtered.map((sig, i) => (
            <SignalCard
              key={`${sig.fixture_id}-${sig.outcome_id}-${sig.detected_at}`}
              signal={sig}
              index={i}
            />
          ))
        )}
      </div>
    </main>
  );
}

function StatChip({
  label,
  value,
  color,
}: {
  label: string;
  value: number;
  color?: string;
}) {
  return (
    <div style={styles.statChip}>
      <span style={styles.statChipLabel}>{label}</span>
      <span style={{ ...styles.statChipValue, color: color ?? "var(--text)" }}>
        {value}
      </span>
    </div>
  );
}

function EmptyState() {
  return (
    <div style={styles.empty}>
      <div style={styles.emptyIcon}>◎</div>
      <div style={styles.emptyTitle}>Monitoring markets</div>
      <div style={styles.emptyDesc}>
        Signals appear here when odds shift by more than 4% between polls.
        France vs England kicks off at 21:00 UTC.
      </div>
      <div style={styles.emptyTip}>
        <span style={styles.tipLabel}>TIP</span>
        The most significant movements happen in the 2 hours before kickoff
        and during in-play score changes.
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  main: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
    minWidth: 0,
  },
  feedHeader: {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    padding: "10px 16px",
    borderBottom: "1px solid var(--border)",
    flexShrink: 0,
    flexWrap: "wrap",
    gap: 8,
  },
  feedTitle: {
    display: "flex",
    alignItems: "center",
    gap: 10,
  },
  feedLabel: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 600,
    color: "var(--text-3)",
    letterSpacing: "0.12em",
  },
  feedCount: {
    fontSize: 12,
    color: "var(--text-2)",
    background: "var(--surface-2)",
    padding: "1px 7px",
    borderRadius: 3,
    fontFamily: "var(--mono)",
  },
  highAlert: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 700,
    color: "var(--danger)",
    background: "var(--danger-dim)",
    padding: "2px 8px",
    borderRadius: 3,
    border: "1px solid #EF444444",
    animation: "glow-pulse 2s ease-in-out infinite",
  },
  filters: {
    display: "flex",
    gap: 16,
    alignItems: "center",
  },
  filterGroup: {
    display: "flex",
    alignItems: "center",
    gap: 6,
  },
  filterLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
    letterSpacing: "0.1em",
  },
  filterButtons: {
    display: "flex",
    gap: 2,
  },
  filterBtn: {
    padding: "3px 8px",
    background: "var(--surface-2)",
    border: "1px solid var(--border)",
    borderRadius: 3,
    color: "var(--text-3)",
    fontFamily: "var(--mono)",
    fontSize: 10,
    cursor: "pointer",
    transition: "all 0.1s",
    letterSpacing: "0.04em",
  },
  filterBtnActive: {
    background: "var(--accent-dim)",
    borderColor: "var(--accent-glow)",
    color: "var(--accent)",
  },
  statsBar: {
    display: "flex",
    alignItems: "center",
    gap: 16,
    padding: "8px 16px",
    borderBottom: "1px solid var(--border)",
    background: "var(--surface)",
    flexShrink: 0,
  },
  statChip: {
    display: "flex",
    alignItems: "center",
    gap: 6,
  },
  statChipLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
    letterSpacing: "0.1em",
  },
  statChipValue: {
    fontFamily: "var(--mono)",
    fontSize: 14,
    fontWeight: 600,
  },
  statsDivider: {
    width: 1,
    height: 20,
    background: "var(--border-2)",
    marginLeft: "auto",
  },
  accuracyChip: {
    display: "flex",
    alignItems: "center",
    gap: 6,
    padding: "4px 10px",
    background: "var(--accent-dim)",
    border: "1px solid var(--accent-glow)",
    borderRadius: 4,
  },
  accuracyLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--accent)",
    letterSpacing: "0.1em",
  },
  accuracyValue: {
    fontFamily: "var(--mono)",
    fontSize: 15,
    fontWeight: 700,
    color: "var(--accent)",
  },
  list: {
    flex: 1,
    overflow: "auto",
    display: "flex",
    flexDirection: "column",
    gap: 8,
    padding: "12px 16px",
  },
  empty: {
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
    padding: "60px 40px",
    textAlign: "center",
  },
  emptyIcon: {
    fontSize: 36,
    color: "var(--text-3)",
    animation: "pulse-dot 2s ease-in-out infinite",
  },
  emptyTitle: {
    fontSize: 16,
    fontWeight: 500,
    color: "var(--text-2)",
  },
  emptyDesc: {
    fontSize: 13,
    color: "var(--text-3)",
    lineHeight: 1.6,
    maxWidth: 360,
  },
  emptyTip: {
    display: "flex",
    alignItems: "flex-start",
    gap: 8,
    padding: "10px 14px",
    background: "var(--surface-2)",
    border: "1px solid var(--border)",
    borderRadius: 6,
    maxWidth: 360,
    textAlign: "left",
  },
  tipLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    fontWeight: 700,
    color: "var(--accent)",
    letterSpacing: "0.1em",
    flexShrink: 0,
    paddingTop: 2,
  },
};
