import type { Event } from "./types";

export type CalendarEntry = {
  key: string;
  event: Event;
  layer: "saved" | "current" | "proposed" | "unchanged";
  added?: boolean;
};

function samePlacement(a: Event, b: Event) {
  return (
    Date.parse(a.start) === Date.parse(b.start) &&
    a.duration === b.duration &&
    a.name === b.name &&
    a.department_id === b.department_id &&
    a.tier === b.tier
  );
}

// Keep both placements for changes, sharing one card for unchanged events.
// Comparison keys are separate from persisted event IDs.
export function calendarEntries(
  saved: Event[],
  proposed: Event[] | null,
): CalendarEntry[] {
  if (!proposed)
    return saved.map((event) => ({ key: event.id, event, layer: "saved" }));
  const savedByID = new Map(saved.map((e) => [e.id, e]));
  const proposedByID = new Map(proposed.map((e) => [e.id, e]));
  const entries: CalendarEntry[] = saved.map((event) => {
    const candidate = proposedByID.get(event.id);
    return {
      key: "current:" + event.id,
      event,
      layer:
        candidate && samePlacement(event, candidate) ? "unchanged" : "current",
    };
  });
  for (const event of proposed) {
    const original = savedByID.get(event.id);
    if (!original || !samePlacement(original, event)) {
      entries.push({
        key: "proposed:" + event.id,
        event,
        layer: "proposed",
        added: !original,
      });
    }
  }
  return entries;
}
