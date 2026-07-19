import React, { useEffect, useRef } from "react";

interface Props {
  before: number; // 0.0 to 1.0
  after: number;  // 0.0 to 1.0
  direction: "SHORTENING" | "DRIFTING";
  animate?: boolean;
}

export function ProbabilityBar({ before, after, direction, animate = true }: Props) {
  const barRef = useRef<HTMLDivElement>(null);

  const color = direction === "SHORTENING" ? "var(--accent)" : "var(--warn)";
  const beforePct = (before * 100).toFixed(1);
  const afterPct = (after * 100).toFixed(1);
  const deltaPct = (Math.abs(after - before) * 100).toFixed(1);

  useEffect(() => {
    if (!barRef.current || !animate) return;
    const el = barRef.current;

    // Start at before width, animate to after width
    el.style.width = `${before * 100}%`;
    el.style.transition = "none";

    const raf = requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        el.style.transition = "width 0.8s cubic-bezier(0.34, 1.56, 0.64, 1)";
        el.style.width = `${after * 100}%`;
      });
    });

    return () => cancelAnimationFrame(raf);
  }, [before, after, animate]);

  return (
    <div style={styles.wrapper}>
      <div style={styles.labels}>
        <span style={styles.labelBefore}>
          {beforePct}%
        </span>
        <span style={{ ...styles.delta, color }}>
          {direction === "SHORTENING" ? "▲" : "▼"} {deltaPct}%
        </span>
        <span style={{ ...styles.labelAfter, color }}>
          {afterPct}%
        </span>
      </div>

      <div style={styles.track}>
        {/* Ghost bar showing previous position */}
        <div style={{
          ...styles.ghost,
          width: `${before * 100}%`,
        }} />
        {/* Animated live bar */}
        <div
          ref={barRef}
          style={{
            ...styles.bar,
            width: `${after * 100}%`,
            background: color,
            boxShadow: `0 0 8px ${color}66`,
          }}
        />
      </div>

      <div style={styles.scale}>
        {[0, 25, 50, 75, 100].map((v) => (
          <span key={v} style={styles.scaleTick}>{v}%</span>
        ))}
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  wrapper: {
    display: "flex",
    flexDirection: "column",
    gap: 4,
    padding: "10px 0 4px",
  },
  labels: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 2,
  },
  labelBefore: {
    fontFamily: "var(--mono)",
    fontSize: 11,
    color: "var(--text-3)",
    textDecoration: "line-through",
    textDecorationColor: "var(--text-3)",
  },
  delta: {
    fontFamily: "var(--mono)",
    fontSize: 12,
    fontWeight: 600,
  },
  labelAfter: {
    fontFamily: "var(--mono)",
    fontSize: 13,
    fontWeight: 700,
  },
  track: {
    position: "relative",
    height: 8,
    background: "var(--surface-2)",
    borderRadius: 4,
    overflow: "hidden",
    border: "1px solid var(--border)",
  },
  ghost: {
    position: "absolute",
    top: 0,
    left: 0,
    height: "100%",
    background: "var(--border-2)",
    borderRadius: 4,
  },
  bar: {
    position: "absolute",
    top: 0,
    left: 0,
    height: "100%",
    borderRadius: 4,
  },
  scale: {
    display: "flex",
    justifyContent: "space-between",
    marginTop: 2,
  },
  scaleTick: {
    fontFamily: "var(--mono)",
    fontSize: 9,
    color: "var(--text-3)",
  },
};
