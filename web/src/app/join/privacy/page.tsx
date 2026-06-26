import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { isLandingEnabled } from "@/lib/flags";

export const metadata: Metadata = { title: "ECHO — Privacy Policy" };

export default function PrivacyPage() {
  if (!isLandingEnabled()) notFound();
  return (
    <main className="min-h-screen bg-glacial-bg text-white">
      <article className="mx-auto max-w-2xl px-6 py-16">
        <Link href="/join" className="text-sm text-glacial-muted hover:text-white">← Back</Link>
        <h1 className="mt-4 text-3xl font-bold">Privacy Policy</h1>
        <p className="mt-2 rounded-lg border border-glacial-border bg-glacial-surface px-4 py-3 text-sm text-glacial-muted">
          Draft — under legal review. The binding version is published at launch by the operating
          entity. This summary is informational only.
        </p>
        <div className="mt-6 space-y-4 text-sm text-glacial-muted">
          <p><strong className="text-white">Messages.</strong> End-to-end encrypted; we cannot read
            your message content.</p>
          <p><strong className="text-white">No phone number or email to use ECHO.</strong> Your
            identity is a key on your device.</p>
          <p><strong className="text-white">Identity verification.</strong> ID/selfie checks are handled
            by a third-party verification provider; ECHO stores a verification result, not your raw
            documents.</p>
          <p><strong className="text-white">Waitlist.</strong> If you join the waitlist, we use your
            email/handle solely to contact you about access; unsubscribe anytime.</p>
          <p><strong className="text-white">Your rights.</strong> Access, correction, and deletion
            requests are honored per applicable law.</p>
          <p>Full draft for counsel review: <code>docs/legal/PRIVACY_POLICY_DRAFT.md</code>.</p>
        </div>
      </article>
    </main>
  );
}
