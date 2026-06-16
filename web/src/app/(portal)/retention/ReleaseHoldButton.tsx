"use client";

export function ReleaseHoldButton({
  matterId,
  action,
}: {
  matterId: string;
  action: (formData: FormData) => Promise<void>;
}) {
  return (
    <form action={action}>
      <input type="hidden" name="matter_id" value={matterId} />
      <button
        type="submit"
        className="rounded-lg border border-glacial-border px-2 py-1 text-[10px] text-glacial-muted hover:text-white"
      >
        Release hold
      </button>
    </form>
  );
}
