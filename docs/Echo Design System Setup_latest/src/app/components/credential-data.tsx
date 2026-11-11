// Credential Verification types & mock data matching the Go backend models

export type CredentialType =
  | "passport"
  | "drivers_license"
  | "national_id"
  | "bank_account"
  | "education"
  | "employment"
  | "phone"
  | "email";

export type IssuerType =
  | "government"
  | "financial"
  | "educational"
  | "employment"
  | "telecom"
  | "identity_provider";

export type IssuerTrustLevel = "high" | "medium" | "basic";

export type Jurisdiction = "US" | "EU" | "UK" | "CA" | "AU" | "Global";

export type VerificationStage =
  | "structure"
  | "issuer_trust"
  | "cryptographic"
  | "revocation"
  | "claims";

export interface TrustedIssuer {
  id: string;
  did: string;
  name: string;
  trustLevel: IssuerTrustLevel;
  type: IssuerType;
  jurisdiction: Jurisdiction;
  supportedCredentials: CredentialType[];
  riskScore: number; // 0-100, lower is better
  onboardingWeight: number;
  status: "active" | "suspended" | "revoked";
  qualifications: { name: string; expiresAt: string }[];
  anchorHash?: string;
}

export interface VerifiableCredential {
  id: string;
  type: CredentialType;
  issuer: string;
  issuerName: string;
  issuanceDate: string;
  expirationDate: string;
  claims: Record<string, string>;
  proof: { type: string; created: string };
  status: "active" | "revoked";
}

export interface VerificationResult {
  valid: boolean;
  credentialId: string;
  credentialType: CredentialType;
  stages: {
    stage: VerificationStage;
    label: string;
    passed: boolean;
    message: string;
  }[];
  extractedClaims: Record<string, string>;
  issuerInfo: TrustedIssuer;
  badge: string;
  trustPoints: number;
}

export interface TrustScoreBreakdown {
  baseScore: number;
  credentialScores: { type: CredentialType; points: number; badge: string }[];
  multiCredentialBonus: number;
  totalScore: number;
}

export interface FunnelMetrics {
  sessions: number;
  steps: {
    name: string;
    count: number;
    dropoffRate: number;
    skipped: number;
    avgDuration: number; // ms
  }[];
  conversionRate: number;
  avgSessionDuration: number;
  credentialStats: {
    type: CredentialType;
    attempts: number;
    successes: number;
    failures: number;
  }[];
  issuerDistribution: { issuer: string; count: number }[];
}

// Credential type display info
export const credentialTypeInfo: Record<
  CredentialType,
  { emoji: string; label: string; description: string; pointsHigh: number; pointsMedium: number; pointsBasic: number }
> = {
  passport: { emoji: "🛂", label: "Passport", description: "Government-issued travel document", pointsHigh: 90, pointsMedium: 75, pointsBasic: 60 },
  drivers_license: { emoji: "🪪", label: "Driver's License", description: "Government-issued photo ID", pointsHigh: 85, pointsMedium: 70, pointsBasic: 55 },
  national_id: { emoji: "🆔", label: "National ID", description: "Official identification card", pointsHigh: 85, pointsMedium: 70, pointsBasic: 55 },
  bank_account: { emoji: "🏦", label: "Bank Account", description: "Financial institution verification", pointsHigh: 75, pointsMedium: 60, pointsBasic: 45 },
  education: { emoji: "🎓", label: "Education", description: "Educational institution credential", pointsHigh: 60, pointsMedium: 50, pointsBasic: 35 },
  employment: { emoji: "💼", label: "Employment", description: "Employment verification credential", pointsHigh: 55, pointsMedium: 45, pointsBasic: 30 },
  phone: { emoji: "📱", label: "Phone Number", description: "Telecom provider verification", pointsHigh: 35, pointsMedium: 30, pointsBasic: 20 },
  email: { emoji: "📧", label: "Email Address", description: "Email provider verification", pointsHigh: 30, pointsMedium: 25, pointsBasic: 15 },
};

export const issuerTypeInfo: Record<IssuerType, { emoji: string; label: string; color: string }> = {
  government: { emoji: "🏛️", label: "Government", color: "#3B82F6" },
  financial: { emoji: "🏦", label: "Financial", color: "#10B981" },
  educational: { emoji: "🎓", label: "Educational", color: "#8B5CF6" },
  employment: { emoji: "💼", label: "Employment", color: "#F59E0B" },
  telecom: { emoji: "📡", label: "Telecom", color: "#06B6D4" },
  identity_provider: { emoji: "🔐", label: "Identity Provider", color: "#F43F5E" },
};

export const trustLevelColors: Record<IssuerTrustLevel, { color: string; bg: string; label: string }> = {
  high: { color: "#10B981", bg: "#ECFDF5", label: "High Trust" },
  medium: { color: "#F59E0B", bg: "#FFFBEB", label: "Medium Trust" },
  basic: { color: "#64748B", bg: "#F1F5F9", label: "Basic Trust" },
};

export const verificationStageInfo: Record<
  VerificationStage,
  { emoji: string; label: string; description: string }
> = {
  structure: { emoji: "📋", label: "Structure Validation", description: "Checking credential format and required fields" },
  issuer_trust: { emoji: "🏛️", label: "Issuer Verification", description: "Validating issuer is registered and trusted" },
  cryptographic: { emoji: "🔐", label: "Cryptographic Proof", description: "Verifying digital signature integrity" },
  revocation: { emoji: "✅", label: "Revocation Check", description: "Confirming credential hasn't been revoked" },
  claims: { emoji: "📝", label: "Claims Extraction", description: "Extracting and validating identity claims" },
};

// Mock trusted issuers (matches Go backend well-known issuers)
export const mockTrustedIssuers: TrustedIssuer[] = [
  {
    id: "gvt_us_dmv",
    did: "did:echo:gvt_us_dmv",
    name: "US Department of Motor Vehicles",
    trustLevel: "high",
    type: "government",
    jurisdiction: "US",
    supportedCredentials: ["passport", "drivers_license", "national_id"],
    riskScore: 5,
    onboardingWeight: 95,
    status: "active",
    qualifications: [
      { name: "SOC2 Type II", expiresAt: "2027-01-15" },
      { name: "ISO 27001", expiresAt: "2026-09-01" },
    ],
    anchorHash: "dag://L0:abc123...ef456",
  },
  {
    id: "fin_wells_fargo",
    did: "did:echo:fin_wells_fargo",
    name: "Wells Fargo",
    trustLevel: "high",
    type: "financial",
    jurisdiction: "US",
    supportedCredentials: ["bank_account"],
    riskScore: 12,
    onboardingWeight: 80,
    status: "active",
    qualifications: [
      { name: "PCI DSS Level 1", expiresAt: "2026-12-31" },
      { name: "SOC2 Type II", expiresAt: "2027-03-15" },
    ],
  },
  {
    id: "edu_stanford",
    did: "did:echo:edu_stanford",
    name: "Stanford University",
    trustLevel: "medium",
    type: "educational",
    jurisdiction: "US",
    supportedCredentials: ["education"],
    riskScore: 18,
    onboardingWeight: 60,
    status: "active",
    qualifications: [
      { name: "FERPA Compliant", expiresAt: "2027-08-01" },
    ],
  },
  {
    id: "gvt_uk_hmpo",
    did: "did:echo:gvt_uk_hmpo",
    name: "UK HM Passport Office",
    trustLevel: "high",
    type: "government",
    jurisdiction: "UK",
    supportedCredentials: ["passport", "national_id"],
    riskScore: 3,
    onboardingWeight: 95,
    status: "active",
    qualifications: [
      { name: "ISO 27001", expiresAt: "2027-06-01" },
    ],
  },
  {
    id: "fin_revolut",
    did: "did:echo:fin_revolut",
    name: "Revolut",
    trustLevel: "medium",
    type: "financial",
    jurisdiction: "EU",
    supportedCredentials: ["bank_account"],
    riskScore: 22,
    onboardingWeight: 65,
    status: "active",
    qualifications: [
      { name: "PSD2 Licensed", expiresAt: "2026-11-01" },
    ],
  },
  {
    id: "tel_twilio",
    did: "did:echo:tel_twilio",
    name: "Twilio",
    trustLevel: "medium",
    type: "telecom",
    jurisdiction: "Global",
    supportedCredentials: ["phone"],
    riskScore: 15,
    onboardingWeight: 40,
    status: "active",
    qualifications: [],
  },
];

// Mock verification results for demo
export function simulateVerification(
  credType: CredentialType,
  issuer: TrustedIssuer
): VerificationResult {
  const info = credentialTypeInfo[credType];
  const points =
    issuer.trustLevel === "high"
      ? info.pointsHigh
      : issuer.trustLevel === "medium"
      ? info.pointsMedium
      : info.pointsBasic;

  return {
    valid: true,
    credentialId: `vc_${credType}_${Date.now()}`,
    credentialType: credType,
    stages: [
      { stage: "structure", label: "Structure Validation", passed: true, message: "All required fields present" },
      { stage: "issuer_trust", label: "Issuer Verification", passed: true, message: `${issuer.name} is a trusted ${issuer.trustLevel}-level issuer` },
      { stage: "cryptographic", label: "Cryptographic Proof", passed: true, message: "Ed25519 signature verified" },
      { stage: "revocation", label: "Revocation Check", passed: true, message: "Credential is active and valid" },
      { stage: "claims", label: "Claims Extraction", passed: true, message: "Identity claims successfully extracted" },
    ],
    extractedClaims: {
      type: credType,
      issuer: issuer.name,
      jurisdiction: issuer.jurisdiction,
    },
    issuerInfo: issuer,
    badge: `${info.emoji} ${info.label} Verified`,
    trustPoints: points,
  };
}

export function calculateTrustScore(
  results: VerificationResult[],
  baseScore: number = 15
): TrustScoreBreakdown {
  const credentialScores = results
    .filter((r) => r.valid)
    .map((r) => ({
      type: r.credentialType,
      points: r.trustPoints,
      badge: r.badge,
    }));

  // Take the highest scoring credential as base
  const highestCredScore = credentialScores.length > 0
    ? Math.max(...credentialScores.map((c) => c.points))
    : 0;

  // Multi-credential bonus
  let multiBonus = 0;
  const validCount = credentialScores.length;
  if (validCount >= 4) multiBonus = 15;
  else if (validCount >= 3) multiBonus = 10;
  else if (validCount >= 2) multiBonus = 5;

  const total = Math.min(100, baseScore + highestCredScore + multiBonus);

  return {
    baseScore,
    credentialScores,
    multiCredentialBonus: multiBonus,
    totalScore: total,
  };
}

// Mock funnel analytics
export const mockFunnelMetrics: FunnelMetrics = {
  sessions: 12847,
  steps: [
    { name: "Intro Carousel", count: 12847, dropoffRate: 0, skipped: 0, avgDuration: 8200 },
    { name: "Welcome", count: 11562, dropoffRate: 10, skipped: 0, avgDuration: 4100 },
    { name: "Phone Entry", count: 10406, dropoffRate: 10, skipped: 0, avgDuration: 12500 },
    { name: "OTP Verify", count: 9886, dropoffRate: 5, skipped: 0, avgDuration: 18700 },
    { name: "Passkey Setup", count: 8897, dropoffRate: 10, skipped: 1241, avgDuration: 6300 },
    { name: "Recovery Method", count: 8452, dropoffRate: 5, skipped: 2890, avgDuration: 9800 },
    { name: "Profile Setup", count: 8029, dropoffRate: 5, skipped: 845, avgDuration: 22400 },
    { name: "Credential Verify", count: 7226, dropoffRate: 10, skipped: 3421, avgDuration: 35200 },
    { name: "Trust Dashboard", count: 6864, dropoffRate: 5, skipped: 0, avgDuration: 11600 },
    { name: "Completed", count: 6521, dropoffRate: 5, skipped: 0, avgDuration: 0 },
  ],
  conversionRate: 50.8,
  avgSessionDuration: 128800,
  credentialStats: [
    { type: "passport", attempts: 2841, successes: 2698, failures: 143 },
    { type: "drivers_license", attempts: 3102, successes: 2945, failures: 157 },
    { type: "bank_account", attempts: 1856, successes: 1721, failures: 135 },
    { type: "education", attempts: 987, successes: 924, failures: 63 },
    { type: "phone", attempts: 5421, successes: 5189, failures: 232 },
    { type: "email", attempts: 4102, successes: 3987, failures: 115 },
  ],
  issuerDistribution: [
    { issuer: "US DMV", count: 4210 },
    { issuer: "Wells Fargo", count: 1856 },
    { issuer: "UK HMPO", count: 1733 },
    { issuer: "Stanford University", count: 987 },
    { issuer: "Revolut", count: 821 },
    { issuer: "Twilio", count: 5421 },
  ],
};
