"use client";

export function RetentionPolicyForm({
  action,
  holdAction,
}: {
  action: (formData: FormData) => Promise<void>;
  holdAction: (formData: FormData) => Promise<void>;
}) {
  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <form action={action} className="rounded-2xl border border-glacial-border bg-glacial-surface p-5 space-y-3">
        <h2 className="text-sm font-semibold text-white">New retention policy</h2>
        <label className="block text-xs text-glacial-muted">
          Policy type
          <select
            name="policy_type"
            className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
          >
            <option value="permanent">Permanent</option>
            <option value="time_limited">Time limited</option>
            <option value="litigation_hold">Litigation hold</option>
          </select>
        </label>
        <label className="block text-xs text-glacial-muted">
          Conversation ID (optional)
          <input
            name="conversation_id"
            className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
          />
        </label>
        <label className="block text-xs text-glacial-muted">
          Scope label
          <input
            name="scope_label"
            className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
          />
        </label>
        <button type="submit" className="rounded-lg bg-white/10 px-3 py-2 text-xs text-white">
          Create policy
        </button>
      </form>

      <form action={holdAction} className="rounded-2xl border border-glacial-border bg-glacial-surface p-5 space-y-3">
        <h2 className="text-sm font-semibold text-white">Activate litigation hold</h2>
        <label className="block text-xs text-glacial-muted">
          Matter ID
          <input
            name="matter_id"
            required
            className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
          />
        </label>
        <label className="block text-xs text-glacial-muted">
          Custodian DIDs (comma-separated)
          <input
            name="custodian_dids"
            required
            placeholder="did:key:alice,did:key:bob"
            className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
          />
        </label>
        <label className="block text-xs text-glacial-muted">
          Scope
          <input
            name="scope"
            className="mt-1 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-2 text-sm text-white"
          />
        </label>
        <button type="submit" className="rounded-lg bg-amber-500/20 px-3 py-2 text-xs text-amber-200">
          Place hold
        </button>
      </form>
    </div>
  );
}
