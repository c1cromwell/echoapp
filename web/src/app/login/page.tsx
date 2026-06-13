"use client";

import { useState } from "react";
import { createClient } from "@/lib/supabase/client";

/**
 * Operator sign-in. Enterprise SSO/SAML/OIDC is configured per-tenant in Supabase Auth
 * (SAML 2.0 / OIDC providers); email magic-link is the dev/fallback path.
 */
export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "sent" | "error">("idle");

  const supabase = createClient();

  async function signInWithSSO(domain: string) {
    const { data, error } = await supabase.auth.signInWithSSO({ domain });
    if (error) return setStatus("error");
    if (data?.url) window.location.href = data.url;
  }

  async function signInWithEmail(e: React.FormEvent) {
    e.preventDefault();
    const { error } = await supabase.auth.signInWithOtp({
      email,
      options: { emailRedirectTo: `${window.location.origin}/auth/callback` },
    });
    setStatus(error ? "error" : "sent");
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-glacial-bg px-4">
      <div className="w-full max-w-sm rounded-2xl border border-glacial-border bg-glacial-surface p-8">
        <h1 className="text-xl font-semibold text-white">Echo Comply</h1>
        <p className="mt-1 text-sm text-glacial-muted">
          Sign in to your organization&apos;s compliance console.
        </p>

        <button
          onClick={() => signInWithSSO(email.split("@")[1] ?? "")}
          disabled={!email.includes("@")}
          className="mt-6 w-full rounded-lg bg-glacial-accent px-4 py-2 font-medium text-glacial-bg disabled:opacity-40"
        >
          Continue with SSO
        </button>

        <form onSubmit={signInWithEmail} className="mt-4 space-y-3">
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            className="w-full rounded-lg border border-glacial-border bg-glacial-bg px-3 py-2 text-white placeholder:text-glacial-muted"
          />
          <button
            type="submit"
            className="w-full rounded-lg border border-glacial-border px-4 py-2 text-sm text-white"
          >
            Email me a sign-in link
          </button>
        </form>

        {status === "sent" && (
          <p className="mt-4 text-sm text-glacial-accent">Check your email for a sign-in link.</p>
        )}
        {status === "error" && (
          <p className="mt-4 text-sm text-red-400">Sign-in failed. Check the domain/email and try again.</p>
        )}
      </div>
    </main>
  );
}
