import React from "react";
import type { ArenaResults, Decision } from "../types";

interface Props {
  results: ArenaResults;
}

export function ArenaPanel({ results }: Props) {
  const decisions = Array.isArray(results?.decisions) ? results.decisions : [];
  const summary = results?.summary ?? {};
  const summaries = Object.values(summary);
  const agentA = summaries.find((s) => s.agent_name.includes("Momentum"));
  const agentB = summaries.find((s) => s.agent_name.includes("Contrarian"));

  const recentDecisions = [...decisions]
    .sort((a, b) => new Date(b.decided_at).getTime() - new Date(a.decided_at).getTime())
    .slice(0, 8);

  return (
    <aside style={styles.panel}>
      <div style={styles.header}>
        <span style={styles.headerLabel}>ARENA</span>
        <span style={styles.headerSub}>Sim · Backtested</span>
      </div>

      {/* Agent comparison */}
      <div style={styles.agents}>
        {agentA && <AgentCard summary={agentA} color="var(--accent)" label="A" />}
        {agentB && <AgentCard summary={agentB} color="var(--warn)" label="B" />}
      </div>

      {/* VS divider */}
      <div style={styles.vsDivider}>
        <div style={styles.vsLine} />
        <span style={styles.vsText}>VS</span>
        <div style={styles.vsLine} />
      </div>

      {/* Strategy descriptions */}
      <div style={styles.strategies}>
        <StrategyRow
          label="Momentum"
          desc="Follows sharp money"
          color="var(--accent)"
        />
        <StrategyRow
          label="Contrarian"
          desc="Fades the move"
          color="var(--warn)"
        />
      </div>

      {/* Recent decisions */}
      <div style={styles.decisionsHeader}>
        <span style={styles.headerLabel}>RECENT DECISIONS</span>
      </div>

      <div style={styles.decisions}>
        {recentDecisions.length === 0 ? (
          <div style={styles.empty}>
            <span style={styles.emptyText}>Waiting for signals...</span>
            <span style={styles.emptySubtext}>Agents activate when odds move</span>
          </div>
        ) : (
          recentDecisions.map((d, i) => (
            <DecisionRow key={i} decision={d} />
          ))
        )}
      </div>

      <div style={styles.disclaimer}>
        Hypothetical P&L only. Simulated on historical signal data.
        Not financial advice.
      </div>
    </aside>
  );
}

function AgentCard({
  summary,
  color,
  label,
}: {
  summary: { agent_name: string; total_signals: number; backed: number; skipped: number; settled: number; won: number; total_pnl: number; roi_pct: number };
  color: string;
  label: string;
}) {
  const pnlPositive = summary.total_pnl >= 0;

  return (
    <div style={{ ...styles.agentCard, borderColor: `${color}44` }}>
      <div style={styles.agentHeader}>
        <div style={{ ...styles.agentLabel, background: color, color: "#0A0E17" }}>
          {label}
        </div>
        <span style={styles.agentName}>
          {summary.agent_name.replace("Agent-A ", "").replace("Agent-B ", "")}
        </span>
      </div>

      <div style={styles.agentStats}>
        <Stat label="SIGNALS" value={String(summary.total_signals)} />
        <Stat label="ACTIVE" value={String(summary.backed)} />
        <Stat label="SETTLED" value={String(summary.settled)} />
        <Stat label="WON" value={String(summary.won)} color={color} />
      </div>

      <div style={styles.pnlRow}>
        <span style={styles.pnlLabel}>P&L</span>
        <span style={{
          ...styles.pnlValue,
          color: pnlPositive ? "var(--accent)" : "var(--danger)",
        }}>
          {pnlPositive ? "+" : ""}{summary.total_pnl.toFixed(2)} USDT
        </span>
      </div>

      {summary.settled > 0 && (
        <div style={styles.roiRow}>
          <span style={styles.roiLabel}>ROI</span>
          <span style={{
            ...styles.roiValue,
            color: summary.roi_pct >= 0 ? "var(--accent)" : "var(--danger)",
          }}>
            {summary.roi_pct >= 0 ? "+" : ""}{summary.roi_pct.toFixed(1)}%
          </span>
        </div>
      )}
    </div>
  );
}

function Stat({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div style={styles.stat}>
      <span style={styles.statLabel}>{label}</span>
      <span style={{ ...styles.statValue, color: color ?? "var(--text)" }}>{value}</span>
    </div>
  );
}

function StrategyRow({ label, desc, color }: { label: string; desc: string; color: string }) {
  return (
    <div style={styles.strategyRow}>
      <div style={{ ...styles.strategyDot, background: color }} />
      <div>
        <span style={{ ...styles.strategyLabel, color }}>{label}</span>
        <span style={styles.strategyDesc}> · {desc}</span>
      </div>
    </div>
  );
}

function DecisionRow({ decision }: { decision: Decision }) {
  const [expanded, setExpanded] = React.useState(false);
  const isA = decision.agent_name.includes("Momentum");
  const color = isA ? "var(--accent)" : "var(--warn)";
  const time = new Date(decision.decided_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div style={{ ...styles.decisionRow, flexDirection: "column", gap: 0 }}>
      <div
        style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer", padding: "6px 6px" }}
        onClick={() => setExpanded((e) => !e)}
      >
        <div style={{ ...styles.decisionAgent, color, borderColor: `${color}44` }}>
          {isA ? "A" : "B"}
        </div>
        <div style={styles.decisionInfo}>
          <span style={{
            ...styles.decisionType,
            color: decision.decision_type === "SKIP" ? "var(--text-3)" : color,
          }}>
            {decision.decision_type}
          </span>
          <span style={styles.decisionTime}>{time}</span>
        </div>
        {decision.decision_type !== "SKIP" && (
          <div style={styles.decisionRight}>
            <span style={styles.decisionStake}>${decision.stake}</span>
            {decision.settled && (
              <span style={{
                fontSize: 11,
                fontFamily: "var(--mono)",
                color: decision.won ? "var(--accent)" : "var(--danger)",
              }}>
                {decision.pnl >= 0 ? "+" : ""}{decision.pnl.toFixed(0)}
              </span>
            )}
          </div>
        )}
        <span style={{ fontSize: 9, color: "var(--text-3)", marginLeft: "auto" }}>
          {expanded ? "▲" : "▼"}
        </span>
      </div>
      {expanded && decision.reasoning && (
        <div style={{
          padding: "6px 8px 8px",
          borderTop: "1px solid var(--border)",
          fontSize: 11,
          color: "var(--text-2)",
          lineHeight: 1.5,
          fontStyle: "italic",
        }}>
          {decision.reasoning}
        </div>
      )}
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    width: 240,
    flexShrink: 0,
    background: "var(--surface)",
    borderLeft: "1px solid var(--border)",
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
  },
  header: {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    padding: "12px 14px 8px",
    borderBottom: "1px solid var(--border)",
  },
  headerLabel: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 600,
    color: "var(--text-3)",
    letterSpacing: "0.12em",
  },
  headerSub: {
    fontSize: 10,
    color: "var(--text-3)",
    fontStyle: "italic",
  },
  agents: {
    display: "flex",
    flexDirection: "column",
    gap: 8,
    padding: "10px 10px 0",
  },
  agentCard: {
    background: "var(--surface-2)",
    border: "1px solid",
    borderRadius: 6,
    padding: "10px 12px",
    display: "flex",
    flexDirection: "column",
    gap: 8,
  },
  agentHeader: {
    display: "flex",
    alignItems: "center",
    gap: 8,
  },
  agentLabel: {
    width: 20,
    height: 20,
    borderRadius: 4,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontFamily: "var(--mono)",
    fontSize: 11,
    fontWeight: 700,
  },
  agentName: {
    fontSize: 12,
    fontWeight: 500,
    color: "var(--text-2)",
  },
  agentStats: {
    display: "flex",
    gap: 10,
  },
  stat: {
    display: "flex",
    flexDirection: "column",
    gap: 1,
  },
  statLabel: {
    fontFamily: "var(--mono)",
    fontSize: 8,
    color: "var(--text-3)",
    letterSpacing: "0.1em",
  },
  statValue: {
    fontFamily: "var(--mono)",
    fontSize: 14,
    fontWeight: 600,
  },
  pnlRow: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    paddingTop: 4,
    borderTop: "1px solid var(--border)",
  },
  pnlLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
    letterSpacing: "0.1em",
  },
  pnlValue: {
    fontFamily: "var(--mono)",
    fontSize: 14,
    fontWeight: 700,
  },
  roiRow: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
  },
  roiLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
    letterSpacing: "0.1em",
  },
  roiValue: {
    fontFamily: "var(--mono)",
    fontSize: 12,
    fontWeight: 600,
  },
  vsDivider: {
    display: "flex",
    alignItems: "center",
    gap: 8,
    padding: "8px 14px",
  },
  vsLine: {
    flex: 1,
    height: 1,
    background: "var(--border)",
  },
  vsText: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    color: "var(--text-3)",
    letterSpacing: "0.12em",
  },
  strategies: {
    display: "flex",
    flexDirection: "column",
    gap: 6,
    padding: "0 14px 10px",
  },
  strategyRow: {
    display: "flex",
    alignItems: "center",
    gap: 8,
  },
  strategyDot: {
    width: 6,
    height: 6,
    borderRadius: "50%",
    flexShrink: 0,
  },
  strategyLabel: {
    fontSize: 12,
    fontWeight: 500,
  },
  strategyDesc: {
    fontSize: 11,
    color: "var(--text-3)",
  },
  decisionsHeader: {
    padding: "8px 14px 6px",
    borderTop: "1px solid var(--border)",
  },
  decisions: {
    flex: 1,
    overflowY: "scroll",
    overflowX: "hidden",
    display: "flex",
    flexDirection: "column",
    gap: 2,
    padding: "4px 8px",
    maxHeight: 320,
  },
  empty: {
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
    gap: 4,
    padding: "24px 0",
  },
  emptyText: {
    fontSize: 12,
    color: "var(--text-2)",
  },
  emptySubtext: {
    fontSize: 11,
    color: "var(--text-3)",
  },
  decisionRow: {
    display: "flex",
    alignItems: "center",
    gap: 8,
    padding: "6px 6px",
    background: "var(--surface-2)",
    borderRadius: 4,
    border: "1px solid var(--border)",
  },
  decisionAgent: {
    width: 18,
    height: 18,
    borderRadius: 3,
    border: "1px solid",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 700,
    flexShrink: 0,
  },
  decisionInfo: {
    flex: 1,
    display: "flex",
    justifyContent: "space-between",
  },
  decisionType: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    fontWeight: 600,
    letterSpacing: "0.04em",
  },
  decisionTime: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    color: "var(--text-3)",
  },
  decisionRight: {
    display: "flex",
    flexDirection: "column",
    alignItems: "flex-end",
    gap: 1,
  },
  decisionStake: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    color: "var(--text-2)",
  },
  disclaimer: {
    padding: "8px 12px",
    borderTop: "1px solid var(--border)",
    fontSize: 9,
    color: "var(--text-3)",
    lineHeight: 1.5,
  },
};
