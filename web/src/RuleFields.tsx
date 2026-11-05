import type { Data, Rule } from "./types";

export function RuleFields({
  rule: r,
  setRule,
  data,
  toLocal,
  toISO,
  showName = true,
}: {
  rule: Rule;
  setRule: (r: Rule) => void;
  data: Data;
  toLocal: (value: string) => string;
  toISO: (value: string) => string;
  showName?: boolean;
}) {
  const set = (k: string, v: unknown) => setRule({ ...r, [k]: v });
  return (
    <>
      {showName && (
        <label>
          Constraint name
          <input
            value={r.name}
            onChange={(x) => set("name", x.currentTarget.value)}
          />
        </label>
      )}
      <div className="form-row">
        <label>
          Rule type
          <select
            value={r.type}
            onChange={(x) => set("type", x.currentTarget.value)}
          >
            {[
              ["qualification", "Qualified personnel"],
              ["person", "Named participant"],
              ["asset", "Asset demand"],
              ["allowed", "Allowed window"],
              ["excluded", "Excluded window"],
              ["incompatibility", "Cannot overlap"],
              ["rest", "Minimum rest"],
              ["note", "Unenforced note"],
            ].map(([id, name]) => (
              <option value={id} key={id}>
                {name}
              </option>
            ))}
          </select>
        </label>
        <label>
          Scope
          <select
            value={r.scope}
            onChange={(x) => set("scope", x.currentTarget.value)}
          >
            <option value="event">Attached events</option>
            <option value="global">Entire plan</option>
          </select>
        </label>
      </div>
      {r.type === "qualification" && (
        <label>
          Qualification
          <select
            value={r.qualification_id || ""}
            onChange={(x) => set("qualification_id", x.currentTarget.value)}
          >
            <option value="">Select qualification</option>
            {data.qualifications.map((q) => (
              <option value={q.id} key={q.id}>
                {q.name}
              </option>
            ))}
          </select>
        </label>
      )}
      {r.type === "person" && (
        <label>
          Required person
          <select
            value={r.person_id || ""}
            onChange={(x) => set("person_id", x.currentTarget.value)}
          >
            <option value="">Select person</option>
            {data.people.map((p) => (
              <option value={p.id} key={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
      )}
      {r.type === "asset" && (
        <label>
          Asset
          <select
            value={r.asset_id || ""}
            onChange={(x) => set("asset_id", x.currentTarget.value)}
          >
            <option value="">Select asset</option>
            {data.assets.map((a) => (
              <option value={a.id} key={a.id}>
                {a.name}
              </option>
            ))}
          </select>
        </label>
      )}
      {["asset", "qualification"].includes(r.type) && (
        <label>
          Required quantity
          <input
            type="number"
            min={1}
            value={r.count || ""}
            onChange={(x) => set("count", Number(x.currentTarget.value))}
          />
        </label>
      )}
      {["allowed", "excluded"].includes(r.type) && (
        <div className="form-row">
          <label>
            Window start
            <input
              type="datetime-local"
              value={toLocal(r.start || "")}
              onInput={(x) => set("start", toISO(x.currentTarget.value))}
            />
          </label>
          <label>
            Window end
            <input
              type="datetime-local"
              value={toLocal(r.end || "")}
              onInput={(x) => set("end", toISO(x.currentTarget.value))}
            />
          </label>
        </div>
      )}
      {r.type === "incompatibility" && (
        <>
          <label>
            Incompatible event tag
            <input
              value={r.target_tag || ""}
              placeholder="e.g. flight"
              onChange={(x) => set("target_tag", x.currentTarget.value)}
            />
          </label>
          <label>
            Or specific event
            <select
              value={r.target_event_id || ""}
              onChange={(x) => set("target_event_id", x.currentTarget.value)}
            >
              <option value="">Any matching tag</option>
              {data.events.map((e) => (
                <option value={e.id} key={e.id}>
                  {e.name}
                </option>
              ))}
            </select>
          </label>
        </>
      )}
      {r.type === "rest" && (
        <label>
          Minimum rest (minutes)
          <input
            type="number"
            min={0}
            value={r.minutes || 0}
            onChange={(x) => set("minutes", Number(x.currentTarget.value))}
          />
        </label>
      )}
      {r.type === "note" && (
        <div className="notice error">
          This note cannot be enforced. Remove or replace it before full
          validation.
        </div>
      )}
    </>
  );
}
