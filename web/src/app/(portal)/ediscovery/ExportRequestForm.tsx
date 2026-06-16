"use client";

export function ExportRequestForm({ action }: { action: (formData: FormData) => Promise<void> }) {
  return (
    <form
      action={action}
      className="max-w-lg rounded-2xl border border-glacial-border bg-glacial-surface p-5 space-y-3"
    >
      <h2 className="text-sm font-semibold text-white">Request eDiscovery export</h2>
      <label className="block text-xs text-glacial-muted">
        Matter ID
        <input
          name="matter_id"
          required
          className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
        />
      </label>
      <label className="block text-xs text-glacial-muted">
        Custodian set (optional, comma-separated DIDs)
        <input
          name="custodian_set"
          className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
        />
      </label>
      <button type="submit" className="rounded-lg bg-white/10 px-3 py-2 text-xs text-white">
        Start export
      </button>
    </form>
  );
}
