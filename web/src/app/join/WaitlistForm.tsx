"use client";

import { useEffect, useState } from "react";

type Result = { referralUrl: string } | null;

export function WaitlistForm() {
  const [contact, setContact] = useState("");
  const [ref, setRef] = useState<string | null>(null);
  const [status, setStatus] = useState<"idle" | "submitting" | "done" | "error">("idle");
  const [error, setError] = useState("");
  const [result, setResult] = useState<Result>(null);
  const [copied, setCopied] = useState(false);

  // Read the referral code from the URL once mounted (avoids a useSearchParams Suspense boundary).
  useEffect(() => {
    const code = new URLSearchParams(window.location.search).get("ref");
    if (code) setRef(code);
  }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!contact.trim()) return;
    setStatus("submitting");
    setError("");
    try {
      const res = await fetch("/api/waitlist", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ contact: contact.trim(), ref }),
      });
      const data = await res.json();
      if (!res.ok || !data.ok) throw new Error(data.error ?? "Something went wrong");
      setResult({ referralUrl: data.referralUrl });
      setStatus("done");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
      setStatus("error");
    }
  }

  if (status === "done" && result) {
    return (
      <div className="rounded-2xl border border-glacial-border bg-glacial-surface p-6 text-left">
        <p className="text-lg font-semibold">You&apos;re on the list ✅</p>
        <p className="mt-1 text-sm text-glacial-muted">
          Verified humans move up faster. Share your link to skip ahead:
        </p>
        <div className="mt-3 flex gap-2">
          <input
            readOnly
            value={result.referralUrl}
            className="w-full rounded-lg border border-glacial-border bg-glacial-bg px-3 py-2 text-sm"
          />
          <button
            type="button"
            onClick={() => {
              void navigator.clipboard.writeText(result.referralUrl);
              setCopied(true);
              setTimeout(() => setCopied(false), 1500);
            }}
            className="shrink-0 rounded-lg bg-glacial-accent px-3 py-2 text-sm font-semibold text-glacial-bg"
          >
            {copied ? "Copied" : "Copy"}
          </button>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-3 sm:flex-row">
      <input
        type="text"
        inputMode="email"
        autoComplete="email"
        placeholder="Email or @handle"
        value={contact}
        onChange={(e) => setContact(e.target.value)}
        className="w-full rounded-xl border border-glacial-border bg-glacial-surface px-4 py-3 text-sm placeholder:text-glacial-muted focus:border-glacial-accent focus:outline-none"
        aria-label="Email or social handle"
      />
      <button
        type="submit"
        disabled={status === "submitting"}
        className="shrink-0 rounded-xl bg-glacial-accent px-5 py-3 text-sm font-semibold text-glacial-bg disabled:opacity-60"
      >
        {status === "submitting" ? "…" : "Request an invite"}
      </button>
      {status === "error" && (
        <p className="text-sm text-red-400 sm:hidden" role="alert">{error}</p>
      )}
    </form>
  );
}
