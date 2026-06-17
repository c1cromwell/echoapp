/**
 * Zero-PII gate for Comply portal payloads (WO-309 CI).
 * Fails if sample dashboard/audit JSON contains forbidden PII field names.
 */

const FORBIDDEN_KEYS = [
  "messageContent",
  "message_content",
  "plaintext",
  "body",
  "email",
  "phone",
  "displayName",
  "display_name",
  "fullName",
  "full_name",
  "custodianEmail",
  "custodianName",
  "userName",
  "user_name",
  "content",
  "text",
  "subject",
  "address",
  "ssn",
  "socialSecurity",
];

const ALLOWED_KEYS = new Set([
  "deCoverageRate",
  "activeRetentionPolicies",
  "litigationHolds",
  "pendingExports",
  "anchorHealth",
  "orgDid",
  "generatedAt",
  "activeLitigationHolds",
  "verificationNotice",
  "id",
  "policyType",
  "conversationId",
  "scopeLabel",
  "active",
  "matterId",
  "status",
  "custodianCount",
  "dataL1Ref",
  "exportId",
  "messageCount",
  "queryHash",
  "policies",
  "exports",
  "segments",
  "status",
  "metrics",
]);

function collectKeys(value, path = "$", keys = new Set()) {
  if (value === null || value === undefined) return keys;
  if (Array.isArray(value)) {
    for (let i = 0; i < value.length; i++) {
      collectKeys(value[i], `${path}[${i}]`, keys);
    }
    return keys;
  }
  if (typeof value === "object") {
    for (const [k, v] of Object.entries(value)) {
      keys.add(`${path}.${k}`);
      if (FORBIDDEN_KEYS.includes(k)) {
        throw new Error(`Forbidden PII key "${k}" at ${path}`);
      }
      collectKeys(v, `${path}.${k}`, keys);
    }
  }
  return keys;
}

const dashboardSample = {
  deCoverageRate: "42%",
  activeRetentionPolicies: 3,
  litigationHolds: 1,
  pendingExports: 0,
  anchorHealth: "healthy",
};

const auditSample = {
  orgDid: "did:key:z6Mkexample",
  generatedAt: new Date().toISOString(),
  activeRetentionPolicies: 3,
  activeLitigationHolds: 1,
  pendingExports: 0,
  anchorHealth: "healthy",
  verificationNotice: "Aggregate metrics only — no message content.",
};

const policyListSample = {
  policies: [
    {
      id: "pol-1",
      orgDid: "did:key:z6Mkexample",
      policyType: "permanent",
      conversationId: "conv-hash-abc",
      scopeLabel: "Legal hold scope",
      active: true,
    },
  ],
};

const exportListSample = {
  exports: [
    {
      exportId: "exp-1",
      matterId: "matter-1",
      status: "pending",
      messageCount: 120,
      queryHash: "sha256:abc123",
      dataL1Ref: "bafyexample",
    },
  ],
};

const segmentSample = {
  orgDid: "did:key:z6Mkexample",
  segments: [
    {
      segment: "hipaa",
      label: "Healthcare (HIPAA)",
      status: "not_configured",
      metrics: [{ key: "retentionPolicies", label: "ePHI retention policies", value: "0" }],
    },
  ],
};

for (const sample of [dashboardSample, auditSample, policyListSample, exportListSample, segmentSample]) {
  collectKeys(sample);
}

console.log("zero-PII payload test passed (forbidden keys absent in sample Comply shapes)");
