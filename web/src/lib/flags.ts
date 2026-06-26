// Feature flags for the web app.
//
// The public launch landing page ("/join") + waitlist are OFF by default and gated behind a
// server-side env flag so they can be flipped on at deploy time once the legal entity + domain
// are ready — without shipping a half-built marketing surface in the meantime.

/** True only when the launch landing + waitlist should be served. */
export function isLandingEnabled(): boolean {
  return process.env.LANDING_ENABLED === "true";
}
