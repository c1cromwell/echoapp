import { NextResponse } from "next/server";
import { isLandingEnabled } from "@/lib/flags";

function newReferralCode(): string {
  return Math.random().toString(36).slice(2, 10);
}

export async function POST(request: Request) {
  if (!isLandingEnabled()) {
    return NextResponse.json({ ok: false, error: "Not available" }, { status: 404 });
  }

  let body: { contact?: unknown; ref?: unknown };
  try {
    body = (await request.json()) as { contact?: unknown; ref?: unknown };
  } catch {
    return NextResponse.json({ ok: false, error: "Invalid request" }, { status: 400 });
  }

  const contact = typeof body.contact === "string" ? body.contact.trim() : "";
  if (contact.length < 3 || contact.length > 320) {
    return NextResponse.json({ ok: false, error: "Enter a valid email or handle" }, { status: 400 });
  }
  const ref = typeof body.ref === "string" ? body.ref.slice(0, 32) : null;

  // TODO(legal-entity): persist { contact, ref, createdAt } to the waitlist store (Supabase table)
  // with double opt-in + consent, and credit the referrer — wire to ECHO's rewards/referral system.
  // No PII is stored yet; this is a stub until the entity + Privacy Policy are live.
  void ref;

  const referralCode = newReferralCode();
  const base = process.env.NEXT_PUBLIC_SITE_URL ?? "";
  const referralUrl = `${base}/join?ref=${referralCode}`;

  return NextResponse.json({ ok: true, referralCode, referralUrl });
}
