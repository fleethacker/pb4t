import { useEffect, useRef, useState } from "react";
import {
  Upload,
  FileText,
  ExternalLink,
  Check,
  LoaderCircle,
  ChevronDown,
  Save,
  Plus,
  RefreshCw,
  Trash2,
  Undo2,
} from "lucide-react";
import { api } from "./api";
import { RuleFields } from "./RuleFields";
import type {
  Data,
  SourceReference,
  TrainingDocument,
  TrainingSuggestion,
} from "./types";

export function SourceLink({
  source,
  compact = false,
}: {
  source: SourceReference;
  compact?: boolean;
}) {
  return (
    <div className={compact ? "source-link compact" : "source-link"}>
      <a
        href={`/api/documents/${source.document_id}/file#page=${source.page}`}
        target="_blank"
        rel="noreferrer"
      >
        <FileText size={13} /> Source document · page {source.page}{" "}
        <ExternalLink size={12} />
      </a>
      {!compact && <blockquote>{source.excerpt}</blockquote>}
    </div>
  );
}

export function Documents({
  data,
  week,
  toISO,
  toLocal,
  ai,
  onCreated,
}: {
  data: Data;
  week: string;
  toISO: (s: string) => string;
  toLocal: (s: string) => string;
  ai: boolean;
  onCreated: () => Promise<void>;
}) {
  const [docs, setDocs] = useState<TrainingDocument[]>([]),
    [selected, setSelected] = useState(""),
    [error, setError] = useState(""),
    [uploading, setUploading] = useState(false),
    [loading, setLoading] = useState(true),
    [paused, setPaused] = useState(false),
    [deleted, setDeleted] = useState<{
      documentId: string;
      suggestion: TrainingSuggestion;
    } | null>(null),
    [restoring, setRestoring] = useState(false),
    [checking, setChecking] = useState(false);
  const file = useRef<HTMLInputElement>(null);
  const update = (doc: TrainingDocument) =>
    setDocs((current) =>
      [...current.filter((d) => d.id !== doc.id), doc].sort((a, b) =>
        b.imported_at.localeCompare(a.imported_at),
      ),
    );
  async function reload() {
    const next = await api<TrainingDocument[]>("/api/documents");
    setDocs(next.sort((a, b) => b.imported_at.localeCompare(a.imported_at)));
  }
  useEffect(() => {
    let active = true;
    api<TrainingDocument[]>("/api/documents")
      .then((next) => {
        if (active)
          setDocs(
            next.sort((a, b) => b.imported_at.localeCompare(a.imported_at)),
          );
      })
      .catch((e) => {
        if (active) setError(e.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);
  const analyzing = docs.some((d) => d.status === "analyzing");
  useEffect(() => {
    if (!analyzing || paused) return;
    let active = true,
      pending = false;
    const timer = setInterval(() => {
      if (pending) return;
      pending = true;
      api<TrainingDocument[]>("/api/documents")
        .then((next) => {
          if (active)
            setDocs(
              next.sort((a, b) => b.imported_at.localeCompare(a.imported_at)),
            );
        })
        .catch((e) => {
          if (active) {
            clearInterval(timer);
            setError(e.message);
            setPaused(true);
          }
        })
        .finally(() => {
          pending = false;
        });
    }, 1500);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [analyzing, paused]);
  async function upload(chosen: File) {
    setUploading(true);
    setError("");
    try {
      const form = new FormData();
      form.append("file", chosen);
      const r = await fetch("/api/documents", { method: "POST", body: form });
      const doc = await r.json();
      if (!r.ok) throw new Error(doc.error || "Import failed");
      update(doc);
      setSelected(doc.id);
      setPaused(false);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setUploading(false);
      if (file.current) file.current.value = "";
    }
  }
  const doc = docs.find((d) => d.id === selected) || docs[0];
  return (
    <section className="documents-workspace">
      <div className="document-upload">
        <div>
          <span className="eyebrow">REFERENCE LIBRARY</span>
          <h2>Import a training document</h2>
          <p>
            The original PDF stays in your local database. Suggestions are
            drafts; you decide what becomes a requirement, event template, or
            constraint.
          </p>
        </div>
        <input
          ref={file}
          hidden
          type="file"
          accept=".pdf,application/pdf"
          aria-label="Training PDF"
          onChange={(e) => {
            const chosen = e.currentTarget.files?.[0];
            if (chosen) upload(chosen);
          }}
        />
        <button
          className="button primary"
          disabled={uploading}
          onClick={() => file.current?.click()}
        >
          {uploading ? (
            <LoaderCircle className="spin" size={16} />
          ) : (
            <Upload size={16} />
          )}{" "}
          {uploading ? "Importing…" : "Import PDF"}
        </button>
      </div>
      {error && (
        <div className="notice error">
          {error}
          <button
            className="button"
            onClick={async () => {
              try {
                await reload();
                setError("");
                setPaused(false);
              } catch (e) {
                setError((e as Error).message);
              }
            }}
          >
            <RefreshCw size={14} /> Retry
          </button>
        </div>
      )}
      {!ai && (
        <div className="notice">
          PDFs can be saved and referenced now. Configure a server-side OpenAI
          key and restart to generate suggestions.
        </div>
      )}
      {deleted && (
        <div className="notice" role="status">
          Deleted “{deleted.suggestion.name}”.
          <button
            className="button"
            disabled={restoring}
            onClick={async () => {
              setRestoring(true);
              try {
                update(
                  await api<TrainingDocument>(
                    `/api/documents/${deleted.documentId}/suggestions/${deleted.suggestion.id}`,
                    "PUT",
                    deleted.suggestion,
                  ),
                );
                setDeleted(null);
              } catch (e) {
                setError((e as Error).message);
              } finally {
                setRestoring(false);
              }
            }}
          >
            <Undo2 size={14} /> Undo
          </button>
        </div>
      )}
      {loading ? (
        <div className="empty">
          <LoaderCircle className="spin" /> Loading documents…
        </div>
      ) : !doc ? (
        <div className="empty document-empty">
          <FileText size={36} />
          <h3>No training documents</h3>
          <p>
            Import a PDF to review suggested requirements, reusable activities,
            and constraints, with page references back to the original.
          </p>
        </div>
      ) : (
        <div className="document-layout">
          <aside className="document-list">
            {docs.map((d) => (
              <button
                key={d.id}
                className={d.id === doc.id ? "selected" : ""}
                onClick={() => setSelected(d.id)}
              >
                <FileText size={18} />
                <span>
                  <strong>{d.title || d.filename}</strong>
                  <small>
                    {new Date(d.imported_at).toLocaleDateString()} ·{" "}
                    {Math.ceil(d.size / 1024)} KB
                  </small>
                  <small>
                    {d.status === "ready"
                      ? `${d.suggestions.filter((s) => !s.created_id && !s.deleted).length} suggestions to review`
                      : d.status}
                  </small>
                </span>
              </button>
            ))}
          </aside>
          <div className="document-review">
            <div className="document-heading">
              <div>
                <span className="eyebrow">SOURCE DOCUMENT</span>
                <h2>{doc.title}</h2>
                <p>{doc.filename}</p>
              </div>
              <a
                className="button"
                href={`/api/documents/${doc.id}/file`}
                target="_blank"
                rel="noreferrer"
              >
                <ExternalLink size={14} /> Open original PDF
              </a>
              {doc.status === "ready" && (
                <button
                  className="button"
                  disabled={checking}
                  onClick={async () => {
                    setChecking(true);
                    setError("");
                    try {
                      update(
                        await api<TrainingDocument>(
                          `/api/documents/${doc.id}/citations/check`,
                          "POST",
                        ),
                      );
                    } catch (e) {
                      setError((e as Error).message);
                    } finally {
                      setChecking(false);
                    }
                  }}
                >
                  <RefreshCw size={14} />{" "}
                  {checking ? "Checking citations…" : "Recheck citations"}
                </button>
              )}
            </div>
            {doc.status === "analyzing" ? (
              <div className="run-progress">
                <LoaderCircle className="spin" />
                <strong>Reading the training document</strong>
                <p>
                  Identifying source-cited suggestions. Nothing is added to your
                  calendar or planning inputs automatically.
                </p>
              </div>
            ) : doc.status === "failed" ? (
              <div className="notice error">
                <div>
                  <strong>Analysis could not finish</strong>
                  <p>{doc.error}</p>
                  <p>The original PDF is saved and can still be opened.</p>
                </div>
              </div>
            ) : null}
            {(doc.status === "uploaded" || doc.status === "failed") && (
              <button
                className="button"
                disabled={!ai}
                onClick={async () => {
                  setError("");
                  try {
                    update(
                      await api<TrainingDocument>(
                        `/api/documents/${doc.id}/analyze`,
                        "POST",
                      ),
                    );
                    setPaused(false);
                  } catch (e) {
                    setError((e as Error).message);
                  }
                }}
              >
                <RefreshCw size={14} />{" "}
                {doc.status === "failed"
                  ? "Retry analysis"
                  : "Generate suggestions"}
              </button>
            )}
            {doc.summary && <p className="document-summary">{doc.summary}</p>}
            {doc.status === "ready" && (
              <div className="suggestion-intro">
                <strong>Review before creating</strong>
                <p>
                  Recommendations are distinguished from stated obligations.
                  Missing durations, counts, departments, and scheduling periods
                  require your decision. Source excerpts are AI-extracted;
                  verify them against the PDF.
                </p>
              </div>
            )}
            {doc.status === "ready" &&
              doc.suggestions.every((s) => s.deleted) && (
                <p className="empty">
                  No suggestions to display. The PDF remains available for
                  reference.
                </p>
              )}
            {doc.suggestions
              .filter((s) => !s.deleted)
              .map((s) => (
                <SuggestionCard
                  key={doc.id + s.id}
                  doc={doc}
                  suggestion={s}
                  data={data}
                  week={week}
                  toISO={toISO}
                  toLocal={toLocal}
                  updated={update}
                  onDeleted={() =>
                    setDeleted({ documentId: doc.id, suggestion: s })
                  }
                  onCreated={onCreated}
                />
              ))}
          </div>
        </div>
      )}
    </section>
  );
}

function SuggestionCard({
  doc,
  suggestion: s,
  data,
  week,
  toISO,
  toLocal,
  updated,
  onDeleted,
  onCreated,
}: {
  doc: TrainingDocument;
  suggestion: TrainingSuggestion;
  data: Data;
  week: string;
  toISO: (s: string) => string;
  toLocal: (s: string) => string;
  updated: (d: TrainingDocument) => void;
  onDeleted: () => void;
  onCreated: () => Promise<void>;
}) {
  const [open, setOpen] = useState(false),
    [draft, setDraft] = useState(s),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  useEffect(() => {
    setDraft(s);
  }, [s]);
  const set = (key: keyof TrainingSuggestion, value: unknown) =>
    setDraft((current) => ({
      ...current,
      [key]: value,
      ...(key === "name" && current.rule
        ? { rule: { ...current.rule, name: String(value), confirmed: false } }
        : {}),
    }));
  function edit() {
    setDraft({
      ...s,
      start: s.start?.startsWith("0001") ? toISO(week + "T00:00") : s.start,
      end: s.end?.startsWith("0001")
        ? toISO(nextDay(week, 7) + "T00:00")
        : s.end,
    });
    setOpen(!open);
    setError("");
  }
  async function save(create: boolean) {
    setBusy(true);
    setError("");
    try {
      const next = await api<TrainingDocument>(
        `/api/documents/${doc.id}/suggestions/${s.id}` +
          (create ? "/create" : ""),
        create ? "POST" : "PUT",
        draft,
      );
      updated(next);
      if (create) {
        setOpen(false);
        await onCreated();
      }
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function remove() {
    setBusy(true);
    setError("");
    try {
      const next = await api<TrainingDocument>(
        `/api/documents/${doc.id}/suggestions/${s.id}`,
        "DELETE",
      );
      onDeleted();
      updated(next);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <article className={"suggestion-card " + (s.created_id ? "created" : "")}>
      <div className="suggestion-title">
        <span className="pill">
          {s.kind === "template"
            ? "Event template"
            : s.kind === "constraint"
              ? "Constraint"
              : "Training requirement"}
        </span>
        {s.kind === "constraint" && (
          <span className="pill amber-pill">
            {s.rule?.type === "note"
              ? "Unenforced note"
              : s.created_id
                ? "Confirmed rule"
                : "Rule draft · not active"}
          </span>
        )}
        <span
          className={"pill " + (s.basis === "required" ? "amber-pill" : "")}
        >
          {s.basis === "required"
            ? "Source states required"
            : "Source recommends"}
        </span>
        {s.created_id && (
          <span className="pill green-pill">
            <Check size={12} /> Created
          </span>
        )}
      </div>
      <h3>{s.name}</h3>
      <p>{s.description}</p>
      <SourceLink
        source={{ document_id: doc.id, page: s.page, excerpt: s.excerpt }}
      />
      {s.citation_status === "needs_review" && (
        <p className="notice">
          Source citation needs review. Open the original PDF, then use Review
          &amp; edit → Source citation to confirm the wording and page.
        </p>
      )}
      {s.review_notes && <p className="review-note">{s.review_notes}</p>}
      {s.missing?.length > 0 && (
        <p className="missing-fields">Needs review: {s.missing.join(", ")}</p>
      )}
      {!s.created_id && (
        <button className="button" onClick={edit}>
          Review & edit <ChevronDown size={14} />
        </button>
      )}
      {s.created_id && (
        <p className="creation-note">
          Created in Planning inputs with a permanent reference to this
          document. Further edits are available there.
        </p>
      )}
      {open && !s.created_id && (
        <div className="suggestion-form">
          {error && <div className="notice error">{error}</div>}
          <label>
            Name
            <input
              value={draft.name}
              onChange={(e) => set("name", e.currentTarget.value)}
            />
          </label>
          <label>
            Description
            <textarea
              value={draft.description}
              onChange={(e) => set("description", e.currentTarget.value)}
            />
          </label>
          <details>
            <summary>Source citation</summary>
            <p className="review-note">
              Correct any shortened AI wording using the original PDF. Saving
              the draft checks the quotation locally.
            </p>
            <label>
              Source PDF page
              <input
                type="number"
                min={1}
                value={draft.page}
                onChange={(e) => set("page", Number(e.currentTarget.value))}
              />
            </label>
            <label>
              Source excerpt
              <textarea
                value={draft.excerpt}
                onChange={(e) => set("excerpt", e.currentTarget.value)}
              />
            </label>
          </details>
          {s.kind !== "constraint" && (
            <div className="form-row">
              <label>
                Department
                <select
                  value={draft.department_id}
                  onChange={(e) => set("department_id", e.currentTarget.value)}
                >
                  <option value="">Choose a department…</option>
                  {data.departments.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Drill type
                <input
                  value={draft.type}
                  onChange={(e) => set("type", e.currentTarget.value)}
                />
              </label>
            </div>
          )}
          {s.kind === "constraint" ? (
            <>
              <RuleFields
                rule={
                  draft.rule || {
                    id: s.id,
                    name: draft.name,
                    type: "note",
                    scope: "event",
                    confirmed: false,
                  }
                }
                setRule={(rule) =>
                  setDraft((current) => ({
                    ...current,
                    name: rule.name,
                    rule: { ...rule, confirmed: false },
                  }))
                }
                data={data}
                toLocal={toLocal}
                toISO={toISO}
                showName={false}
              />
              <p className="review-note">
                Attached-event rules apply only after you attach them to event
                templates or events. Entire-plan rules apply to all events.
                Notes are saved for reference and never enforced.
              </p>
              {draft.rule?.type !== "note" && (
                <label className="inline-check">
                  <input
                    type="checkbox"
                    checked={draft.rule?.confirmed || false}
                    onChange={(e) => {
                      const confirmed = e.currentTarget.checked;
                      setDraft((current) => ({
                        ...current,
                        rule: {
                          ...current.rule!,
                          confirmed,
                        },
                      }));
                    }}
                  />
                  I reviewed this rule and want it available for scheduling
                </label>
              )}
            </>
          ) : s.kind === "template" ? (
            <>
              <div className="form-row">
                <label>
                  Point of contact
                  <input
                    value={draft.contact}
                    onChange={(e) => set("contact", e.currentTarget.value)}
                  />
                </label>
                <label>
                  Duration (minutes)
                  {s.duration === 0 && (
                    <small>Not specified by source—enter your choice.</small>
                  )}
                  <input
                    type="number"
                    min={15}
                    step={15}
                    value={draft.duration || ""}
                    onChange={(e) =>
                      set("duration", Number(e.currentTarget.value))
                    }
                  />
                </label>
              </div>
              <label>
                Tags (comma separated)
                <input
                  defaultValue={(draft.tags || []).join(", ")}
                  onBlur={(e) =>
                    set(
                      "tags",
                      e.currentTarget.value
                        .split(",")
                        .map((t) => t.trim())
                        .filter(Boolean),
                    )
                  }
                />
              </label>
              <fieldset>
                <legend>Attach existing constraints (your choice)</legend>
                {data.constraints
                  .filter((c) => c.scope !== "global")
                  .map((c) => (
                    <label className="inline-check" key={c.id}>
                      <input
                        type="checkbox"
                        checked={(draft.constraint_ids || []).includes(c.id)}
                        onChange={(e) =>
                          set(
                            "constraint_ids",
                            e.currentTarget.checked
                              ? [...(draft.constraint_ids || []), c.id]
                              : (draft.constraint_ids || []).filter(
                                  (id) => id !== c.id,
                                ),
                          )
                        }
                      />
                      {c.name}
                    </label>
                  ))}
              </fieldset>
              <label className="inline-check">
                <input
                  type="checkbox"
                  checked={draft.fatigue}
                  onChange={(e) => set("fatigue", e.currentTarget.checked)}
                />{" "}
                Count this activity toward fatigue
              </label>
            </>
          ) : (
            <>
              <p className="review-note">
                Choose the planning window and required count. These are your
                planning decisions unless the source explicitly states them.
              </p>
              <div className="form-row">
                <label>
                  Coverage window start
                  <input
                    type="date"
                    value={draft.start?.slice(0, 10) || ""}
                    onInput={(e) =>
                      set("start", toISO(e.currentTarget.value + "T00:00"))
                    }
                  />
                </label>
                <label>
                  Coverage window end (exclusive)
                  <input
                    type="date"
                    value={draft.end?.slice(0, 10) || ""}
                    onInput={(e) =>
                      set("end", toISO(e.currentTarget.value + "T00:00"))
                    }
                  />
                </label>
              </div>
              <div className="form-row">
                <label>
                  Required occurrences
                  <input
                    type="number"
                    min={1}
                    value={draft.count || ""}
                    onChange={(e) =>
                      set("count", Number(e.currentTarget.value))
                    }
                  />
                </label>
                <label>
                  Repeat every (days; 0 = no repeat)
                  <input
                    type="number"
                    min={0}
                    value={draft.repeat_days}
                    onChange={(e) =>
                      set("repeat_days", Number(e.currentTarget.value))
                    }
                  />
                </label>
              </div>
            </>
          )}
          <label>
            Review notes
            <textarea
              value={draft.review_notes}
              onChange={(e) => set("review_notes", e.currentTarget.value)}
            />
          </label>
          <div className="suggestion-actions">
            <button className="button danger" disabled={busy} onClick={remove}>
              <Trash2 size={14} /> Delete suggestion
            </button>
            <button
              className="button"
              disabled={busy}
              onClick={() => save(false)}
            >
              <Save size={14} /> Save draft edits
            </button>
            <button
              className="button primary"
              disabled={busy}
              onClick={() => save(true)}
            >
              {busy ? (
                <LoaderCircle size={14} className="spin" />
              ) : (
                <Plus size={14} />
              )}{" "}
              {s.kind === "constraint"
                ? draft.rule?.type === "note"
                  ? "Save unenforced note"
                  : "Create constraint"
                : s.kind === "template"
                  ? "Create event template"
                  : "Create training requirement"}
            </button>
          </div>
        </div>
      )}
    </article>
  );
}
function nextDay(date: string, n: number) {
  const d = new Date(date + "T12:00:00Z");
  d.setUTCDate(d.getUTCDate() + n);
  return d.toISOString().slice(0, 10);
}
