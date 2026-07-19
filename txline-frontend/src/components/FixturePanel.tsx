import React, { useEffect, useState } from "react";
import type { Fixture, AgentStatus } from "../types";

interface Props {
  fixtures: Fixture[];
  status: AgentStatus;
  selectedId: number | null;
  onSelect: (id: number) => void;
}

export function FixturePanel({ fixtures, status, selectedId, onSelect }: Props) {
  return (
    <aside style={styles.panel}>
      <div style={styles.header}>
        <span style={styles.headerLabel}>FIXTURES</span>
        <span style={styles.headerCount}>{fixtures.length}</span>
      </div>

      <div style={styles.fixtures}>
        {fixtures.map((f) => (
          <FixtureCard
            key={f.id}
            fixture={f}
            selected={selectedId === f.id}
            onSelect={() => onSelect(f.id)}
          />
        ))}
      </div>

      <div style={styles.agentInfo}>
        <div style={styles.infoHeader}>AGENT CONFIG</div>
        <InfoRow label="Wallet" value={`${status.wallet.slice(0, 4)}...${status.wallet.slice(-4)}`} mono />
        <InfoRow label="Poll" value={`${status.poll_interval_sec}s`} mono />
        <InfoRow label="Threshold" value={`${(status.movement_threshold * 100).toFixed(0)}%`} mono />
        <InfoRow label="Activated" value={new Date(status.activated_at).toLocaleDateString()} />
        <InfoRow
          label="AI Layer"
          value={status.ai_interpreter_enabled ? "Enabled" : "Disabled"}
          accent={status.ai_interpreter_enabled}
        />
      </div>

      <div style={styles.footer}>
        <div style={styles.footerRow}>
          <span style={styles.footerLabel}>Powered by</span>
          <span style={styles.footerValue}>TxLINE · Solana</span>
        </div>
        <div style={styles.footerRow}>
          <span style={styles.footerLabel}>Service Level</span>
          <span style={{ ...styles.footerValue, color: "var(--accent)" }}>
            {status.service_level} (Real-time)
          </span>
        </div>
      </div>
    </aside>
  );
}

function FixtureCard({
  fixture,
  selected,
  onSelect,
}: {
  fixture: Fixture;
  selected: boolean;
  onSelect: () => void;
}) {
  const [, setTick] = useState(0);
  const kickoff = new Date(fixture.kickoff);
  const now = new Date();
  const isLive = fixture.status === "live" || kickoff < now;
  const minutesElapsed = Math.floor((now.getTime() - kickoff.getTime()) / 60000);
  const displayMinute = Math.min(minutesElapsed, 90);

  useEffect(() => {
    if (!isLive) return;
    const id = setInterval(() => setTick((t) => t + 1), 60000);
    return () => clearInterval(id);
  }, [isLive]);

  return (
    <button
      onClick={onSelect}
      style={{
        ...styles.card,
        ...(selected ? styles.cardSelected : {}),
        ...(isLive ? styles.cardLive : {}),
      }}
    >
      <div style={styles.cardTop}>
        <span style={{
          ...styles.statusBadge,
          background: isLive ? "var(--accent-dim)" : "var(--surface-2)",
          color: isLive ? "var(--accent)" : "var(--text-3)",
          border: `1px solid ${isLive ? "var(--accent-glow)" : "var(--border)"}`,
        }}>
          {isLive ? "● LIVE" : fixture.status.toUpperCase()}
        </span>
        {isLive && (
          <span style={{
            fontFamily: "var(--mono)",
            fontSize: 12,
            fontWeight: 700,
            color: "var(--accent)",
          }}>
            {displayMinute}'
          </span>
        )}
      </div>

      <div style={styles.teams}>
        <span style={styles.team}>{fixture.home_team}</span>
        <span style={styles.vs}>vs</span>
        <span style={styles.team}>{fixture.away_team}</span>
      </div>

      {isLive && (
        <div style={{ display: "flex", flexDirection: "column", gap: 0 }}>
          <div style={{
            height: 3,
            background: "var(--border)",
            borderRadius: 2,
            overflow: "hidden",
          }}>
            <div style={{
              height: "100%",
              width: `${Math.min((displayMinute / 90) * 100, 100)}%`,
              background: "var(--accent)",
              borderRadius: 2,
              transition: "width 1s linear",
            }} />
          </div>
          <div style={{
            display: "flex",
            justifyContent: "space-between",
            marginTop: 2,
          }}>
            <span style={{ fontFamily: "var(--mono)", fontSize: 9, color: "var(--text-3)" }}>0'</span>
            <span style={{ fontFamily: "var(--mono)", fontSize: 9, color: "var(--text-3)" }}>45'</span>
            <span style={{ fontFamily: "var(--mono)", fontSize: 9, color: "var(--text-3)" }}>90'</span>
          </div>
        </div>
      )}

      <div style={styles.cardBottom}>
        {!isLive ? (
          <span style={styles.kickoffTime}>
            KO: {kickoff.toLocaleDateString()} {kickoff.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })} UTC
          </span>
        ) : (
          <span style={styles.kickoffTime}>World Cup 2026 · Semi-final</span>
        )}
      </div>
    </button>
  );
}

function InfoRow({
  label,
  value,
  mono,
  accent,
}: {
  label: string;
  value: string;
  mono?: boolean;
  accent?: boolean;
}) {
  return (
    <div style={styles.infoRow}>
      <span style={styles.infoLabel}>{label}</span>
      <span style={{
        ...styles.infoValue,
        fontFamily: mono ? "var(--mono)" : "var(--sans)",
        color: accent ? "var(--accent)" : "var(--text-2)",
      }}>
        {value}
      </span>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    width: 220,
    flexShrink: 0,
    background: "var(--surface)",
    borderRight: "1px solid var(--border)",
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
  headerCount: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    color: "var(--text-2)",
    background: "var(--surface-2)",
    padding: "1px 6px",
    borderRadius: 3,
  },
  fixtures: {
    display: "flex",
    flexDirection: "column",
    gap: 1,
    padding: "8px 8px",
  },
  card: {
    display: "flex",
    flexDirection: "column",
    gap: 6,
    padding: "10px 12px",
    background: "var(--surface-2)",
    border: "1px solid var(--border)",
    borderRadius: 6,
    cursor: "pointer",
    textAlign: "left",
    transition: "border-color 0.15s, background 0.15s",
    width: "100%",
  },
  cardSelected: {
    borderColor: "var(--accent-glow)",
    background: "#0F1E2E",
  },
  cardLive: {
    borderColor: "var(--accent-glow)",
  },
  cardTop: {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
  },
  statusBadge: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    fontWeight: 600,
    letterSpacing: "0.08em",
    padding: "2px 6px",
    borderRadius: 3,
  },
  liveMinutes: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
  },
  teams: {
    display: "flex",
    flexDirection: "column",
    gap: 2,
  },
  team: {
    fontSize: 13,
    fontWeight: 500,
    color: "var(--text)",
  },
  vs: {
    fontSize: 10,
    color: "var(--text-3)",
    fontStyle: "italic",
  },
  cardBottom: {},
  kickoffTime: {
    fontSize: 10,
    color: "var(--text-3)",
    fontFamily: "var(--mono)",
  },
  agentInfo: {
    padding: "12px 14px",
    borderTop: "1px solid var(--border)",
    display: "flex",
    flexDirection: "column",
    gap: 8,
    marginTop: "auto",
  },
  infoHeader: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    fontWeight: 600,
    color: "var(--text-3)",
    letterSpacing: "0.12em",
    marginBottom: 2,
  },
  infoRow: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
  },
  infoLabel: {
    fontSize: 11,
    color: "var(--text-3)",
  },
  infoValue: {
    fontSize: 11,
    color: "var(--text-2)",
  },
  footer: {
    padding: "10px 14px",
    borderTop: "1px solid var(--border)",
    display: "flex",
    flexDirection: "column",
    gap: 4,
  },
  footerRow: {
    display: "flex",
    justifyContent: "space-between",
  },
  footerLabel: {
    fontSize: 10,
    color: "var(--text-3)",
  },
  footerValue: {
    fontSize: 10,
    color: "var(--text-2)",
    fontFamily: "var(--mono)",
  },
};
