export type Window = { start: string; end: string };
export type Department = {
  id: string;
  name: string;
  color: string;
  goal: string;
};
export type SourceReference = {
  document_id: string;
  page: number;
  excerpt: string;
};
export type Event = {
  ignore_conflicts?: boolean;
  ignore_conflicts_description?: string;
  source_ref?: SourceReference;
  description?: string;
  template_id?: string;
  id: string;
  name: string;
  department_id: string;
  tier: number;
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
  | "tier"
  | "contact"
  | "type"
  | "tags"
  | "duration"
  | "constraint_ids"
  | "fatigue"
  | "source_ref"
  | "description"
>;

export type Rule = {
  source_ref?: SourceReference;
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
export type ConstraintDraft = {
  rule: Rule;
  requested_qualification: string;
  explanation: string;
};
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
  source_ref?: SourceReference;
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

export type TrainingSuggestion = {
  training_activity?: string;
  duration_evidence?: string;
  citation_status?: "matched" | "needs_review";
  id: string;
  kind: "template" | "requirement" | "constraint";
  name: string;
  description: string;
  basis: "required" | "recommended";
  department_id: string;
  contact: string;
  type: string;
  duration: number;
  tags: string[];
  constraint_ids: string[];
  fatigue: boolean;
  count: number;
  repeat_days: number;
  start: string;
  end: string;
  page: number;
  excerpt: string;
  review_notes: string;
  missing: string[];
  rule?: Rule;
  created_id?: string;
  deleted?: boolean;
};
export type TrainingDocument = {
  id: string;
  filename: string;
  size: number;
  sha256: string;
  imported_at: string;
  status: string;
  error?: string;
  title: string;
  summary: string;
  suggestions: TrainingSuggestion[];
};
