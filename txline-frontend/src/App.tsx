import React, { useState } from "react";
import { TopBar } from "./components/TopBar";
import { FixturePanel } from "./components/FixturePanel";
import { SignalFeed } from "./components/SignalFeed";
import { ArenaPanel } from "./components/ArenaPanel";
import { useLiveData } from "./hooks/useLiveData";

export default function App() {
  const [selectedFixtureId, setSelectedFixtureId] = useState<number | null>(null);
  const { status, fixtures, signals, arena, liveMode, lastFetch } = useLiveData();

  return (
    <div style={styles.root}>
      <TopBar status={status} liveMode={liveMode} lastFetch={lastFetch} />

      <div style={styles.body}>
        <FixturePanel
          fixtures={fixtures}
          status={status}
          selectedId={selectedFixtureId}
          onSelect={(id) =>
            setSelectedFixtureId((prev) => (prev === id ? null : id))
          }
        />

        <SignalFeed
          signals={signals}
          selectedFixtureId={selectedFixtureId}
        />

        <ArenaPanel results={arena} />
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  root: {
    height: "100vh",
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
    background: "var(--bg)",
  },
  body: {
    flex: 1,
    display: "flex",
    overflow: "hidden",
    minHeight: 0,
  },
};
