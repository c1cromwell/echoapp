// Trust Network types and mock data based on the Dynamic Trust Network Blueprint

export type TrustLevel = "unverified" | "newcomer" | "member" | "trusted" | "verified";

export interface TrustScoreComponents {
  verification: number; // 0-30 (30% weight)
  network: number;      // 0-25 (25% weight)
  behavior: number;     // 0-25 (25% weight)
  transactions: number; // 0-20 (20% weight)
}

export interface TrustHistoryPoint {
  date: string;
  score: number;
  event?: string;
}

export interface TrustCircle {
  id: string;
  name: string;
  level: "inner" | "trusted" | "known" | "public";
  maxSize: number;
  currentSize: number;
  permissions: string[];
  contacts: TrustContact[];
}

export interface TrustContact {
  id: string;
  name: string;
  trustScore: number;
  trustLevel: TrustLevel;
  gradient: string;
  circleLevel: "inner" | "trusted" | "known";
  endorsedYou?: boolean;
  endorsedBy?: string[];
}

export interface Endorsement {
  id: string;
  fromUserId: string;
  fromUserName: string;
  fromTrustScore: number;
  fromGradient: string;
  category: EndorsementCategory;
  timestamp: string;
  weight: number;
}

export type EndorsementCategory =
  | "reliable"
  | "responsive"
  | "professional"
  | "knowledgeable"
  | "helpful"
  | "trustworthy";

export interface Dispute {
  id: string;
  type: "false_report" | "coordinated_attack" | "system_error" | "account_compromise";
  status: "pending" | "reviewing" | "resolved" | "rejected";
  filedAt: string;
  resolvedAt?: string;
  description: string;
  jurorCount: number;
  votesFor: number;
  votesAgainst: number;
}

export interface VerificationMethod {
  id: string;
  method: string;
  points: number;
  icon: string;
  completed: boolean;
  completedAt?: string;
}

export interface FeatureGate {
  feature: string;
  requiredLevel: TrustLevel;
  currentAccess: boolean;
  description: string;
}

// ===== Constants =====

export const trustLevelConfig: Record<TrustLevel, {
  label: string;
  range: [number, number];
  color: string;
  bgColor: string;
  description: string;
}> = {
  unverified: {
    label: "Unverified",
    range: [0, 19],
    color: "#94A3B8",
    bgColor: "#F1F5F9",
    description: "Basic messaging only",
  },
  newcomer: {
    label: "Newcomer",
    range: [20, 39],
    color: "#F59E0B",
    bgColor: "#FFFBEB",
    description: "Small groups allowed",
  },
  member: {
    label: "Member",
    range: [40, 59],
    color: "#3B82F6",
    bgColor: "#EFF6FF",
    description: "Standard features",
  },
  trusted: {
    label: "Trusted",
    range: [60, 79],
    color: "#8B5CF6",
    bgColor: "#F5F3FF",
    description: "Can endorse others",
  },
  verified: {
    label: "Verified",
    range: [80, 100],
    color: "#10B981",
    bgColor: "#ECFDF5",
    description: "Priority features",
  },
};

export const endorsementCategoryInfo: Record<EndorsementCategory, {
  label: string;
  emoji: string;
  color: string;
}> = {
  reliable: { label: "Reliable", emoji: "🎯", color: "#3B82F6" },
  responsive: { label: "Responsive", emoji: "⚡", color: "#F59E0B" },
  professional: { label: "Professional", emoji: "💼", color: "#8B5CF6" },
  knowledgeable: { label: "Knowledgeable", emoji: "🧠", color: "#06B6D4" },
  helpful: { label: "Helpful", emoji: "🤝", color: "#10B981" },
  trustworthy: { label: "Trustworthy", emoji: "🛡️", color: "#F43F5E" },
};

export const circleLevelConfig: Record<string, {
  label: string;
  color: string;
  bgColor: string;
  maxSize: number;
  permissions: string[];
}> = {
  inner: {
    label: "Inner Circle",
    color: "#F43F5E",
    bgColor: "#FFF1F2",
    maxSize: 15,
    permissions: ["Real-time status", "Direct calls", "Full profile", "Location sharing"],
  },
  trusted: {
    label: "Trusted",
    color: "#8B5CF6",
    bgColor: "#F5F3FF",
    maxSize: 50,
    permissions: ["Online status", "Call requests", "Standard profile"],
  },
  known: {
    label: "Known",
    color: "#3B82F6",
    bgColor: "#EFF6FF",
    maxSize: 200,
    permissions: ["Approximate status", "Filtered messages"],
  },
  public: {
    label: "Public",
    color: "#94A3B8",
    bgColor: "#F1F5F9",
    maxSize: -1,
    permissions: ["Minimal profile", "Strict message filtering"],
  },
};

// ===== Helpers =====

export function getTrustLevel(score: number): TrustLevel {
  if (score >= 80) return "verified";
  if (score >= 60) return "trusted";
  if (score >= 40) return "member";
  if (score >= 20) return "newcomer";
  return "unverified";
}

export function getDecayInfo(lastActiveDate: string): { daysInactive: number; decayPercent: number } {
  const now = new Date("2026-03-04");
  const lastActive = new Date(lastActiveDate);
  const diffDays = Math.floor((now.getTime() - lastActive.getTime()) / (1000 * 60 * 60 * 24));
  return {
    daysInactive: diffDays,
    decayPercent: Math.min(40, diffDays * 0.2), // 0.2% daily, floor at 60% of peak
  };
}

// ===== Mock Data =====

export const mockTrustComponents: TrustScoreComponents = {
  verification: 26, // out of 30
  network: 22,      // out of 25
  behavior: 21,     // out of 25
  transactions: 18, // out of 20
};

export const mockTrustHistory: TrustHistoryPoint[] = [
  { date: "Dec 5", score: 42, event: "Account created" },
  { date: "Dec 12", score: 48 },
  { date: "Dec 19", score: 52, event: "Phone verified" },
  { date: "Dec 26", score: 55 },
  { date: "Jan 2", score: 58 },
  { date: "Jan 9", score: 62, event: "Passport credential" },
  { date: "Jan 16", score: 65 },
  { date: "Jan 23", score: 68, event: "Bank account verified" },
  { date: "Jan 30", score: 71 },
  { date: "Feb 6", score: 74, event: "First endorsement received" },
  { date: "Feb 13", score: 78 },
  { date: "Feb 20", score: 81, event: "Education credential" },
  { date: "Feb 27", score: 84 },
  { date: "Mar 3", score: 87, event: "Multi-credential bonus" },
];

export const mockTrustContacts: TrustContact[] = [
  { id: "c1", name: "Sarah Chen", trustScore: 92, trustLevel: "verified", gradient: "primary", circleLevel: "inner", endorsedYou: true, endorsedBy: ["reliable", "responsive"] },
  { id: "c2", name: "Marcus Johnson", trustScore: 78, trustLevel: "trusted", gradient: "secondary", circleLevel: "inner", endorsedYou: true, endorsedBy: ["professional"] },
  { id: "c3", name: "Elena Rodriguez", trustScore: 85, trustLevel: "verified", gradient: "warm", circleLevel: "inner", endorsedYou: false },
  { id: "c4", name: "James Park", trustScore: 72, trustLevel: "trusted", gradient: "warm", circleLevel: "trusted" },
  { id: "c5", name: "Mia Davis", trustScore: 88, trustLevel: "verified", gradient: "cool", circleLevel: "trusted", endorsedYou: true, endorsedBy: ["knowledgeable"] },
  { id: "c6", name: "Chase Williams", trustScore: 65, trustLevel: "trusted", gradient: "cool", circleLevel: "trusted" },
  { id: "c7", name: "Ava Thompson", trustScore: 58, trustLevel: "member", gradient: "secondary", circleLevel: "known" },
  { id: "c8", name: "Noah Kim", trustScore: 91, trustLevel: "verified", gradient: "primary", circleLevel: "known", endorsedYou: true, endorsedBy: ["trustworthy"] },
  { id: "c9", name: "Raj Patel", trustScore: 76, trustLevel: "trusted", gradient: "warm", circleLevel: "known" },
  { id: "c10", name: "Liam Chen", trustScore: 44, trustLevel: "member", gradient: "secondary", circleLevel: "known" },
];

export const mockTrustCircles: TrustCircle[] = [
  {
    id: "circle-inner",
    name: "Inner Circle",
    level: "inner",
    maxSize: 15,
    currentSize: 3,
    permissions: circleLevelConfig.inner.permissions,
    contacts: mockTrustContacts.filter((c) => c.circleLevel === "inner"),
  },
  {
    id: "circle-trusted",
    name: "Trusted",
    level: "trusted",
    maxSize: 50,
    currentSize: 3,
    permissions: circleLevelConfig.trusted.permissions,
    contacts: mockTrustContacts.filter((c) => c.circleLevel === "trusted"),
  },
  {
    id: "circle-known",
    name: "Known",
    level: "known",
    maxSize: 200,
    currentSize: 4,
    permissions: circleLevelConfig.known.permissions,
    contacts: mockTrustContacts.filter((c) => c.circleLevel === "known"),
  },
];

export const mockEndorsements: Endorsement[] = [
  { id: "e1", fromUserId: "c1", fromUserName: "Sarah Chen", fromTrustScore: 92, fromGradient: "primary", category: "reliable", timestamp: "2026-03-02T10:30:00Z", weight: 1.5 },
  { id: "e2", fromUserId: "c1", fromUserName: "Sarah Chen", fromTrustScore: 92, fromGradient: "primary", category: "responsive", timestamp: "2026-03-02T10:30:00Z", weight: 1.5 },
  { id: "e3", fromUserId: "c2", fromUserName: "Marcus Johnson", fromTrustScore: 78, fromGradient: "secondary", category: "professional", timestamp: "2026-02-28T14:20:00Z", weight: 1.0 },
  { id: "e4", fromUserId: "c5", fromUserName: "Mia Davis", fromTrustScore: 88, fromGradient: "cool", category: "knowledgeable", timestamp: "2026-02-25T09:15:00Z", weight: 1.5 },
  { id: "e5", fromUserId: "c8", fromUserName: "Noah Kim", fromTrustScore: 91, fromGradient: "primary", category: "trustworthy", timestamp: "2026-02-20T16:45:00Z", weight: 1.5 },
];

export const mockVerificationMethods: VerificationMethod[] = [
  { id: "v1", method: "Email", points: 10, icon: "📧", completed: true, completedAt: "2025-12-05" },
  { id: "v2", method: "Phone (SMS)", points: 15, icon: "📱", completed: true, completedAt: "2025-12-19" },
  { id: "v3", method: "Government ID", points: 40, icon: "🛂", completed: true, completedAt: "2026-01-09" },
  { id: "v4", method: "Social (GitHub)", points: 5, icon: "🔗", completed: true, completedAt: "2026-01-15" },
  { id: "v5", method: "Social (LinkedIn)", points: 5, icon: "🔗", completed: true, completedAt: "2026-01-15" },
  { id: "v6", method: "Biometric", points: 15, icon: "🔐", completed: true, completedAt: "2026-01-20" },
  { id: "v7", method: "Phone (Carrier)", points: 20, icon: "📡", completed: false },
  { id: "v8", method: "Social (Twitter)", points: 5, icon: "🔗", completed: false },
];

export const mockFeatureGates: FeatureGate[] = [
  { feature: "Basic Messaging", requiredLevel: "unverified", currentAccess: true, description: "Send and receive messages" },
  { feature: "Small Groups (10)", requiredLevel: "newcomer", currentAccess: true, description: "Create groups up to 10 members" },
  { feature: "Standard Groups (50)", requiredLevel: "member", currentAccess: true, description: "Create groups up to 50 members" },
  { feature: "Endorse Others", requiredLevel: "trusted", currentAccess: true, description: "Give endorsements to contacts" },
  { feature: "Large Groups (200)", requiredLevel: "trusted", currentAccess: true, description: "Create groups up to 200 members" },
  { feature: "Priority Features", requiredLevel: "verified", currentAccess: true, description: "Unlimited groups, priority support" },
  { feature: "Call Anyone", requiredLevel: "trusted", currentAccess: true, description: "Call users outside your contacts" },
  { feature: "Broadcast Channels", requiredLevel: "member", currentAccess: true, description: "Create broadcast channels" },
];

export const mockDisputes: Dispute[] = [
  {
    id: "d1",
    type: "false_report",
    status: "resolved",
    filedAt: "2026-01-15T10:00:00Z",
    resolvedAt: "2026-01-18T14:30:00Z",
    description: "False spam report filed by unknown user",
    jurorCount: 5,
    votesFor: 4,
    votesAgainst: 1,
  },
];

// Rate limit status
export const mockRateLimits = {
  endorsementsGiven: { used: 2, limit: 5, resetsIn: "8h" },
  circlePromotions: { used: 1, limit: 10, resetsIn: "8h" },
  reportsFiled: { used: 0, limit: 10, resetsIn: "8h" },
  verificationAttempts: { used: 0, limit: 3, resetsIn: "24h" },
  disputesFiled: { used: 0, limit: 1, resetsIn: "78d" },
};
