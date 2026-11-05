import React, { useEffect, useState, useRef } from "react";
import { createRoot } from "react-dom/client";
import {
  Anchor,
  CalendarDays,
  Users,
  ShieldCheck,
  Layers,
  ChevronLeft,
  ChevronRight,
  Plus,
  X,
  LockKeyhole,
  ArrowRight,
  Check,
  AlertTriangle,
  WandSparkles,
  LoaderCircle,
  Settings2,
  ArrowUpRight,
  Trash2,
  GripVertical,
  CheckCheck,
  Database,
  FileText,
} from "lucide-react";
import type {
  Context,
  Data,
  Event,
  EventTemplate,
  Rule,
  ConstraintDraft,
  Evaluation,
  Proposal,
  Window,
} from "./types";
import "./style.css";
import { api } from "./api";
import { Documents, SourceLink } from "./Documents";
import { RuleFields } from "./RuleFields";
import { calendarEntries } from "./calendar-overlay";

const uid = () => crypto.randomUUID();
let zone = "America/New_York";
function localInput(value: string) {
  if (!value || value.startsWith("0001")) return "";
  const p = new Intl.DateTimeFormat("en-CA", {
    timeZone: zone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).formatToParts(new Date(value));
  const v = (key: string) => p.find((x) => x.type === key)?.value;
  return `${v("year")}-${v("month")}-${v("day")}T${v("hour")}:${v("minute")}`;
}
function iso(value: string) {
  if (!value) return "0001-01-01T00:00:00Z";
  let date = new Date(value + "Z");
  for (let i = 0; i < 3; i++) {
    const diff =
      new Date(localInput(date.toISOString()) + "Z").getTime() -
      new Date(value + "Z").getTime();
    date = new Date(date.getTime() - diff);
  }
  return date.toISOString();
}
const stamp = (value: string) =>
  new Intl.DateTimeFormat("en-US", {
    timeZone: zone,
    weekday: "short",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
const time = (value: string) =>
  new Intl.DateTimeFormat("en-US", {
    timeZone: zone,
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
const dateOnly = (value: string) => localInput(value).slice(0, 10);
function addDays(date: string, n: number) {
  const t = new Date(date.slice(0, 10) + "T12:00:00Z");
  t.setUTCDate(t.getUTCDate() + n);
  return t.toISOString().slice(0, 10);
}
const minutes = (value: string) => {
  const s = localInput(value);
  return Number(s.slice(11, 13)) * 60 + Number(s.slice(14, 16));
};
const dep = (data: Data, id: string) =>
  data.departments.find((x) => x.id === id);
const personName = (data: Data, id: string) =>
  data.people.find((x) => x.id === id)?.name || id;
function newEvent(data: Data, week: string, start?: string): Event {
  const day = dateOnly(week);
  const s = start || iso(day + "T10:00");
  return {
    id: uid(),
    name: "",
    department_id: data.departments[0]?.id || "",
    contact: "",
    type: "drill",
    tags: ["drill"],
    start: s,
    requested_start: s,
    duration: 60,
    window_start: iso(day + "T08:00"),
    window_end: iso(addDays(day, 4) + "T17:00"),
    fixed: false,
    status: "draft",
    constraint_ids: [],
    fatigue: true,
  };
}

function templateDefaults(e: EventTemplate): EventTemplate {
  return {
    id: e.id,
    name: e.name,
    department_id: e.department_id,
    contact: e.contact,
    type: e.type,
    duration: e.duration,
    tags: [...e.tags],
    constraint_ids: [...e.constraint_ids],
    fatigue: e.fatigue,
    source_ref: e.source_ref,
    description: e.description,
  };
}
function occurrence(t: EventTemplate, data: Data, week: string): Event {
  return {
    ...newEvent(data, week),
    ...templateDefaults(t),
    id: uid(),
    template_id: t.id,
  };
}

function App() {
  const drag = useRef<{
    event: Event;
    x: number;
    y: number;
    offset: number;
    moved: boolean;
  } | null>(null);
  const [ctx, setCtx] = useState<Context | null>(null),
    [week, setWeek] = useState(""),
    [page, setPage] = useState("calendar"),
    [error, setError] = useState(""),
    [filters, setFilters] = useState<string[]>([]),
    [editing, setEditing] = useState<Event | null>(null),
    [template, setTemplate] = useState(false),
    [choosingTemplate, setChoosingTemplate] = useState(false),
    [proposal, setProposal] = useState<Proposal | null>(null),
    [pollingPaused, setPollingPaused] = useState(false),
    [busy, setBusy] = useState(false),
    [toast, setToast] = useState(""),
    [viewRun, setViewRun] = useState(false),
    [overlayHidden, setOverlayHidden] = useState(false),
    [highlightedEventId, setHighlightedEventId] = useState<string | null>(null);
  const load = async () => {
    const c = await api<Context>(
      "/api/context" + (week ? "?week=" + week : ""),
    );
    zone = c.timezone;
    setCtx(c);
  };
  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, [week]);
  useEffect(() => {
    if (!proposal || proposal.status !== "running" || pollingPaused) return;
    let active = true;
    let inFlight = false;
    const id = setInterval(() => {
      if (inFlight) return;
      inFlight = true;
      api<Proposal>("/api/runs/" + proposal.id)
        .then((p) => {
          if (active) setProposal(p);
        })
        .catch((e) => {
          if (!active) return;
          clearInterval(id);
          setPollingPaused(true);
          setError(e.message);
        })
        .finally(() => {
          inFlight = false;
        });
    }, 900);
    return () => {
      active = false;
      clearInterval(id);
    };
  }, [proposal?.id, proposal?.status, pollingPaused]);
  useEffect(() => {
    if (!toast) return;
    const id = setTimeout(() => setToast(""), 3500);
    return () => clearTimeout(id);
  }, [toast]);
  if (!ctx)
    return (
      <div className="loading">
        <Anchor size={40} />
        <h2>Opening your planning workspace</h2>
        {error ? (
          <p className="error">{error}</p>
        ) : (
          <LoaderCircle className="spin" />
        )}
      </div>
    );
  const data = ctx.data,
    day = dateOnly(ctx.week),
    dates = Array.from({ length: 7 }, (_, i) => addDays(day, i));
  const events = data.events.filter(
    (e) => dateOnly(e.start) >= day && dateOnly(e.start) <= dates[6],
  );
  const conflicting = new Set(
    ctx.evaluation.conflicts.flatMap((c) => c.event_ids),
  );
  const reviewingProposal = !!(
    viewRun &&
    proposal &&
    ["feasible", "partial"].includes(proposal.status) &&
    dateOnly(proposal.week) === day
  );
  const showingOverlay = reviewingProposal && !overlayHidden;
  const proposedEvents = showingOverlay
    ? proposal!.events.filter(
        (e) => dateOnly(e.start) >= day && dateOnly(e.start) <= dates[6],
      )
    : null;
  const entries = calendarEntries(events, proposedEvents);
  const activeHighlight =
    reviewingProposal &&
    entries.some(
      ({ event }) =>
        event.id === highlightedEventId &&
        (!filters.length || filters.includes(event.department_id)),
    )
      ? highlightedEventId
      : null;
  const proposedConflicts = new Set(
    proposal?.evaluation?.conflicts?.flatMap((c) => c.event_ids) || [],
  );
  const coverage = ctx.evaluation.coverage;
  async function resolve(event?: Event) {
    setBusy(true);
    setError("");
    try {
      const p = await api<Proposal>("/api/runs", "POST", {
        week: day,
        event,
        agents: ctx!.ai_available,
      });
      setProposal(p);
      setOverlayHidden(false);
      setPollingPaused(false);
      setEditing(null);
      setViewRun(true);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function apply() {
    if (!proposal) return;
    setBusy(true);
    try {
      const p = await api<Proposal>(
        "/api/runs/" + proposal.id + "/apply",
        "POST",
      );
      setProposal(p);
      await load();
      setToast("Weekly plan applied");
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  function open(e: Event, t = false) {
    setEditing(structuredClone(e));
    setTemplate(t);
  }
  function finishMove(ev: React.PointerEvent<HTMLButtonElement>) {
    const pending = drag.current;
    if (!pending?.moved) return;
    const column = document
      .elementFromPoint(ev.clientX, ev.clientY)
      ?.closest<HTMLElement>(".day-column");
    if (!column?.dataset.day) return;
    const box = column.getBoundingClientRect();
    const slot =
      Math.round(((ev.clientY - box.top - pending.offset) / hourHeight) * 4) *
        15 +
      startHour * 60;
    const start = iso(
      column.dataset.day +
        "T" +
        String(Math.floor(slot / 60)).padStart(2, "0") +
        ":" +
        String(slot % 60).padStart(2, "0"),
    );
    open({ ...pending.event, start, requested_start: start });
  }
  const startHour = Math.min(
      7,
      ...entries.map(({ event: e }) => Math.floor(minutes(e.start) / 60)),
    ),
    endHour = Math.max(
      19,
      ...entries.map(({ event: e }) =>
        Math.ceil((minutes(e.start) + e.duration) / 60),
      ),
    ),
    hourHeight = 72;
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <a
          className="brand"
          href="#"
          onClick={(e) => {
            e.preventDefault();
            setPage("calendar");
          }}
        >
          <span className="brand-mark">
            <Anchor size={26} />
          </span>
          <span>
            PB4T<span className="brand-sub">PLANNING WORKSPACE</span>
          </span>
        </a>
        <div className="ship-label">
          <span className="status-dot" /> SYNTHETIC SHIP · DEMO
        </div>
        <nav>
          {[
            ["calendar", "Weekly calendar", CalendarDays],
            ["training", "Training readiness", ShieldCheck],
            ["data", "Planning inputs", Layers],
            ["documents", "Training documents", FileText],
          ].map(([id, label, Icon]) => (
            <button
              key={id as string}
              className={page === id ? "nav-item active" : "nav-item"}
              onClick={() => setPage(id as string)}
            >
              <Icon size={19} />
              {label as string}
              {page === id && <span className="nav-dot" />}
            </button>
          ))}
        </nav>
        <div className="sidebar-section">
          <span className="eyebrow">DEPARTMENTS</span>
          {data.departments.map((d) => (
            <button
              className="department-filter"
              key={d.id}
              onClick={() => {
                setPage("calendar");
                setFilters(
                  filters.includes(d.id)
                    ? filters.filter((x) => x !== d.id)
                    : [...filters, d.id],
                );
              }}
            >
              <span
                className="dept-dot"
                style={{
                  background: d.color,
                  opacity: filters.length && !filters.includes(d.id) ? 0.3 : 1,
                }}
              />
              {d.name}
              {filters.includes(d.id) && <Check size={13} />}
            </button>
          ))}
        </div>
        <div className="sidebar-bottom">
          <div className="local-badge">
            <Database size={15} /> Local workspace{" "}
            <span className="status-dot" />
          </div>
          <p>
            Illustrative data & policies.
            <br />
            For planning demonstrations.
          </p>
          <div className="profile">
            <div className="avatar">TP</div>
            <div>
              Training planner<small>Shipboard coordination</small>
            </div>
          </div>
        </div>
      </aside>
      <main>
        <header className="topbar">
          <span>
            WORKSPACE <span className="slash">/</span> TRAINING & READINESS
          </span>
          <span className="connection">
            <span className="status-dot" /> SQLite connected{" "}
            <span className="top-separator" /> {ctx.timezone}
          </span>
        </header>
        <div className="workspace">
          <div className="heading">
            <div>
              <div className="eyebrow green">PLAN TOGETHER. OPERATE READY.</div>
              <h1>
                {page === "calendar"
                  ? "Your week, coordinated."
                  : page === "training"
                    ? "Readiness starts here."
                    : page === "documents"
                      ? "Training, from the source."
                      : "A shared planning foundation."}
              </h1>
              <p>
                {page === "calendar"
                  ? "Bring every department into one clear, workable plan."
                  : page === "training"
                    ? "Track what’s complete, what’s planned, and what still needs a place."
                    : page === "documents"
                      ? "Import reference documents and review suggested training and constraints."
                      : "The people, resources, and rules behind a feasible schedule."}
              </p>
            </div>
            {page === "calendar" && (
              <div className="heading-actions">
                <button
                  className="button"
                  onClick={() => setChoosingTemplate(true)}
                >
                  <Plus size={16} /> Add event
                </button>
                <button
                  className="button primary"
                  onClick={() => resolve()}
                  disabled={busy}
                >
                  <WandSparkles size={16} /> Generate week
                </button>
              </div>
            )}
          </div>
          {error && (
            <div className="notice error">
              <AlertTriangle size={18} />
              {error}
              <button className="icon-button" onClick={() => setError("")}>
                <X size={16} />
              </button>
            </div>
          )}
          {page === "calendar" ? (
            <div className="calendar-layout">
              <section className="calendar-card">
                <div className="calendar-toolbar">
                  <div className="week-title">
                    <CalendarDays size={18} />
                    <strong>
                      {new Intl.DateTimeFormat("en-US", {
                        timeZone: zone,
                        month: "long",
                        year: "numeric",
                      }).format(new Date(ctx.week))}
                    </strong>
                    <span className="pill">
                      Week{" "}
                      {Math.ceil(
                        (new Date(ctx.week).getTime() -
                          new Date(
                            new Date(ctx.week).getFullYear(),
                            0,
                            1,
                          ).getTime()) /
                          604800000,
                      )}
                    </span>
                  </div>
                  <div className="toolbar-right">
                    <button className="text-button" onClick={() => setWeek("")}>
                      This week
                    </button>
                    <button
                      className="icon-button"
                      aria-label="Previous week"
                      onClick={() => setWeek(addDays(day, -7))}
                    >
                      <ChevronLeft size={17} />
                    </button>
                    <button
                      className="icon-button"
                      aria-label="Next week"
                      onClick={() => setWeek(addDays(day, 7))}
                    >
                      <ChevronRight size={17} />
                    </button>
                    <span className="view-label">
                      Week <CalendarDays size={14} />
                    </span>
                  </div>
                </div>
                {reviewingProposal && (
                  <div className="calendar-comparison">
                    <div>
                      <strong>
                        {proposal?.status === "partial"
                          ? "Partial proposal overlay"
                          : "Proposal overlay"}
                      </strong>
                      <span>
                        Compare current positions with proposed changes before
                        applying.
                      </span>
                    </div>
                    <button
                      className="button"
                      aria-pressed={!overlayHidden}
                      onClick={() => setOverlayHidden(!overlayHidden)}
                    >
                      {overlayHidden ? "Show proposal" : "Hide proposal"}
                    </button>
                  </div>
                )}
                <div className="calendar-scroll">
                  <div className="calendar-grid" style={{ minWidth: 760 }}>
                    <div className="day-headers">
                      <div className="timezone">
                        SHIP
                        <br />
                        LOCAL
                      </div>
                      {dates.map((d) => (
                        <div
                          className={
                            "day-header " +
                            (d === dateOnly(new Date().toISOString())
                              ? "today"
                              : "")
                          }
                          key={d}
                        >
                          <span>
                            {new Date(d + "T12:00:00Z")
                              .toLocaleDateString("en-US", {
                                weekday: "short",
                                timeZone: "UTC",
                              })
                              .toUpperCase()}
                          </span>
                          <strong>{Number(d.slice(-2))}</strong>
                        </div>
                      ))}
                    </div>
                    <div
                      className="calendar-body"
                      style={{ height: (endHour - startHour) * hourHeight }}
                    >
                      <div className="time-axis">
                        {Array.from({ length: endHour - startHour }, (_, i) => (
                          <span key={i} style={{ top: i * hourHeight }}>
                            {String(i + startHour).padStart(2, "0")}:00
                          </span>
                        ))}
                      </div>
                      {dates.map((d, index) => {
                        const visible = entries
                          .filter(
                            ({ event: e }) =>
                              dateOnly(e.start) === d &&
                              (!filters.length ||
                                filters.includes(e.department_id)),
                          )
                          .sort(
                            (a, b) =>
                              Date.parse(a.event.start) -
                              Date.parse(b.event.start),
                          );
                        const positions = layout(
                          visible.map(({ key, event }) => ({
                            id: key,
                            start: event.start,
                            duration: event.duration,
                          })),
                        );
                        return (
                          <div
                            className={
                              "day-column " + (index > 4 ? "weekend" : "")
                            }
                            key={d}
                            data-day={d}
                            style={{ backgroundSize: `100% ${hourHeight}px` }}
                          >
                            {visible.map((entry) => {
                              const e = entry.event;
                              const department = dep(data, e.department_id),
                                pos = positions.get(entry.key)!,
                                label =
                                  entry.layer === "proposed"
                                    ? entry.added
                                      ? "New"
                                      : "Proposed"
                                    : entry.layer === "current"
                                      ? "Current"
                                      : "",
                                hasConflict =
                                  showingOverlay && entry.layer !== "current"
                                    ? proposedConflicts.has(e.id)
                                    : conflicting.has(e.id);
                              const description = `${label ? label + ": " : ""}${e.name} · ${stamp(e.start)} · ${e.duration} min`;
                              return (
                                <button
                                  key={entry.key}
                                  aria-label={description}
                                  title={description}
                                  draggable={false}
                                  onPointerDown={(ev) => {
                                    drag.current = null;
                                    if (e.fixed || reviewingProposal) return;
                                    drag.current = {
                                      event: e,
                                      x: ev.clientX,
                                      y: ev.clientY,
                                      offset:
                                        ev.clientY -
                                        ev.currentTarget.getBoundingClientRect()
                                          .top,
                                      moved: false,
                                    };
                                    ev.currentTarget.setPointerCapture(
                                      ev.pointerId,
                                    );
                                  }}
                                  onPointerMove={(ev) => {
                                    if (
                                      drag.current &&
                                      ev.buttons &&
                                      Math.hypot(
                                        ev.clientX - drag.current.x,
                                        ev.clientY - drag.current.y,
                                      ) > 8
                                    )
                                      drag.current.moved = true;
                                  }}
                                  onPointerUp={finishMove}
                                  onClick={() => {
                                    if (reviewingProposal) {
                                      setToast(description);
                                      return;
                                    }
                                    if (!drag.current?.moved) open(e);
                                  }}
                                  className={
                                    "calendar-event " +
                                    (e.fixed ? "fixed " : "") +
                                    (e.status === "draft" &&
                                    entry.layer !== "proposed"
                                      ? "draft "
                                      : "") +
                                    (hasConflict ? "conflicted " : "") +
                                    (entry.layer === "proposed"
                                      ? "proposed"
                                      : entry.layer === "current"
                                        ? "current-position"
                                        : "") +
                                    (activeHighlight
                                      ? e.id === activeHighlight
                                        ? " change-highlighted"
                                        : " change-muted"
                                      : "")
                                  }
                                  style={
                                    {
                                      top:
                                        ((minutes(e.start) - startHour * 60) /
                                          60) *
                                          hourHeight +
                                        3,
                                      height: Math.max(
                                        28,
                                        (e.duration / 60) * hourHeight - 6,
                                      ),
                                      left: `calc(${(pos.lane / pos.total) * 100}% + 3px)`,
                                      width: `calc(${100 / pos.total}% - 6px)`,
                                      "--event-color":
                                        department?.color || "#607e74",
                                    } as React.CSSProperties
                                  }
                                >
                                  <span className="event-time">
                                    {e.fixed ? (
                                      <LockKeyhole size={10} />
                                    ) : !reviewingProposal ? (
                                      <GripVertical size={10} />
                                    ) : null}{" "}
                                    {time(e.start)}{" "}
                                    {hasConflict && <AlertTriangle size={10} />}
                                    {label && (
                                      <span className="event-position-label">
                                        {label}
                                      </span>
                                    )}
                                  </span>
                                  <strong>{e.name}</strong>
                                  {e.duration >= 60 && (
                                    <span className="event-dept">
                                      {department?.name}
                                    </span>
                                  )}
                                  {e.status === "draft" &&
                                    entry.layer !== "proposed" &&
                                    e.duration >= 60 && (
                                      <span className="draft-label">DRAFT</span>
                                    )}
                                </button>
                              );
                            })}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>
                <div className="calendar-footer">
                  {showingOverlay ? (
                    <>
                      <span>
                        <span className="legend-current" /> Current position
                      </span>
                      <span>
                        <span className="legend-proposed" /> Proposed position
                      </span>
                      <span>Unchanged events shown once</span>
                    </>
                  ) : (
                    <span>
                      <span className="legend-draft" /> Draft event
                    </span>
                  )}
                  <span>
                    <LockKeyhole size={12} /> Fixed commitment
                  </span>
                  {!reviewingProposal && (
                    <span>
                      <GripVertical size={13} /> Drag an event to preview a move
                    </span>
                  )}
                  {filters.length > 0 && (
                    <button
                      className="text-button"
                      onClick={() => setFilters([])}
                    >
                      Clear filters
                    </button>
                  )}
                </div>
              </section>
              <aside className="insight-column">
                {proposal && viewRun && dateOnly(proposal.week) === day ? (
                  <ProposalPanel
                    proposal={proposal}
                    data={data}
                    busy={busy}
                    comparisonActive={reviewingProposal}
                    highlightedEventId={activeHighlight}
                    onHighlight={setHighlightedEventId}
                    apply={apply}
                    pollingPaused={pollingPaused}
                    retry={() => {
                      setError("");
                      setPollingPaused(false);
                    }}
                    cancel={() => {
                      setViewRun(false);
                      setProposal(null);
                    }}
                  />
                ) : (
                  <>
                    <div className="insight-card">
                      <div className="panel-title">
                        <span className="icon-tile amber">
                          <AlertTriangle size={18} />
                        </span>
                        <h3>Needs your attention</h3>
                        <span className="count-badge">
                          {ctx.evaluation.conflicts.length}
                        </span>
                      </div>
                      <p className="panel-description">
                        Overlaps are okay. Competing requirements need a closer
                        look.
                      </p>
                      <ConflictList
                        evaluation={ctx.evaluation}
                        onEvent={(id) => {
                          const e = events.find((x) => x.id === id);
                          if (e) open(e);
                        }}
                      />
                      <button
                        className="button wide"
                        onClick={() => resolve()}
                        disabled={busy}
                      >
                        <WandSparkles size={15} /> Resolve conflicts{" "}
                        <ArrowRight size={15} />
                      </button>
                    </div>
                    <div className="insight-card readiness-card">
                      <div className="panel-title">
                        <ShieldCheck size={19} />
                        <h3>Training outlook</h3>
                      </div>
                      {coverage.map((c) => (
                        <div className="coverage-mini" key={c.requirement.id}>
                          <div>
                            <span>{c.requirement.name}</span>
                            <strong>
                              {c.completed + c.planned}/{c.requirement.count}
                            </strong>
                          </div>
                          <div className="progress-track">
                            <span
                              style={{
                                width:
                                  Math.min(
                                    100,
                                    ((c.completed + c.planned) /
                                      c.requirement.count) *
                                      100,
                                  ) + "%",
                              }}
                            />
                          </div>
                          <small>
                            {c.completed} completed · {c.planned} planned
                            {c.gap > 0 ? ` · ${c.gap} missing` : ""}
                          </small>
                        </div>
                      ))}
                      <button
                        className="text-button"
                        onClick={() => setPage("training")}
                      >
                        View training readiness <ArrowUpRight size={14} />
                      </button>
                    </div>
                    <div className="agent-note">
                      <WandSparkles size={18} />
                      <div>
                        <strong>
                          {ctx.ai_available
                            ? "Your departments, in conversation."
                            : "Start with a workable plan."}
                        </strong>
                        <p>
                          {ctx.ai_available
                            ? "Agents negotiate alternatives. Every result is checked against your constraints."
                            : "The scheduling engine is ready. Live AI negotiation is optional and currently unconfigured."}
                        </p>
                      </div>
                    </div>
                    {ctx.runs.length > 0 && (
                      <div className="insight-card">
                        <h3>Saved proposals</h3>
                        {ctx.runs
                          .slice(-5)
                          .reverse()
                          .map((r) => (
                            <button
                              className="saved-run"
                              key={r.id}
                              onClick={async () => {
                                try {
                                  setProposal(
                                    await api<Proposal>("/api/runs/" + r.id),
                                  );
                                  setPollingPaused(false);
                                  setOverlayHidden(false);
                                  setViewRun(true);
                                } catch (e) {
                                  setError((e as Error).message);
                                }
                              }}
                            >
                              <span>
                                {r.mode === "live-agents"
                                  ? "Department negotiation"
                                  : "Schedule proposal"}
                                <small>{r.status.replaceAll("_", " ")}</small>
                              </span>
                              <ArrowUpRight size={14} />
                            </button>
                          ))}
                      </div>
                    )}
                  </>
                )}
              </aside>
            </div>
          ) : page === "training" ? (
            <section className="table-card">
              <div className="section-heading">
                <div>
                  <h2>Training obligations</h2>
                  <p>
                    Illustrative policies · completed and projected coverage are
                    separate.
                  </p>
                </div>
                <button className="button" onClick={() => setPage("data")}>
                  <Settings2 size={15} /> Manage requirements
                </button>
              </div>
              {coverage.length === 0 && (
                <p className="empty">No requirements overlap this week.</p>
              )}
              <div className="readiness-grid">
                {coverage.map((c) => (
                  <article key={c.requirement.id}>
                    <span
                      className="eyebrow"
                      style={{
                        color: dep(data, c.requirement.department_id)?.color,
                      }}
                    >
                      {dep(data, c.requirement.department_id)?.name}
                    </span>
                    <h3>{c.requirement.name}</h3>
                    <span
                      className={
                        "pill " + (c.gap ? "amber-pill" : "green-pill")
                      }
                    >
                      {c.gap
                        ? `${c.gap} occurrence${c.gap > 1 ? "s" : ""} to schedule`
                        : "Projected coverage met"}
                    </span>
                    <div className="coverage-numbers">
                      <div>
                        <strong>{c.completed}</strong>
                        <small>COMPLETED</small>
                      </div>
                      <div>
                        <strong>{c.planned}</strong>
                        <small>PLANNED</small>
                      </div>
                      <div>
                        <strong>{c.requirement.count}</strong>
                        <small>REQUIRED</small>
                      </div>
                    </div>
                    <p>Due {stamp(c.requirement.end)}</p>
                    <small>{c.requirement.source}</small>
                    {c.gap > 0 && (
                      <button
                        className="text-button"
                        onClick={() => {
                          setPage("calendar");
                          resolve();
                        }}
                      >
                        Find a place in the week <ArrowRight size={14} />
                      </button>
                    )}
                  </article>
                ))}
              </div>
            </section>
          ) : page === "documents" ? (
            <Documents
              data={data}
              week={day}
              toISO={iso}
              toLocal={localInput}
              ai={ctx.ai_available}
              onCreated={load}
            />
          ) : (
            <PlanningData
              data={data}
              week={ctx.week}
              ai={ctx.ai_available}
              reload={load}
              setError={setError}
              openTemplate={(e) => open(e, true)}
              scheduleTemplate={(t) => open(occurrence(t, data, ctx.week))}
            />
          )}
          <footer className="workspace-footer">
            <span>
              <Anchor size={13} /> PB4T · A clearer plan. A readier crew.
            </span>
            <span>Local data · Illustrative policies · {ctx.timezone}</span>
          </footer>
        </div>
      </main>
      {choosingTemplate && (
        <div className="modal-shade">
          <section className="event-editor">
            <div className="editor-header">
              <div>
                <span className="eyebrow">NEW EVENT</span>
                <h2>Choose an event template</h2>
              </div>
              <button
                className="icon-button"
                aria-label="Close template picker"
                onClick={() => setChoosingTemplate(false)}
              >
                <X size={20} />
              </button>
            </div>
            <div className="editor-content">
              <p className="panel-description">
                Choose the activity, then set when it should happen. Department,
                contact, duration, tags, and constraints are filled in for you.
              </p>
              {data.templates.length === 0 && (
                <p className="empty">
                  Create a reusable template to schedule your first event.
                </p>
              )}
              {data.templates.map((t) => (
                <button
                  className="template-choice"
                  key={t.id}
                  onClick={() => {
                    setChoosingTemplate(false);
                    open(occurrence(t, data, ctx.week));
                  }}
                >
                  <span
                    className="dept-dot"
                    style={{ background: dep(data, t.department_id)?.color }}
                  />
                  <span>
                    <strong>{t.name}</strong>
                    <small>
                      {dep(data, t.department_id)?.name} · {t.duration} min ·{" "}
                      {t.constraint_ids.length} constraints
                    </small>
                    <small>{t.contact || "No default contact"}</small>
                  </span>
                  <ArrowRight size={17} />
                </button>
              ))}
            </div>
            <div className="editor-footer">
              <button
                className="button"
                onClick={() => {
                  setChoosingTemplate(false);
                  open(newEvent(data, ctx.week), true);
                }}
              >
                <Plus size={15} /> Create template
              </button>
            </div>
          </section>
        </div>
      )}
      {editing && (
        <EventEditor
          key={editing.id + editing.start}
          event={editing}
          data={data}
          week={day}
          template={template}
          ai={ctx.ai_available}
          close={() => setEditing(null)}
          resolve={resolve}
          saveAsTemplate={(e) =>
            open({ ...e, id: uid(), template_id: undefined }, true)
          }
          saved={async () => {
            setEditing(null);
            setTemplate(false);
            await load();
            setToast(template ? "Template saved" : "Event saved");
          }}
        />
      )}
      {toast && (
        <div className="toast">
          <Check size={18} />
          {toast}
        </div>
      )}
    </div>
  );
}
function layout(events: Pick<Event, "id" | "start" | "duration">[]) {
  const result = new Map<string, { lane: number; total: number }>();
  let cluster: typeof events = [];
  let end = -1;
  function flush() {
    const lanes: number[] = [];
    const assigned: { e: (typeof events)[number]; lane: number }[] = [];
    for (const e of cluster) {
      let lane = lanes.findIndex((t) => t <= minutes(e.start));
      if (lane < 0) lane = lanes.length;
      lanes[lane] = minutes(e.start) + e.duration;
      assigned.push({ e, lane });
    }
    for (const { e, lane } of assigned)
      result.set(e.id, { lane, total: lanes.length });
  }
  for (const e of events) {
    if (minutes(e.start) >= end) {
      flush();
      cluster = [];
      end = -1;
    }
    cluster.push(e);
    end = Math.max(end, minutes(e.start) + e.duration);
  }
  flush();
  return result;
}
function ConflictList({
  evaluation,
  onEvent,
}: {
  evaluation: Evaluation;
  onEvent?: (id: string) => void;
}) {
  return (
    <div className="conflict-list">
      {evaluation.conflicts.length === 0 ? (
        <div className="all-clear">
          <CheckCheck size={23} />
          <strong>No hard conflicts</strong>
          <span>All supplied constraints are satisfied.</span>
        </div>
      ) : (
        evaluation.conflicts.map((c, i) => (
          <div className="conflict" key={i}>
            <div>
              <span className="conflict-type">
                {c.type.replaceAll("_", " ")}
              </span>
              <small>
                {c.start && !c.start.startsWith("0001") ? stamp(c.start) : ""}
              </small>
            </div>
            <p>{c.message}</p>
            {onEvent && c.event_ids?.[0] && (
              <button
                className="text-button"
                onClick={() => onEvent(c.event_ids[0])}
              >
                Review event <ArrowUpRight size={12} />
              </button>
            )}
          </div>
        ))
      )}
    </div>
  );
}
function ProposalPanel({
  proposal: p,
  data,
  busy,
  apply,
  cancel,
  pollingPaused,
  retry,
  comparisonActive,
  highlightedEventId,
  onHighlight,
}: {
  proposal: Proposal;
  data: Data;
  busy: boolean;
  apply: () => void;
  cancel: () => void;
  pollingPaused: boolean;
  retry: () => void;
  comparisonActive: boolean;
  highlightedEventId: string | null;
  onHighlight: (id: string | null) => void;
}) {
  const [hovered, setHovered] = useState<string | null>(null),
    [focused, setFocused] = useState<string | null>(null);
  useEffect(() => {
    onHighlight(comparisonActive ? hovered || focused : null);
    return () => onHighlight(null);
  }, [comparisonActive, hovered, focused, onHighlight]);
  return (
    <div className="insight-card proposal-panel">
      <div className="panel-title">
        <WandSparkles size={19} />
        <h3>Proposed weekly plan</h3>
        <button
          className="icon-button"
          onClick={cancel}
          aria-label="Close proposal"
        >
          <X size={15} />
        </button>
      </div>
      <span className="pill">
        {p.mode === "live-agents"
          ? "Live department negotiation"
          : "Deterministic scheduling"}
      </span>
      {p.status === "running" ? (
        <div className="run-progress">
          {pollingPaused ? (
            <>
              <AlertTriangle />
              <strong>Status checks paused</strong>
              <p>
                The run status could not be fetched. Check that the local app is
                running, then retry.
              </p>
              <button className="button" onClick={retry}>
                Retry status check
              </button>
            </>
          ) : (
            <>
              <LoaderCircle className="spin" />
              <strong>{p.progress}</strong>
              <p>Saved events stay in place while we work.</p>
            </>
          )}
        </div>
      ) : (
        <>
          <p className={p.error ? "error" : ""}>{p.error || p.summary}</p>
          {p.alternate && (
            <div className="notice">
              Alternative: changes your requested event time.
            </div>
          )}
          {(p.changes || []).map((c) => {
            const event =
              p.events?.find((e) => e.id === c.event_id) ||
              data.events.find((e) => e.id === c.event_id);
            return (
              <div
                className={
                  "change" +
                  (highlightedEventId === c.event_id ? " change-active" : "")
                }
                key={c.event_id}
                role="group"
                aria-label={`Proposed change: ${c.name}`}
                tabIndex={comparisonActive ? 0 : undefined}
                onMouseEnter={() => setHovered(c.event_id)}
                onMouseLeave={() => setHovered(null)}
                onFocus={() => setFocused(c.event_id)}
                onBlur={() => setFocused(null)}
                style={
                  {
                    "--event-color":
                      dep(data, event?.department_id || "")?.color || "#607e74",
                  } as React.CSSProperties
                }
              >
                <strong>{c.name}</strong>
                <div>
                  {c.before ? stamp(c.before) : "New occurrence"}{" "}
                  <ArrowRight size={12} /> {stamp(c.after)}
                </div>
                <small>{c.reason}</small>
              </div>
            );
          })}
          {p.evaluation && (
            <>
              <h4>Personnel assignments</h4>
              {(p.evaluation.assignments || [])
                .filter(
                  (a) =>
                    !p.previous_assignments?.some(
                      (old) =>
                        old.event_id === a.event_id &&
                        old.person_id === a.person_id &&
                        old.role === a.role,
                    ),
                )
                .map((a, i) => (
                  <div className="assignment" key={i}>
                    <Users size={13} />
                    <span>
                      {personName(data, a.person_id)}
                      <small>
                        {p.events?.find((e) => e.id === a.event_id)?.name} ·{" "}
                        {data.qualifications.find((q) => q.id === a.role)
                          ?.name || a.role}
                      </small>
                    </span>
                  </div>
                ))}
              {p.evaluation.warnings?.length > 0 && (
                <details>
                  <summary>
                    {p.evaluation.warnings.length} requested-time tradeoffs
                  </summary>
                  {p.evaluation.warnings.map((x, i) => (
                    <p key={i}>{x.message}</p>
                  ))}
                </details>
              )}
              <ConflictList evaluation={p.evaluation} />
              <h4>Projected training</h4>
              {p.evaluation.coverage?.map((c) => (
                <p className="proposal-coverage" key={c.requirement.id}>
                  {c.requirement.name}
                  <strong>
                    {c.completed} done · {c.planned} planned · {c.gap} gap
                  </strong>
                </p>
              ))}
            </>
          )}
          {p.unscheduled?.length > 0 && (
            <div className="notice error">
              <div>
                <strong>Still needs a place</strong>
                {p.unscheduled.map((e) => (
                  <p key={e.id}>{e.name}</p>
                ))}
              </div>
            </div>
          )}
          {p.negotiation?.map((n, i) => (
            <details key={i}>
              <summary>
                {dep(data, n.department_id)?.name} · round {n.round}
              </summary>
              <p>{n.summary}</p>
              {n.preferences.map((x, j) => (
                <p key={j}>
                  {p.events?.find((e) => e.id === x.event_id)?.name ||
                    x.event_id}
                  : {stamp(x.start)}. {x.reason}
                </p>
              ))}
              <small>
                {n.evaluation.feasible
                  ? "Candidate passed hard checks"
                  : `${n.evaluation.conflicts.length} candidate conflicts returned to negotiation`}
              </small>
            </details>
          ))}
          <div className="proposal-actions">
            {p.status === "feasible" && (
              <button
                className="button primary wide"
                disabled={busy}
                onClick={apply}
              >
                <Check size={15} /> Apply plan
              </button>
            )}
            {p.status === "applied" && (
              <span className="pill green-pill">
                <Check size={13} /> Applied to calendar
              </span>
            )}
            <button className="button wide" onClick={cancel}>
              {p.status === "applied" ? "Close" : "Cancel proposal"}
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function EventEditor({
  event,
  data,
  week,
  template,
  ai,
  close,
  resolve,
  saved,
  saveAsTemplate,
}: {
  event: Event;
  data: Data;
  week: string;
  template: boolean;
  ai: boolean;
  close: () => void;
  resolve: (event?: Event) => void;
  saved: () => Promise<void>;
  saveAsTemplate: (e: Event) => void;
}) {
  const [e, setE] = useState(event),
    [ev, setEv] = useState<Evaluation | null>(null),
    [error, setError] = useState(""),
    [working, setWorking] = useState(false),
    [newRules, setNewRules] = useState<ConstraintDraft[]>([]),
    [text, setText] = useState(""),
    [rules, setRules] = useState(data.constraints);
  const change = (key: keyof Event, value: unknown) => {
    setE({ ...e, [key]: value });
    setEv(null);
  };
  useEffect(() => {
    if (template || !e.name) return;
    let active = true;
    const timer = setTimeout(() => {
      api<Evaluation>("/api/validate", "POST", { week, event: e })
        .then((result) => {
          if (active) {
            setEv(result);
            setError("");
          }
        })
        .catch((err) => {
          if (active) setError(err.message);
        });
    }, 300);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [e, week, template]);
  async function validate() {
    setWorking(true);
    setError("");
    try {
      const result = await api<Evaluation>("/api/validate", "POST", {
        week,
        event: e,
      });
      setEv(result);
      return result;
    } catch (err) {
      setError((err as Error).message);
      return null;
    } finally {
      setWorking(false);
    }
  }
  async function save(draft: boolean) {
    setWorking(true);
    setError("");
    try {
      const value = {
        ...e,
        status: template ? "draft" : draft ? "draft" : "scheduled",
      };
      await api(
        "/api/data/" + (template ? "templates" : "events") + "/" + e.id,
        "PUT",
        template ? templateDefaults(value) : value,
      );
      await saved();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setWorking(false);
    }
  }
  async function saveRule(draft: ConstraintDraft) {
    setWorking(true);
    setError("");
    try {
      const r = { ...draft.rule, confirmed: draft.rule.type !== "note" };
      await api("/api/data/constraints/" + r.id, "PUT", r);
      setRules((current) => [...current.filter((x) => x.id !== r.id), r]);
      setE((current) => ({
        ...current,
        constraint_ids: [...new Set([...current.constraint_ids, r.id])],
      }));
      setEv(null);
      setNewRules((current) => current.filter((x) => x.rule.id !== r.id));
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setWorking(false);
    }
  }
  const persisted = data.events.some((x) => x.id === e.id);
  return (
    <div
      className="modal-shade"
      onMouseDown={(x) => {
        if (x.target === x.currentTarget) close();
      }}
    >
      <section className="event-editor">
        <div className="editor-header">
          <div>
            <span className="eyebrow">
              {template
                ? "EVENT TEMPLATE"
                : persisted
                  ? "EVENT DETAILS"
                  : "NEW EVENT"}
            </span>
            <h2>{e.name || "Make room for readiness."}</h2>
          </div>
          <button
            className="icon-button"
            onClick={close}
            aria-label="Close event editor"
          >
            <X size={20} />
          </button>
        </div>
        <div className="editor-content">
          {error && <div className="notice error">{error}</div>}
          {e.source_ref && <SourceLink source={e.source_ref} />}
          <label>
            {template ? "Template name" : "Event name"}
            <input
              value={e.name}
              onChange={(x) => change("name", x.currentTarget.value)}
              placeholder="e.g. Engineering casualty control"
            />
          </label>
          {!template && e.template_id && (
            <div className="template-inheritance">
              <span className="eyebrow">FROM TEMPLATE</span>
              <strong>
                {data.templates.find((t) => t.id === e.template_id)?.name ||
                  e.name}
              </strong>
              <p>
                {dep(data, e.department_id)?.name} · {e.duration} min ·{" "}
                {e.contact || "No default contact"} · {e.constraint_ids.length}{" "}
                constraints
              </p>
            </div>
          )}
          {!template && (
            <>
              <div className="form-row">
                <label>
                  Preferred start time
                  <input
                    type="datetime-local"
                    step={900}
                    value={localInput(e.start)}
                    onInput={(x) => {
                      const start = iso(x.currentTarget.value);
                      setE({ ...e, start, requested_start: start });
                      setEv(null);
                    }}
                  />
                </label>
              </div>
              <div className="form-row">
                <label>
                  Earliest allowed start
                  <input
                    type="datetime-local"
                    value={localInput(e.window_start)}
                    onInput={(x) =>
                      change("window_start", iso(x.currentTarget.value))
                    }
                  />
                </label>
                <label>
                  Latest allowed end
                  <input
                    type="datetime-local"
                    value={localInput(e.window_end)}
                    onInput={(x) =>
                      change("window_end", iso(x.currentTarget.value))
                    }
                  />
                </label>
              </div>
            </>
          )}
          <details
            className="template-overrides"
            open={template || !e.template_id}
          >
            <summary>
              {template
                ? "Reusable event defaults"
                : "Event details and overrides"}
            </summary>
            <label>
              Description
              <textarea
                value={e.description || ""}
                onChange={(x) => change("description", x.currentTarget.value)}
              />
            </label>
            <div className="form-row">
              <label>
                Department
                <select
                  value={e.department_id}
                  onChange={(x) =>
                    change("department_id", x.currentTarget.value)
                  }
                >
                  {data.departments.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Point of contact
                <input
                  value={e.contact}
                  onChange={(x) => change("contact", x.currentTarget.value)}
                  placeholder="Name or role"
                />
              </label>
            </div>
            <label>
              Duration (minutes)
              <input
                type="number"
                min={15}
                step={15}
                value={e.duration}
                onChange={(x) =>
                  change("duration", Number(x.currentTarget.value))
                }
              />
            </label>{" "}
            <div className="form-row">
              <label>
                Drill type
                <input
                  value={e.type}
                  onChange={(x) => change("type", x.currentTarget.value)}
                />
              </label>
              <label>
                Tags (comma separated)
                <input
                  value={e.tags.join(", ")}
                  onChange={(x) =>
                    change(
                      "tags",
                      x.currentTarget.value
                        .split(",")
                        .map((s) => s.trim())
                        .filter(Boolean),
                    )
                  }
                />
              </label>
            </div>
            <div className="checkbox-row">
              {!template && (
                <label>
                  <input
                    type="checkbox"
                    checked={e.fixed}
                    onChange={(x) => change("fixed", x.target.checked)}
                  />{" "}
                  Fixed operational commitment
                </label>
              )}
              <label>
                <input
                  type="checkbox"
                  checked={e.fatigue}
                  onChange={(x) => change("fatigue", x.target.checked)}
                />{" "}
                Counts toward fatigue
              </label>
            </div>
            <div className="constraint-section">
              <div className="section-heading">
                <div>
                  <h3>Event constraints</h3>
                  <p>
                    Attach the people, resources, and rules this event needs.
                  </p>
                </div>
                <span className="pill">{e.constraint_ids.length} attached</span>
              </div>
              {e.constraint_ids.map((id) => (
                <div className="attached-rule" key={id}>
                  <ShieldCheck size={15} />
                  <span>{rules.find((r) => r.id === id)?.name || id}</span>
                  <button
                    className="icon-button"
                    aria-label="Remove constraint"
                    onClick={() =>
                      change(
                        "constraint_ids",
                        e.constraint_ids.filter((x) => x !== id),
                      )
                    }
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
              <div className="form-row">
                <select
                  aria-label="Attach existing constraint"
                  value=""
                  onChange={(x) =>
                    change("constraint_ids", [
                      ...e.constraint_ids,
                      x.currentTarget.value,
                    ])
                  }
                >
                  <option value="">Select an existing constraint…</option>
                  {rules
                    .filter(
                      (r) =>
                        r.scope !== "global" &&
                        !e.constraint_ids.includes(r.id),
                    )
                    .map((r) => (
                      <option value={r.id} key={r.id}>
                        {r.name}
                      </option>
                    ))}
                </select>
                <button
                  className="button"
                  onClick={() =>
                    setNewRules([
                      {
                        rule: {
                          id: uid(),
                          name: "",
                          type: "qualification",
                          scope: "event",
                          confirmed: true,
                          count: 1,
                          qualification_id: data.qualifications[0]?.id,
                        },
                        requested_qualification: "",
                        explanation: "",
                      },
                    ])
                  }
                >
                  <Plus size={14} /> Create new
                </button>
              </div>
              <details>
                <summary>Shared policies also apply</summary>
                {rules
                  .filter((r) => r.scope === "global")
                  .map((r) => (
                    <p key={r.id}>{r.name}</p>
                  ))}
              </details>
            </div>
            <div className="assist-box">
              <div>
                <WandSparkles size={16} />
                <strong>Describe a constraint</strong>
                <span className="pill">
                  {ai ? "Live AI" : "AI not configured"}
                </span>
              </div>
              <textarea
                disabled={!ai}
                value={text}
                onChange={(x) => setText(x.currentTarget.value)}
                placeholder="e.g. This drill cannot overlap flight quarters"
              />
              <button
                className="text-button"
                disabled={!ai || !text || working}
                onClick={async () => {
                  setWorking(true);
                  setError("");
                  try {
                    const drafts = await api<ConstraintDraft[]>(
                      "/api/constraints/interpret",
                      "POST",
                      {
                        text,
                        event: e,
                      },
                    );
                    if (!Array.isArray(drafts))
                      throw new Error(
                        "Restart the app to load the updated constraint interpreter.",
                      );
                    setNewRules(drafts);
                  } catch (err) {
                    setError((err as Error).message);
                  } finally {
                    setWorking(false);
                  }
                }}
              >
                Interpret for review <ArrowRight size={13} />
              </button>
            </div>
            {newRules.map((draft) => (
              <div className="rule-draft" key={draft.rule.id}>
                <h3>Review structured constraint</h3>
                <p>
                  Confirm this interpretation before attaching it. Notes remain
                  unenforced.
                </p>
                {draft.explanation && (
                  <p className="notice">{draft.explanation}</p>
                )}
                <RuleFields
                  rule={draft.rule}
                  setRule={(rule) =>
                    setNewRules((current) =>
                      current.map((x) =>
                        x.rule.id === draft.rule.id ? { ...x, rule } : x,
                      ),
                    )
                  }
                  data={data}
                  toLocal={localInput}
                  toISO={iso}
                />
                <div className="form-row">
                  <button
                    className="button primary"
                    disabled={working}
                    onClick={() => saveRule(draft)}
                  >
                    Confirm & attach
                  </button>
                  <button
                    className="button"
                    onClick={() =>
                      setNewRules((current) =>
                        current.filter((x) => x.rule.id !== draft.rule.id),
                      )
                    }
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ))}
          </details>
          {ev && (
            <div className="validation-result">
              <h3>
                {ev.feasible
                  ? "Placement is feasible"
                  : "Placement needs attention"}
              </h3>
              <ConflictList evaluation={ev} />
              {ev.assignments
                ?.filter((a) => a.event_id === e.id)
                .map((a, i) => (
                  <p key={i}>
                    <Users size={13} /> {personName(data, a.person_id)}
                  </p>
                ))}
            </div>
          )}
          {!template && (
            <button className="text-button" onClick={() => saveAsTemplate(e)}>
              <Plus size={14} /> Save as new template
            </button>
          )}
          {persisted && !template && (
            <div className="event-record-actions">
              <button
                className="text-button"
                onClick={async () => {
                  try {
                    await api("/api/events/" + e.id + "/complete", "POST");
                    await saved();
                  } catch (err) {
                    setError((err as Error).message);
                  }
                }}
                disabled={e.status === "completed"}
              >
                <CheckCheck size={14} /> Record completion
              </button>
              <button
                className="text-button danger"
                onClick={async () => {
                  if (!confirm("Delete this event?")) return;
                  try {
                    await api("/api/data/events/" + e.id, "DELETE");
                    await saved();
                  } catch (err) {
                    setError((err as Error).message);
                  }
                }}
              >
                <Trash2 size={14} /> Delete event
              </button>
            </div>
          )}
        </div>
        <div className="editor-footer">
          {template ? (
            <button
              className="button primary"
              onClick={() => save(true)}
              disabled={working}
            >
              Save template
            </button>
          ) : (
            <>
              <button
                className="button"
                onClick={validate}
                disabled={working || !e.name}
              >
                <ShieldCheck size={15} /> Check placement
              </button>
              {ev?.feasible ? (
                <button
                  className="button primary"
                  onClick={() => save(false)}
                  disabled={working}
                >
                  Save event
                </button>
              ) : (
                <>
                  <button
                    className="button"
                    onClick={() => save(true)}
                    disabled={working || !e.name}
                  >
                    Save as draft
                  </button>
                  <button
                    className="button primary"
                    onClick={() => resolve(e)}
                    disabled={working || !e.name}
                  >
                    <WandSparkles size={15} /> Resolve
                  </button>
                </>
              )}
            </>
          )}
        </div>
      </section>
    </div>
  );
}

const collections = [
  ["people", "Personnel roster"],
  ["qualifications", "Qualifications"],
  ["duties", "Duty commitments"],
  ["assets", "Shared assets"],
  ["constraints", "Constraints & policies"],
  ["templates", "Event templates"],
  ["requirements", "Training requirements"],
  ["completions", "Completion history"],
  ["departments", "Departments"],
];
// Dynamic forms share only presentation; the backend retains typed domain validation.
type FormRecord = Record<string, any>;
function PlanningData({
  data,
  week,
  ai,
  reload,
  setError,
  openTemplate,
  scheduleTemplate,
}: {
  data: Data;
  week: string;
  ai: boolean;
  reload: () => Promise<void>;
  setError: (s: string) => void;
  openTemplate: (e: Event) => void;
  scheduleTemplate: (t: EventTemplate) => void;
}) {
  const [kind, setKind] = useState("people"),
    [record, setRecord] = useState<FormRecord | null>(null),
    [saving, setSaving] = useState(false);
  const defaults = (): FormRecord => {
    const start = iso(dateOnly(week) + "T00:00"),
      end = iso(addDays(dateOnly(week), 7) + "T00:00");
    return {
      id: uid(),
      name: "",
      department_id: data.departments[0]?.id || "",
      qualifications: [],
      availability: [{ start, end }],
      capacity: 1,
      person_id: data.people[0]?.id || "",
      fatigue: true,
      start,
      end,
      type: kind === "constraints" ? "qualification" : "drill",
      count: 1,
      source: "User supplied requirement",
      repeat_days: 7,
      at: start,
      event_id: "",
      scope: "event",
      confirmed: true,
      qualification_id: data.qualifications[0]?.id,
      color: "#5479ab",
      goal: "",
    };
  };
  const list = (data as unknown as Record<string, FormRecord[]>)[kind] || [];
  async function save() {
    if (!record) return;
    setSaving(true);
    try {
      await api("/api/data/" + kind + "/" + record.id, "PUT", record);
      setRecord(null);
      await reload();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSaving(false);
    }
  }
  const set = (key: string, value: unknown) =>
    setRecord({ ...record, [key]: value });
  const select = (
    key: string,
    label: string,
    items: { id: string; name: string }[],
  ) => (
    <label>
      {label}
      <select
        value={record?.[key] || ""}
        onChange={(e) => set(key, e.currentTarget.value)}
      >
        <option value="">Select…</option>
        {items.map((x) => (
          <option key={x.id} value={x.id}>
            {x.name}
          </option>
        ))}
      </select>
    </label>
  );
  const input = (key: string, label: string, type = "text") => (
    <label>
      {label}
      <input
        type={type}
        value={
          type === "datetime-local"
            ? localInput(record?.[key] || "")
            : (record?.[key] ?? "")
        }
        onInput={(e) =>
          set(
            key,
            type === "datetime-local"
              ? iso(e.currentTarget.value)
              : type === "number"
                ? Number(e.currentTarget.value)
                : e.currentTarget.value,
          )
        }
      />
    </label>
  );
  return (
    <section className="table-card">
      <div className="data-tabs">
        {collections.map(([id, label]) => (
          <button
            key={id}
            className={id === kind ? "selected" : ""}
            onClick={() => {
              setKind(id);
              setRecord(null);
            }}
          >
            {label}
          </button>
        ))}
      </div>
      <div className="section-heading">
        <div>
          <h2>{collections.find((x) => x[0] === kind)?.[1]}</h2>
          <p>
            Shared by the calendar, scheduling engine
            {ai ? ", and departmental agents" : ""}.
          </p>
        </div>
        <button
          className="button primary"
          onClick={() =>
            kind === "templates"
              ? openTemplate(newEvent(data, week))
              : setRecord(defaults())
          }
        >
          <Plus size={15} /> Add {kind === "people" ? "person" : "record"}
        </button>
      </div>
      <div className="data-list">
        {!list.length && (
          <div className="empty">
            No records yet. Add one to start planning.
          </div>
        )}
        {list.map((r) => (
          <div className="data-row" key={r.id}>
            <span className="data-icon">
              {kind === "people" ? (
                <Users size={18} />
              ) : kind === "assets" ? (
                <Layers size={18} />
              ) : (
                <ShieldCheck size={18} />
              )}
            </span>
            <div>
              <strong>{r.name || r.type}</strong>
              {r.source_ref && <SourceLink source={r.source_ref} compact />}
              <small>
                {r.department_id ? dep(data, r.department_id)?.name : ""}
                {kind === "people"
                  ? " · " +
                    r.qualifications
                      .map(
                        (id: string) =>
                          data.qualifications.find((q) => q.id === id)?.name ||
                          id,
                      )
                      .join(", ")
                  : kind === "assets"
                    ? `Capacity ${r.capacity}`
                    : kind === "constraints"
                      ? `${r.type} · ${r.scope === "global" ? "Entire plan" : "Attached events"}${!r.confirmed ? " · UNENFORCED" : ""}`
                      : kind === "requirements"
                        ? `${r.count} required · ${r.source}`
                        : kind === "duties"
                          ? `${personName(data, r.person_id)} · ${stamp(r.start)} – ${stamp(r.end)}`
                          : kind === "completions"
                            ? ` · ${stamp(r.at)}`
                            : ""}
              </small>
            </div>
            {kind === "templates" && (
              <button
                className="button"
                onClick={() => scheduleTemplate(r as EventTemplate)}
              >
                Schedule event <Plus size={14} />
              </button>
            )}
            <button
              className="text-button"
              onClick={() =>
                kind === "templates"
                  ? openTemplate({
                      ...occurrence(r as EventTemplate, data, week),
                      id: r.id,
                      template_id: undefined,
                    })
                  : setRecord(structuredClone(r))
              }
            >
              Edit <ArrowUpRight size={14} />
            </button>
            <button
              className="icon-button danger"
              aria-label="Delete record"
              onClick={async () => {
                if (
                  !confirm(
                    "Delete this record? Attached records may need to be updated first.",
                  )
                )
                  return;
                try {
                  await api("/api/data/" + kind + "/" + r.id, "DELETE");
                  await reload();
                } catch (e) {
                  setError((e as Error).message);
                }
              }}
            >
              <Trash2 size={15} />
            </button>
          </div>
        ))}
      </div>
      {record && (
        <div className="modal-shade">
          <section className="data-editor">
            <div className="editor-header">
              <h2>
                {list.some((x) => x.id === record.id) ? "Edit" : "Add"}{" "}
                {collections.find((x) => x[0] === kind)?.[1]}
              </h2>
              <button
                className="icon-button"
                onClick={() => setRecord(null)}
                aria-label="Close planning input editor"
              >
                <X size={20} />
              </button>
            </div>
            <div className="editor-content">
              {record.source_ref && <SourceLink source={record.source_ref} />}
              {kind === "constraints" ? (
                <RuleFields
                  rule={record as Rule}
                  setRule={setRecord}
                  data={data}
                  toLocal={localInput}
                  toISO={iso}
                />
              ) : (
                <>
                  {kind !== "completions" && input("name", "Name")}
                  {["people", "requirements", "completions"].includes(kind) &&
                    select("department_id", "Department", data.departments)}
                  {kind === "departments" && (
                    <>
                      {input("color", "Calendar color", "color")}
                      <label>
                        Agent goals
                        <textarea
                          value={record.goal}
                          onChange={(e) => set("goal", e.currentTarget.value)}
                        />
                      </label>
                    </>
                  )}
                  {kind === "people" && (
                    <fieldset>
                      <legend>Qualifications</legend>
                      {data.qualifications.map((q) => (
                        <label className="inline-check" key={q.id}>
                          <input
                            type="checkbox"
                            checked={(record.qualifications || []).includes(
                              q.id,
                            )}
                            onChange={(e) =>
                              set(
                                "qualifications",
                                e.target.checked
                                  ? [...record.qualifications, q.id]
                                  : record.qualifications.filter(
                                      (x: string) => x !== q.id,
                                    ),
                              )
                            }
                          />
                          {q.name}
                        </label>
                      ))}
                    </fieldset>
                  )}
                  {["people", "assets"].includes(kind) && (
                    <fieldset>
                      <legend>Availability windows</legend>
                      {(record.availability || []).map(
                        (w: Window, i: number) => (
                          <div key={i} className="availability-window">
                            <input
                              type="datetime-local"
                              aria-label="Availability start"
                              value={localInput(w.start)}
                              onInput={(e) =>
                                set(
                                  "availability",
                                  record.availability.map(
                                    (x: Window, j: number) =>
                                      j === i
                                        ? {
                                            ...x,
                                            start: iso(e.currentTarget.value),
                                          }
                                        : x,
                                  ),
                                )
                              }
                            />
                            <input
                              type="datetime-local"
                              aria-label="Availability end"
                              value={localInput(w.end)}
                              onInput={(e) =>
                                set(
                                  "availability",
                                  record.availability.map(
                                    (x: Window, j: number) =>
                                      j === i
                                        ? {
                                            ...x,
                                            end: iso(e.currentTarget.value),
                                          }
                                        : x,
                                  ),
                                )
                              }
                            />
                            <button
                              className="icon-button"
                              aria-label="Remove availability window"
                              onClick={() =>
                                set(
                                  "availability",
                                  record.availability.filter(
                                    (_: Window, j: number) => j !== i,
                                  ),
                                )
                              }
                            >
                              <X size={14} />
                            </button>
                          </div>
                        ),
                      )}
                      <button
                        className="text-button"
                        onClick={() =>
                          set("availability", [
                            ...(record.availability || []),
                            defaults().availability[0],
                          ])
                        }
                      >
                        <Plus size={14} /> Add window
                      </button>
                    </fieldset>
                  )}
                  {kind === "assets" && input("capacity", "Capacity", "number")}
                  {kind === "duties" && (
                    <>
                      {select("person_id", "Person", data.people)}
                      {input("start", "Start", "datetime-local")}
                      {input("end", "End", "datetime-local")}
                      <label className="inline-check">
                        <input
                          type="checkbox"
                          checked={record.fatigue}
                          onChange={(e) => set("fatigue", e.target.checked)}
                        />{" "}
                        Counts toward fatigue
                      </label>
                    </>
                  )}
                  {["requirements", "completions"].includes(kind) &&
                    input("type", "Drill type")}
                  {kind === "requirements" && (
                    <>
                      {input(
                        "start",
                        "Coverage window start",
                        "datetime-local",
                      )}
                      {input("end", "Coverage window end", "datetime-local")}
                      {input("count", "Required occurrences", "number")}
                      {input(
                        "repeat_days",
                        "Repeat every (days; 0 for no repeat)",
                        "number",
                      )}
                      {input("source", "Policy source / reference")}
                    </>
                  )}
                  {kind === "completions" && (
                    <>
                      {input("at", "Completed at", "datetime-local")}
                      <label>
                        Linked event (optional)
                        <select
                          value={record.event_id || ""}
                          onChange={(e) =>
                            set("event_id", e.currentTarget.value)
                          }
                        >
                          <option value="">No linked event</option>
                          {data.events.map((e) => (
                            <option key={e.id} value={e.id}>
                              {e.name}
                            </option>
                          ))}
                        </select>
                      </label>
                    </>
                  )}
                </>
              )}
            </div>
            <div className="editor-footer">
              <button className="button" onClick={() => setRecord(null)}>
                Cancel
              </button>
              <button
                className="button primary"
                onClick={save}
                disabled={saving}
              >
                Save record
              </button>
            </div>
          </section>
        </div>
      )}
    </section>
  );
}
createRoot(document.getElementById("root")!).render(<App />);
