"use client";

import type { components } from "@/lib/api/schema";

export type SegmentReport = components["schemas"]["ComplySegmentReport"];

export function SegmentReports({ segments }: { segments: SegmentReport[] }) {
  return (
    <section className="mt-8">
      <h2 className="text-sm font-semibold text-white">Segment reports</h2>
      <p className="mt-1 text-xs text-glacial-muted">
        HIPAA, FOIA, and law-firm verticals — aggregate counts only (WO-313).
      </p>
      <div className="mt-4 grid gap-4 lg:grid-cols-3">
        {segments.map((seg) => (
          <div
            key={seg.segment}
            className="rounded-2xl border border-glacial-border bg-glacial-surface p-5"
          >
            <div className="flex items-center justify-between gap-2">
              <h3 className="text-sm font-medium text-white">{seg.label}</h3>
              <span
                className={`text-[10px] uppercase ${
                  seg.status === "active" ? "text-emerald-400" : "text-glacial-muted"
                }`}
              >
                {seg.status?.replace("_", " ")}
              </span>
            </div>
            <dl className="mt-4 space-y-2">
              {(seg.metrics ?? []).map((m) => (
                <div key={m.key} className="flex justify-between gap-2 text-xs">
                  <dt className="text-glacial-muted">{m.label}</dt>
                  <dd className="font-mono text-white">{m.value}</dd>
                </div>
              ))}
            </dl>
          </div>
        ))}
      </div>
    </section>
  );
}
