import { useState, useEffect, useCallback } from "react";
import type {
  Signal,
  ArenaResults,
  AgentStatus,
  Fixture,
} from "../types";
import {
  mockSignals,
  mockArenaResults,
  mockAgentStatus,
  mockFixtures,
} from "../data/mock";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8081";
const POLL_MS = 30_000; // refresh every 30 seconds

// useLiveData fetches all four endpoints and refreshes on a timer.
// Falls back to mock data if the API is unreachable (demo mode).
export function useLiveData() {
  const [status, setStatus]   = useState<AgentStatus>(mockAgentStatus);
  const [fixtures, setFixtures] = useState<Fixture[]>(mockFixtures);
  const [signals, setSignals] = useState<Signal[]>(mockSignals);
  const [arena, setArena]     = useState<ArenaResults>(mockArenaResults);
  const [liveMode, setLiveMode] = useState(false);
  const [lastFetch, setLastFetch] = useState<Date | null>(null);

  const fetchAll = useCallback(async () => {
    try {
      const [statusRes, fixturesRes, signalsRes, arenaRes] = await Promise.all([
        fetch(`${API_BASE}/api/status`),
        fetch(`${API_BASE}/api/fixtures`),
        fetch(`${API_BASE}/api/signals`),
        fetch(`${API_BASE}/api/arena`),
      ]);

      if (!statusRes.ok || !fixturesRes.ok || !signalsRes.ok || !arenaRes.ok) {
        throw new Error("One or more API endpoints returned an error");
      }

      const [statusData, fixturesData, signalsData, arenaData] = await Promise.all([
        statusRes.json(),
        fixturesRes.json(),
        signalsRes.json(),
        arenaRes.json(),
      ]);

      setStatus(statusData);
      setFixtures(fixturesData);
      setSignals(signalsData ?? []);
      setArena(arenaData);
      setLiveMode(true);
      setLastFetch(new Date());
    } catch {
      // API unreachable: stay on mock data, show demo mode indicator.
      setLiveMode(false);
    }
  }, []);

  useEffect(() => {
    fetchAll();
    const id = setInterval(fetchAll, POLL_MS);
    return () => clearInterval(id);
  }, [fetchAll]);

  return { status, fixtures, signals, arena, liveMode, lastFetch, refetch: fetchAll };
}
