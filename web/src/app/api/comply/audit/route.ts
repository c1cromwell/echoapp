import { getActiveOrg } from "@/lib/auth/org";

const BASE = process.env.COMPLY_API_BASE_URL;
const TOKEN = process.env.COMPLY_API_SERVICE_TOKEN;

/** Server-side audit PDF proxy (service token never exposed to browser). */
export async function GET() {
  const org = await getActiveOrg();
  if (!org || !BASE) {
    return new Response("Not configured", { status: 503 });
  }

  const res = await fetch(`${BASE}/comply/audit/report?format=pdf`, {
    headers: {
      Authorization: `Bearer ${TOKEN ?? ""}`,
      "X-Org-DID": org.orgDID,
    },
    cache: "no-store",
  });

  if (!res.ok) {
    return new Response("Comply audit export failed", { status: res.status });
  }

  const body = await res.arrayBuffer();
  return new Response(body, {
    headers: {
      "Content-Type": "application/pdf",
      "Content-Disposition": 'attachment; filename="echo-comply-audit.pdf"',
    },
  });
}
