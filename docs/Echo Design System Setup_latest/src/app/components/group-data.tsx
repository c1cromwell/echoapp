// Group types and mock data matching the Go backend models

export type GroupType = "public" | "private" | "secret";
export type GroupRole = "owner" | "admin" | "moderator" | "verified_member" | "member" | "restricted" | "pending";
export type TrustLevel = "unverified" | "newcomer" | "member" | "trusted" | "verified";
export type GroupCategory = "social" | "professional" | "education" | "community" | "marketplace" | "governance" | "support" | "other";
export type ApprovalMode = "auto" | "manual" | "vote";

export interface GroupMember {
  id: string;
  name: string;
  role: GroupRole;
  trustScore: number;
  trustLevel: TrustLevel;
  isVerified: boolean;
  online: boolean;
  gradient: "primary" | "secondary" | "warm" | "cool";
  joinedAt: string;
  isMuted?: boolean;
  isBanned?: boolean;
  warnings?: number;
}

export interface GroupStats {
  verifiedPercentage: number;
  averageTrustScore: number;
  activityLevel: "low" | "moderate" | "high" | "very_high";
  securityRating: "A" | "B" | "C" | "D" | "F";
  totalMessages: number;
  activeMembers: number;
}

export interface Group {
  id: string;
  name: string;
  description: string;
  type: GroupType;
  category: GroupCategory;
  memberCount: number;
  maxMembers: number;
  lastMessage: string;
  lastMessageTime: string;
  unread: number;
  tags: string[];
  rules: string[];
  members: GroupMember[];
  stats: GroupStats;
  createdAt: string;
  approvalMode: ApprovalMode;
  minTrustScore: number;
  requiredBadges: string[];
  avatarGradient: "primary" | "secondary" | "warm" | "cool";
}

export const roleLabels: Record<GroupRole, string> = {
  owner: "Owner",
  admin: "Admin",
  moderator: "Moderator",
  verified_member: "Verified",
  member: "Member",
  restricted: "Restricted",
  pending: "Pending",
};

export const roleColors: Record<GroupRole, string> = {
  owner: "#F59E0B",
  admin: "#EF4444",
  moderator: "#3B82F6",
  verified_member: "#10B981",
  member: "var(--echo-text-tertiary)",
  restricted: "#A1A1AA",
  pending: "#D4D4D8",
};

export const categoryIcons: Record<GroupCategory, string> = {
  social: "🌐",
  professional: "💼",
  education: "📚",
  community: "🏘️",
  marketplace: "🛒",
  governance: "🏛️",
  support: "🤝",
  other: "📌",
};

export const trustLevelLabels: Record<TrustLevel, string> = {
  unverified: "Unverified",
  newcomer: "Newcomer",
  member: "Member",
  trusted: "Trusted",
  verified: "Verified",
};

// Mock data
export const mockGroups: Group[] = [
  {
    id: "g1",
    name: "Tech Enthusiasts",
    description: "A community for discussing the latest in technology, AI, and software development. Share ideas, ask questions, and connect with fellow tech lovers.",
    type: "public",
    category: "professional",
    memberCount: 87,
    maxMembers: 100,
    lastMessage: "Alex: Just shipped the new feature! 🚀",
    lastMessageTime: "5m",
    unread: 4,
    tags: ["tech", "programming", "AI"],
    rules: [
      "Be respectful and constructive",
      "No spam or self-promotion",
      "Stay on topic",
      "No sharing of confidential information",
    ],
    approvalMode: "auto",
    minTrustScore: 30,
    requiredBadges: [],
    avatarGradient: "primary",
    createdAt: "2025-09-15",
    stats: {
      verifiedPercentage: 72,
      averageTrustScore: 78,
      activityLevel: "very_high",
      securityRating: "A",
      totalMessages: 12450,
      activeMembers: 64,
    },
    members: [
      { id: "u1", name: "Sarah Chen", role: "owner", trustScore: 95, trustLevel: "verified", isVerified: true, online: true, gradient: "primary", joinedAt: "2025-09-15" },
      { id: "u2", name: "Marcus Johnson", role: "admin", trustScore: 88, trustLevel: "trusted", isVerified: true, online: true, gradient: "secondary", joinedAt: "2025-09-16" },
      { id: "u3", name: "Elena Rodriguez", role: "moderator", trustScore: 91, trustLevel: "verified", isVerified: true, online: false, gradient: "primary", joinedAt: "2025-09-18" },
      { id: "u4", name: "James Park", role: "verified_member", trustScore: 72, trustLevel: "trusted", isVerified: true, online: false, gradient: "warm", joinedAt: "2025-10-01" },
      { id: "u5", name: "Chase Williams", role: "member", trustScore: 65, trustLevel: "member", isVerified: false, online: true, gradient: "cool", joinedAt: "2025-10-15" },
      { id: "u6", name: "Ava Thompson", role: "member", trustScore: 83, trustLevel: "trusted", isVerified: false, online: false, gradient: "secondary", joinedAt: "2025-11-01" },
      { id: "u7", name: "Noah Kim", role: "pending", trustScore: 28, trustLevel: "newcomer", isVerified: false, online: true, gradient: "warm", joinedAt: "2026-02-28" },
    ],
  },
  {
    id: "g2",
    name: "Design Guild",
    description: "Exclusive community for verified designers to share work, critique, and collaborate on projects.",
    type: "private",
    category: "professional",
    memberCount: 34,
    maxMembers: 50,
    lastMessage: "Mia: New design system is live!",
    lastMessageTime: "1h",
    unread: 0,
    tags: ["design", "UI/UX", "figma"],
    rules: [
      "Share constructive feedback only",
      "Credit original creators",
      "No AI-generated work without disclosure",
    ],
    approvalMode: "manual",
    minTrustScore: 60,
    requiredBadges: ["designer"],
    avatarGradient: "cool",
    createdAt: "2025-11-02",
    stats: {
      verifiedPercentage: 88,
      averageTrustScore: 85,
      activityLevel: "high",
      securityRating: "A",
      totalMessages: 5890,
      activeMembers: 28,
    },
    members: [
      { id: "u8", name: "Mia Davis", role: "owner", trustScore: 94, trustLevel: "verified", isVerified: true, online: false, gradient: "cool", joinedAt: "2025-11-02" },
      { id: "u1", name: "Sarah Chen", role: "admin", trustScore: 95, trustLevel: "verified", isVerified: true, online: true, gradient: "primary", joinedAt: "2025-11-03" },
      { id: "u9", name: "Liam Wilson", role: "verified_member", trustScore: 82, trustLevel: "trusted", isVerified: true, online: true, gradient: "primary", joinedAt: "2025-11-10" },
    ],
  },
  {
    id: "g3",
    name: "Crypto & Web3",
    description: "Deep dives into blockchain, DeFi, and the decentralized web. Verified members only.",
    type: "private",
    category: "community",
    memberCount: 156,
    maxMembers: 500,
    lastMessage: "Raj: ETH merge was a game changer",
    lastMessageTime: "3h",
    unread: 23,
    tags: ["crypto", "web3", "blockchain", "DeFi"],
    rules: [
      "No financial advice — DYOR",
      "No shilling or pump-and-dump schemes",
      "Respect privacy",
    ],
    approvalMode: "auto",
    minTrustScore: 50,
    requiredBadges: [],
    avatarGradient: "warm",
    createdAt: "2025-08-01",
    stats: {
      verifiedPercentage: 61,
      averageTrustScore: 71,
      activityLevel: "high",
      securityRating: "B",
      totalMessages: 28900,
      activeMembers: 98,
    },
    members: [
      { id: "u10", name: "Raj Patel", role: "owner", trustScore: 92, trustLevel: "verified", isVerified: true, online: true, gradient: "warm", joinedAt: "2025-08-01" },
      { id: "u5", name: "Chase Williams", role: "moderator", trustScore: 65, trustLevel: "member", isVerified: false, online: true, gradient: "cool", joinedAt: "2025-08-15" },
    ],
  },
  {
    id: "g4",
    name: "Local Runners Club",
    description: "Weekly runs, race coordination, and fitness tips for our city running community.",
    type: "public",
    category: "social",
    memberCount: 45,
    maxMembers: 100,
    lastMessage: "See everyone Saturday at 7am! 🏃",
    lastMessageTime: "1d",
    unread: 0,
    tags: ["running", "fitness", "local"],
    rules: ["Be encouraging", "Share route safety tips"],
    approvalMode: "auto",
    minTrustScore: 0,
    requiredBadges: [],
    avatarGradient: "secondary",
    createdAt: "2025-12-10",
    stats: {
      verifiedPercentage: 42,
      averageTrustScore: 58,
      activityLevel: "moderate",
      securityRating: "C",
      totalMessages: 2100,
      activeMembers: 22,
    },
    members: [
      { id: "u4", name: "James Park", role: "owner", trustScore: 72, trustLevel: "trusted", isVerified: true, online: false, gradient: "warm", joinedAt: "2025-12-10" },
    ],
  },
  {
    id: "g5",
    name: "Inner Circle",
    description: "Invite-only group for close friends. What happens here stays here.",
    type: "secret",
    category: "social",
    memberCount: 8,
    maxMembers: 20,
    lastMessage: "You: Anyone up for dinner Friday?",
    lastMessageTime: "2h",
    unread: 0,
    tags: [],
    rules: [],
    approvalMode: "manual",
    minTrustScore: 80,
    requiredBadges: [],
    avatarGradient: "primary",
    createdAt: "2026-01-05",
    stats: {
      verifiedPercentage: 100,
      averageTrustScore: 91,
      activityLevel: "moderate",
      securityRating: "A",
      totalMessages: 890,
      activeMembers: 8,
    },
    members: [
      { id: "u_me", name: "You", role: "owner", trustScore: 95, trustLevel: "verified", isVerified: true, online: true, gradient: "primary", joinedAt: "2026-01-05" },
      { id: "u1", name: "Sarah Chen", role: "admin", trustScore: 95, trustLevel: "verified", isVerified: true, online: true, gradient: "primary", joinedAt: "2026-01-05" },
      { id: "u2", name: "Marcus Johnson", role: "member", trustScore: 88, trustLevel: "trusted", isVerified: true, online: true, gradient: "secondary", joinedAt: "2026-01-06" },
    ],
  },
];

export const discoverableGroups: Group[] = [
  {
    id: "gd1",
    name: "Photography Club",
    description: "Share your best shots and learn from fellow photographers.",
    type: "public",
    category: "social",
    memberCount: 234,
    maxMembers: 500,
    lastMessage: "",
    lastMessageTime: "",
    unread: 0,
    tags: ["photography", "art", "creative"],
    rules: [],
    approvalMode: "auto",
    minTrustScore: 10,
    requiredBadges: [],
    avatarGradient: "warm",
    createdAt: "2025-06-01",
    stats: { verifiedPercentage: 55, averageTrustScore: 67, activityLevel: "high", securityRating: "B", totalMessages: 15000, activeMembers: 120 },
    members: [],
  },
  {
    id: "gd2",
    name: "Startup Founders",
    description: "Connect with other founders, share experiences, and find co-founders.",
    type: "private",
    category: "professional",
    memberCount: 89,
    maxMembers: 200,
    lastMessage: "",
    lastMessageTime: "",
    unread: 0,
    tags: ["startups", "business", "entrepreneurship"],
    rules: [],
    approvalMode: "manual",
    minTrustScore: 50,
    requiredBadges: ["founder"],
    avatarGradient: "cool",
    createdAt: "2025-07-20",
    stats: { verifiedPercentage: 78, averageTrustScore: 82, activityLevel: "high", securityRating: "A", totalMessages: 9200, activeMembers: 65 },
    members: [],
  },
];
