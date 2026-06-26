import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { isLandingEnabled } from "@/lib/flags";

export const metadata: Metadata = { title: "ECHO — Terms of Service" };

export default function TermsPage() {
  if (!isLandingEnabled()) notFound();
  return (
    <main className="min-h-screen bg-glacial-bg text-white">
      <article className="mx-auto max-w-2xl px-6 py-16">
        <Link href="/join" className="text-sm text-glacial-muted hover:text-white">← Back</Link>
        <h1 className="mt-4 text-3xl font-bold">Terms of Service</h1>
        <p className="mt-2 rounded-lg border border-glacial-border bg-glacial-surface px-4 py-3 text-sm text-glacial-muted">
          Draft — under legal review. The binding version is published at launch by the operating
          entity. This summary is informational only and is not a contract.
        </p>
        <div className="mt-6 space-y-4 text-sm text-glacial-muted">
          <p><strong className="text-white">Beta service.</strong> ECHO is an invite-only beta provided
            &quot;as is&quot;; features may change and some (e.g. offline mesh, in-chat payments) are not
            yet generally available.</p>
          <p><strong className="text-white">Eligibility &amp; verification.</strong> Access requires
            identity verification; you must be of legal age and provide accurate information.</p>
          <p><strong className="text-white">Acceptable use.</strong> No illegal activity, harassment,
            impersonation, spam, or attempts to circumvent verification.</p>
          <p><strong className="text-white">Your content.</strong> Messages are end-to-end encrypted;
            you are responsible for your communications.</p>
          <p><strong className="text-white">No warranties / liability cap.</strong> Provided without
            warranties to the extent permitted by law; liability is limited.</p>
          <p>Full draft for counsel review: <code>docs/legal/TERMS_OF_SERVICE_DRAFT.md</code>.</p>
        </div>
      </article>
    </main>
  );
}
