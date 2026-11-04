export type Window = { start: string; end: string };
export type Department = {
  id: string;
  name: string;
  color: string;
  goal: string;
};
export type Event = {
  template_id?: string;
  id: string;
  name: string;
  department_id: string;
  contact: string;
  type: string;
  tags: string[];
  start: string;
  requested_start: string;
  duration: number;
  window_start: string;
  window_end: string;
  fixed: boolean;
  status: string;
  constraint_ids: string[];
  fatigue: boolean;
};
export type EventTemplate = Pick<
  Event,
  | "id"
  | "name"
  | "department_id"
  | "contact"
  | "type"
  | "tags"
  | "duration"
  | "constraint_ids"
  | "fatigue"
>;

export type Rule = {
  id: string;
  name: string;
  type: string;
  scope: string;
  confirmed: boolean;
  qualification_id?: string;
  person_id?: string;
  asset_id?: string;
  count?: number;
  start?: string;
  end?: string;
  target_tag?: string;
  target_event_id?: string;
  minutes?: number;
};
export type Assignment = { event_id: string; person_id: string; role: string };
export type Conflict = {
  type: string;
  rule_id?: string;
  event_ids: string[];
  resource_ids: string[];
  start: string;
  end: string;
  message: string;
};
export type Requirement = {
  id: string;
  name: string;
  department_id: string;
  type: string;
  start: string;
  end: string;
  count: number;
  source: string;
  repeat_days: number;
};
export type Coverage = {
  requirement: Requirement;
  completed: number;
  planned: number;
  gap: number;
};
export type Evaluation = {
  feasible: boolean;
  conflicts: Conflict[];
  warnings: Conflict[];
  assignments: Assignment[];
  coverage: Coverage[];
};
export type Data = {
  departments: Department[];
  qualifications: { id: string; name: string }[];
  people: {
    id: string;
    name: string;
    department_id: string;
    qualifications: string[];
    availability: Window[];
  }[];
  duties: {
    id: string;
    name: string;
    person_id: string;
    start: string;
    end: string;
    fatigue: boolean;
  }[];
  assets: {
    id: string;
    name: string;
    capacity: number;
    availability: Window[];
  }[];
  constraints: Rule[];
  events: Event[];
  templates: EventTemplate[];
  requirements: Requirement[];
  completions: {
    id: string;
    department_id: string;
    type: string;
    at: string;
    event_id?: string;
  }[];
  assignments: Assignment[];
};
export type Proposal = {
  id: string;
  week: string;
  status: string;
  progress: string;
  mode: string;
  events: Event[];
  evaluation: Evaluation;
  changes: {
    event_id: string;
    name: string;
    before: string | null;
    after: string;
    reason: string;
  }[];
  previous_assignments: Assignment[];
  negotiation: {
    department_id: string;
    round: number;
    summary: string;
    preferences: { event_id: string; start: string; reason: string }[];
    evaluation: Evaluation;
  }[];
  unscheduled: Event[];
  summary: string;
  alternate: boolean;
  score: number[];
  error?: string;
};
export type Context = {
  data: Data;
  week: string;
  timezone: string;
  ai_available: boolean;
  evaluation: Evaluation;
  runs: { id: string; status: string; mode: string; summary: string }[];
};
