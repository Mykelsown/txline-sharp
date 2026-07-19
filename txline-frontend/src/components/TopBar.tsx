import React, { useState, useEffect } from "react";
import type { AgentStatus } from "../types";

interface Props {
  status: AgentStatus;
  liveMode: boolean;
  lastFetch: Date | null;
}

export function TopBar({ status, liveMode, lastFetch }: Props) {
  const [, setTick] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 1000);
    return () => clearInterval(id);
  }, []);

  const secondsSincePoll = Math.floor(
    (Date.now() - new Date(status.last_poll).getTime()) / 1000
  );
  const nextPoll = Math.max(0, status.poll_interval_sec - secondsSincePoll);

  return (
    <header style={styles.bar}>
      <div style={styles.left}>
        <div style={styles.logo}>
          <span style={styles.logoAccent}>TX</span>
          <span style={styles.logoText}>LINE</span>
          <span style={styles.logoSub}>SHARP</span>
        </div>
        <div style={styles.divider} />
        <div style={styles.tagline}>World Cup 2026 · Sharp Movement Detector</div>
      </div>

      <div style={styles.right}>
        {/* Live vs Demo mode indicator */}
        <div style={{
          ...styles.modeBadge,
          background: liveMode ? "var(--accent-dim)" : "var(--warn-dim)",
          border: `1px solid ${liveMode ? "var(--accent-glow)" : "var(--warn)44"}`,
        }}>
          <span style={{
            ...styles.modeDot,
            background: liveMode ? "var(--accent)" : "var(--warn)",
            animation: liveMode ? "pulse-dot 1.4s ease-in-out infinite" : "none",
          }} />
          <span style={{
            ...styles.modeText,
            color: liveMode ? "var(--accent)" : "var(--warn)",
          }}>
            {liveMode ? "LIVE" : "DEMO"}
          </span>
          {lastFetch && liveMode && (
            <span style={styles.lastFetch}>
              {lastFetch.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })}
            </span>
          )}
        </div>

        {status.ai_interpreter_enabled && (
          <div style={styles.badge}>
            <span style={styles.aiBadge}>AI</span>
            <span style={styles.badgeText}>Interpreter Active</span>
          </div>
        )}

        <div style={styles.stat}>
          <span style={styles.statLabel}>NEXT POLL</span>
          <span style={{ ...styles.statValue, color: nextPoll < 10 ? "var(--accent)" : "var(--text)" }}>
            {nextPoll}s
          </span>
        </div>

        <div style={styles.stat}>
          <span style={styles.statLabel}>SIGNALS</span>
          <span style={styles.statValue}>{status.total_signals}</span>
        </div>

        <div style={styles.stat}>
          <span style={styles.statLabel}>SVC LEVEL</span>
          <span style={{ ...styles.statValue, color: "var(--accent)" }}>{status.service_level}</span>
        </div>
      </div>
    </header>
  );
}

const styles: Record<string, React.CSSProperties> = {
  bar: {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    padding: "0 20px",
    height: 52,
    background: "var(--surface)",
    borderBottom: "1px solid var(--border)",
    flexShrink: 0,
    gap: 16,
  },
  left: {
    display: "flex",
    alignItems: "center",
    gap: 16,
  },
  logo: {
    display: "flex",
    alignItems: "baseline",
    gap: 2,
    fontFamily: "var(--mono)",
    fontWeight: 700,
    letterSpacing: "-0.02em",
  },
  logoAccent: { color: "var(--accent)", fontSize: 18 },
  logoText: { color: "var(--text)", fontSize: 18 },
  logoSub: {
    color: "var(--text-3)",
    fontSize: 10,
    letterSpacing: "0.12em",
    marginLeft: 6,
    fontWeight: 500,
  },
  divider: {
    width: 1,
    height: 20,
    background: "var(--border-2)",
  },
  tagline: {
    color: "var(--text-2)",
    fontSize: 12,
    letterSpacing: "0.04em",
  },
  right: {
    display: "flex",
    alignItems: "center",
    gap: 16,
  },
  modeBadge: {
    display: "flex",
    alignItems: "center",
    gap: 6,
    padding: "3px 10px",
    borderRadius: 4,
  },
  modeDot: {
    width: 7,
    height: 7,
    borderRadius: "50%",
    display: "inline-block",
  },
  modeText: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    fontWeight: 700,
    letterSpacing: "0.1em",
  },
  lastFetch: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    color: "var(--text-3)",
    marginLeft: 2,
  },
  badge: {
    display: "flex",
    alignItems: "center",
    gap: 6,
    padding: "3px 8px",
    background: "var(--accent-dim)",
    border: "1px solid var(--accent-glow)",
    borderRadius: 4,
  },
  aiBadge: {
    fontFamily: "var(--mono)",
    fontSize: 10,
    fontWeight: 700,
    color: "var(--accent)",
    letterSpacing: "0.06em",
  },
  badgeText: { fontSize: 11, color: "var(--accent)", opacity: 0.8 },
  stat: {
    display: "flex",
    flexDirection: "column",
    alignItems: "flex-end",
    gap: 1,
  },
  statLabel: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
    letterSpacing: "0.1em",
  },
  statValue: {
    fontFamily: "var(--mono)",
    fontSize: 14,
    fontWeight: 600,
    color: "var(--text)",
  },
};
