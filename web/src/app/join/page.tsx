import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { isLandingEnabled } from "@/lib/flags";
import { WaitlistForm } from "./WaitlistForm";

export const metadata: Metadata = {
  title: "ECHO — Messaging for humans, not bots",
  description:
    "Private, verified, off-grid messaging. No phone number. No email. Everyone verified. Invite-only.",
};

const VALUE_PROPS: { title: string; body: string; note?: string }[] = [
  {
    title: "Verified humans only",
    body: "Every account is an ID-verified person, and ECHO warns you about impersonators. Bots and scammers don't get in.",
  },
  {
    title: "No phone number. No email.",
    body: "Sign up with Face ID. Nothing to leak, nothing to harvest, no number for anyone to find you by.",
  },
  {
    title: "Works with no internet",
    body: "Message people nearby over Bluetooth mesh — festivals, protests, dead zones.",
    note: "Private beta",
  },
  {
    title: "Private by design",
    body: "End-to-end encrypted. On-device AI that never sends your chats to a server.",
  },
];

export default function JoinPage() {
  if (!isLandingEnabled()) notFound();

  return (
    <main className="min-h-screen bg-glacial-bg text-white">
      <section className="mx-auto max-w-3xl px-6 py-20 text-center">
        <p className="mb-4 inline-block rounded-full border border-glacial-border px-3 py-1 text-xs font-semibold uppercase tracking-wide text-glacial-muted">
          Invite-only · verified humans
        </p>
        <h1 className="text-4xl font-bold leading-tight sm:text-5xl">
          Messaging for humans.
          <br />
          Not bots.
        </h1>
        <p className="mx-auto mt-5 max-w-xl text-lg text-glacial-muted">
          No phone number. No email. Everyone verified. Encrypted by default — and built to work even
          when the internet doesn&apos;t.
        </p>

        <div className="mx-auto mt-8 max-w-md">
          <WaitlistForm />
        </div>
      </section>

      <section className="mx-auto grid max-w-4xl gap-4 px-6 pb-20 sm:grid-cols-2">
        {VALUE_PROPS.map((vp) => (
          <div key={vp.title} className="rounded-2xl border border-glacial-border bg-glacial-surface p-6 text-left">
            <div className="flex items-center gap-2">
              <h3 className="text-lg font-semibold">{vp.title}</h3>
              {vp.note && (
                <span className="rounded-full bg-glacial-accent/15 px-2 py-0.5 text-xs font-medium text-glacial-accent">
                  {vp.note}
                </span>
              )}
            </div>
            <p className="mt-2 text-sm text-glacial-muted">{vp.body}</p>
          </div>
        ))}
      </section>

      <footer className="border-t border-glacial-border">
        <div className="mx-auto flex max-w-4xl flex-col items-center gap-2 px-6 py-8 text-sm text-glacial-muted sm:flex-row sm:justify-between">
          <span>© {new Date().getFullYear()} ECHO</span>
          <nav className="flex gap-6">
            <Link href="/join/privacy" className="hover:text-white">Privacy</Link>
            <Link href="/join/terms" className="hover:text-white">Terms</Link>
          </nav>
        </div>
      </footer>
    </main>
  );
}
