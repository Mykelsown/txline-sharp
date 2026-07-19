import React, { useState } from "react";
import type { Signal } from "../types";
import { ProbabilityBar } from "./ProbabilityBar";

interface Props {
  signal: Signal;
  index: number;
}

const SEVERITY_COLORS: Record<string, string> = {
  HIGH: "var(--danger)",
  MEDIUM: "var(--warn)",
  LOW: "var(--text-3)",
};

const SEVERITY_BG: Record<string, string> = {
  HIGH: "var(--danger-dim)",
  MEDIUM: "var(--warn-dim)",
  LOW: "var(--surface-2)",
};

export function SignalCard({ signal, index }: Props) {
  const [expanded, setExpanded] = useState(index === 0);
  const severityColor = SEVERITY_COLORS[signal.severity];
  const severityBg = SEVERITY_BG[signal.severity];
  const time = new Date(signal.detected_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });

  return (
    <div
      style={{
        ...styles.card,
        animation: "slide-in 0.25s ease-out",
        borderLeft: `3px solid ${severityColor}`,
      }}
    >
      {/* Header row */}
      <div style={styles.header} onClick={() => setExpanded((e) => !e)}>
        <div style={styles.headerLeft}>
          <span style={{
            ...styles.severityTag,
            color: severityColor,
            background: severityBg,
            border: `1px solid ${severityColor}44`,
          }}>
            {signal.severity}
          </span>
          <span style={styles.direction}>
            {signal.direction === "SHORTENING" ? "▲" : "▼"} {signal.direction}
          </span>
          <span style={styles.market}>{signal.market_name}</span>
        </div>

        <div style={styles.headerRight}>
          <span style={styles.time}>{time}</span>
          {signal.resolved && (
            <span style={{
              ...styles.resolvedTag,
              color: signal.prediction_correct ? "var(--accent)" : "var(--danger)",
              background: signal.prediction_correct ? "var(--accent-dim)" : "var(--danger-dim)",
            }}>
              {signal.prediction_correct ? "✓ CORRECT" : "✗ INCORRECT"}
            </span>
          )}
          <span style={styles.chevron}>{expanded ? "▲" : "▼"}</span>
        </div>
      </div>

      {/* Outcome row */}
      <div style={styles.outcomeRow}>
        <div style={styles.outcome}>
          <span style={styles.outcomeName}>{signal.outcome_name}</span>
          <span style={styles.matchLabel}>{signal.home_team} vs {signal.away_team}</span>
        </div>
        <div style={styles.priceBlock}>
          <span style={styles.priceBefore}>{signal.price_before.toFixed(3)}</span>
          <span style={styles.priceArrow}>→</span>
          <span style={{ ...styles.priceAfter, color: severityColor }}>
            {signal.price_after.toFixed(3)}
          </span>
        </div>
      </div>

      {/* Probability bar - always visible, this is the signature element */}
      <div style={styles.barWrapper}>
        <ProbabilityBar
          before={signal.prob_before}
          after={signal.prob_after}
          direction={signal.direction}
          animate
        />
      </div>

      {/* Expanded content */}
      {expanded && (
        <div style={styles.expanded}>
          {/* AI Analysis */}
          {signal.ai_analysis && (
            <div style={styles.aiBlock}>
              <div style={styles.aiHeader}>
                <span style={styles.aiLabel}>AI ANALYSIS</span>
                <span style={styles.aiPowered}>Claude · Anthropic</span>
              </div>
              <p style={styles.aiText}>{signal.ai_analysis}</p>
            </div>
          )}

          {/* Technical details */}
          <div style={styles.techRow}>
            <TechItem label="BLOCK HASH" value={`${signal.block_hash.slice(0, 16)}...`} mono />
            <TechItem label="FIXTURE ID" value={String(signal.fixture_id)} mono />
            <TechItem label="OUTCOME ID" value={String(signal.outcome_id)} mono />
          </div>

          {signal.final_score && (
            <div style={styles.finalScore}>
              <span style={styles.finalLabel}>FINAL SCORE</span>
              <span style={styles.finalValue}>{signal.final_score}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function TechItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
      <span style={{
        fontFamily: "var(--mono)",
        fontSize: 9,
        color: "var(--text-3)",
        letterSpacing: "0.1em",
      }}>
        {label}
      </span>
      <span style={{
        fontFamily: mono ? "var(--mono)" : "var(--sans)",
        fontSize: 11,
        color: "var(--text-2)",
      }}>
        {value}
      </span>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  card: {
    background: "var(--surface)",
    border: "1px solid var(--border)",
    borderRadius: 8,
    padding: "14px 16px",
    display: "flex",
    flexDirection: "column",
    gap: 10,
    cursor: "default",
  },
  header: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    cursor: "pointer",
    userSelect: "none",
  },
  headerLeft: {
    display: "flex",
    alignItems: "center",
    gap: 8,
  },
  headerRight: {
    display: "flex",
    alignItems: "center",
    gap: 8,
  },
  severityTag: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 700,
    letterSpacing: "0.08em",
    padding: "2px 7px",
    borderRadius: 3,
  },
  direction: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    color: "var(--text-2)",
    fontWeight: 500,
  },
  market: {
    fontSize: 12,
    color: "var(--text-3)",
  },
  time: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    color: "var(--text-3)",
  },
  resolvedTag: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 600,
    letterSpacing: "0.06em",
    padding: "2px 7px",
    borderRadius: 3,
  },
  chevron: {
    fontSize: 10,
    color: "var(--text-3)",
  },
  outcomeRow: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "flex-end",
  },
  outcome: {
    display: "flex",
    flexDirection: "column",
    gap: 2,
  },
  outcomeName: {
    fontSize: 16,
    fontWeight: 600,
    color: "var(--text)",
  },
  matchLabel: {
    fontSize: 11,
    color: "var(--text-3)",
  },
  priceBlock: {
    display: "flex",
    alignItems: "center",
    gap: 8,
  },
  priceBefore: {
    fontFamily: "var(--mono)",
    fontSize: 16,
    color: "var(--text-3)",
    textDecoration: "line-through",
    textDecorationColor: "var(--text-3)",
  },
  priceArrow: {
    color: "var(--text-3)",
    fontSize: 14,
  },
  priceAfter: {
    fontFamily: "var(--mono)",
    fontSize: 20,
    fontWeight: 700,
  },
  barWrapper: {
    padding: "0 2px",
  },
  expanded: {
    display: "flex",
    flexDirection: "column",
    gap: 12,
    paddingTop: 4,
    borderTop: "1px solid var(--border)",
  },
  aiBlock: {
    background: "var(--surface-2)",
    border: "1px solid var(--border-2)",
    borderRadius: 6,
    padding: "10px 12px",
    display: "flex",
    flexDirection: "column",
    gap: 7,
  },
  aiHeader: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
  },
  aiLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    fontWeight: 600,
    color: "var(--accent)",
    letterSpacing: "0.12em",
  },
  aiPowered: {
    fontSize: 10,
    color: "var(--text-3)",
    fontStyle: "italic",
  },
  aiText: {
    fontSize: 13,
    color: "var(--text-2)",
    lineHeight: 1.6,
  },
  techRow: {
    display: "flex",
    gap: 24,
  },
  finalScore: {
    display: "flex",
    alignItems: "center",
    gap: 10,
    padding: "8px 10px",
    background: "var(--accent-dim)",
    border: "1px solid var(--accent-glow)",
    borderRadius: 5,
  },
  finalLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--accent)",
    letterSpacing: "0.1em",
    fontWeight: 600,
  },
  finalValue: {
    fontFamily: "var(--mono)",
    fontSize: 13,
    color: "var(--text)",
    fontWeight: 600,
  },
};
