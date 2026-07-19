export type Severity = "LOW" | "MEDIUM" | "HIGH";
export type Direction = "SHORTENING" | "DRIFTING";
export type DecisionType = "BACK" | "FADE" | "SKIP";

export interface Signal {
  fixture_id: number;
  home_team: string;
  away_team: string;
  market_name: string;
  outcome_name: string;
  outcome_id: number;
  price_before: number;
  price_after: number;
  prob_before: number;
  prob_after: number;
  prob_delta: number;
  severity: Severity;
  direction: Direction;
  detected_at: string;
  block_hash: string;
  resolved: boolean;
  prediction_correct?: boolean;
  final_score?: string;
  ai_analysis?: string;
}

export interface Decision {
  agent_name: string;
  signal_id: string;
  decision_type: DecisionType;
  reasoning: string;
  stake: number;
  target_price: number;
  decided_at: string;
  settled: boolean;
  won: boolean;
  pnl: number;
  final_score?: string;
}

export interface AgentSummary {
  agent_name: string;
  total_signals: number;
  backed: number;
  skipped: number;
  settled: number;
  won: number;
  lost: number;
  total_staked: number;
  total_pnl: number;
  roi_pct: number;
}

export interface ArenaResults {
  generated_at: string;
  decisions: Decision[];
  summary: Record<string, AgentSummary>;
}

export interface Fixture {
  id: number;
  home_team: string;
  away_team: string;
  kickoff: string;
  status: "upcoming" | "live" | "finished";
  game_state: number;
}

export interface AgentStatus {
  wallet: string;
  service_level: number;
  activated_at: string;
  poll_interval_sec: number;
  movement_threshold: number;
  is_running: boolean;
  ai_interpreter_enabled: boolean;
  total_signals: number;
  last_poll: string;
}
