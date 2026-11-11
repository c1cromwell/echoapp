// Broadcast Channel types and mock data matching the Go backend spec

export type VisibilityMode = "public" | "semi_private" | "private";
export type ChannelType = "news" | "announcement" | "educational" | "community";
export type VerificationStatus = "unverified" | "verified" | "premium";
export type PublishStatus = "draft" | "scheduled" | "published" | "archived" | "deleted";
export type ContentType = "text" | "image" | "video" | "poll" | "file" | "mixed";
export type SubscriberRole = "subscriber" | "moderator" | "admin";
export type SubscriptionTier = "free" | "bronze" | "silver" | "gold";
export type NotificationMode = "all" | "important" | "none";

export interface BroadcastChannel {
  id: string;
  creatorId: string;
  creatorName: string;
  name: string;
  description: string;
  topic: string;
  visibilityMode: VisibilityMode;
  channelType: ChannelType;
  coverGradient: "primary" | "secondary" | "warm" | "cool" | "accent";
  tags: string[];
  isActive: boolean;
  subscriberCount: number;
  totalPostCount: number;
  trustScore: number;
  verificationStatus: VerificationStatus;
  allowComments: boolean;
  allowPolls: boolean;
  allowMedia: boolean;
  language: string;
  createdAt: string;
  lastPostAt?: string;
  // User's relationship
  isSubscribed?: boolean;
  userRole?: SubscriberRole;
  userTier?: SubscriptionTier;
  notificationMode?: NotificationMode;
  isMutedByUser?: boolean;
}

export interface ChannelPost {
  id: string;
  channelId: string;
  creatorId: string;
  creatorName: string;
  content: string;
  contentType: ContentType;
  publishStatus: PublishStatus;
  publishedAt: string;
  likeCount: number;
  commentCount: number;
  shareCount: number;
  isPinned: boolean;
  isFeatured: boolean;
  isSponsored: boolean;
  allowReplies: boolean;
  hasImage?: boolean;
  imageUrl?: string;
  poll?: {
    question: string;
    options: { label: string; votes: number }[];
    multiSelect: boolean;
    expiresAt: string;
    totalVotes: number;
    userVoted?: number;
  };
  isLikedByUser?: boolean;
}

export interface ChannelAnalyticsData {
  totalSubscribers: number;
  newSubscribers: number;
  unsubscribeCount: number;
  viewCount: number;
  likeCount: number;
  commentCount: number;
  shareCount: number;
  postCount: number;
  averageEngagement: number;
  totalRevenue: number;
  subscriberGrowth: { date: string; count: number }[];
  engagementByDay: { date: string; likes: number; comments: number; shares: number }[];
  topPosts: { id: string; title: string; engagement: number }[];
  tierBreakdown: { tier: string; count: number; percentage: number }[];
  topRegions: { region: string; count: number }[];
}

export const channelTypeLabels: Record<ChannelType, string> = {
  news: "News",
  announcement: "Announcement",
  educational: "Educational",
  community: "Community",
};

export const channelTypeIcons: Record<ChannelType, string> = {
  news: "📰",
  announcement: "📢",
  educational: "📚",
  community: "💬",
};

export const visibilityLabels: Record<VisibilityMode, string> = {
  public: "Public",
  semi_private: "Semi-Private",
  private: "Private",
};

export const tierPricing: Record<SubscriptionTier, { label: string; price: string; features: string[] }> = {
  free: { label: "Free", price: "$0", features: ["Basic access", "View all public posts"] },
  bronze: { label: "Bronze", price: "$4.99/mo", features: ["Ad-free experience", "Basic access"] },
  silver: { label: "Silver", price: "$9.99/mo", features: ["Ad-free", "Exclusive posts", "Priority notifications"] },
  gold: { label: "Gold", price: "$19.99/mo", features: ["All Silver features", "Direct message creator", "Early access"] },
};

// Cover gradient mappings
export const coverGradients: Record<string, string> = {
  primary: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
  secondary: "linear-gradient(135deg, #06B6D4, #10B981)",
  warm: "linear-gradient(135deg, #F43F5E, #F59E0B)",
  cool: "linear-gradient(135deg, #3B82F6, #8B5CF6)",
  accent: "linear-gradient(135deg, var(--echo-accent), var(--echo-primary))",
};

// ===== Mock Data =====

export const mockChannels: BroadcastChannel[] = [
  {
    id: "ch1",
    creatorId: "u1",
    creatorName: "Sarah Chen",
    name: "Tech Pulse Daily",
    description: "Breaking tech news, AI developments, and startup coverage. Curated by industry insiders with verified sources.",
    topic: "Technology",
    visibilityMode: "public",
    channelType: "news",
    coverGradient: "primary",
    tags: ["tech", "AI", "startups", "news"],
    isActive: true,
    subscriberCount: 12480,
    totalPostCount: 847,
    trustScore: 92,
    verificationStatus: "verified",
    allowComments: true,
    allowPolls: true,
    allowMedia: true,
    language: "en",
    createdAt: "2025-06-15",
    lastPostAt: "2026-03-03T10:30:00Z",
    isSubscribed: true,
    userRole: "subscriber",
    userTier: "free",
    notificationMode: "important",
  },
  {
    id: "ch2",
    creatorId: "u8",
    creatorName: "Mia Davis",
    name: "Design Systems Weekly",
    description: "Deep dives into design systems, component libraries, and UI/UX best practices. Premium content for working designers.",
    topic: "Design",
    visibilityMode: "public",
    channelType: "educational",
    coverGradient: "cool",
    tags: ["design", "UI/UX", "figma", "design-systems"],
    isActive: true,
    subscriberCount: 5230,
    totalPostCount: 156,
    trustScore: 88,
    verificationStatus: "verified",
    allowComments: true,
    allowPolls: true,
    allowMedia: true,
    language: "en",
    createdAt: "2025-08-20",
    lastPostAt: "2026-03-02T16:00:00Z",
    isSubscribed: true,
    userRole: "subscriber",
    userTier: "silver",
    notificationMode: "all",
  },
  {
    id: "ch3",
    creatorId: "u10",
    creatorName: "Raj Patel",
    name: "Crypto Alpha",
    description: "Market analysis, DeFi protocols, and on-chain insights. Verified signals only, no speculation.",
    topic: "Crypto",
    visibilityMode: "semi_private",
    channelType: "community",
    coverGradient: "warm",
    tags: ["crypto", "DeFi", "web3", "trading"],
    isActive: true,
    subscriberCount: 8920,
    totalPostCount: 432,
    trustScore: 76,
    verificationStatus: "verified",
    allowComments: true,
    allowPolls: true,
    allowMedia: true,
    language: "en",
    createdAt: "2025-07-01",
    lastPostAt: "2026-03-03T08:15:00Z",
    isSubscribed: true,
    userRole: "moderator",
    userTier: "gold",
    notificationMode: "all",
  },
  {
    id: "ch4",
    creatorId: "u_me",
    creatorName: "You",
    name: "Product Launch Notes",
    description: "Official announcements, release notes, and roadmap updates for our product suite.",
    topic: "Product",
    visibilityMode: "public",
    channelType: "announcement",
    coverGradient: "secondary",
    tags: ["product", "releases", "updates"],
    isActive: true,
    subscriberCount: 1845,
    totalPostCount: 67,
    trustScore: 85,
    verificationStatus: "verified",
    allowComments: false,
    allowPolls: false,
    allowMedia: true,
    language: "en",
    createdAt: "2025-11-10",
    lastPostAt: "2026-02-28T14:00:00Z",
    isSubscribed: true,
    userRole: "admin",
    userTier: "free",
    notificationMode: "all",
  },
];

export const discoverableChannels: BroadcastChannel[] = [
  {
    id: "chd1",
    creatorId: "ux1",
    creatorName: "Elena Rodriguez",
    name: "Learn Rust in 30 Days",
    description: "Structured Rust programming course with daily lessons, exercises, and community support.",
    topic: "Programming",
    visibilityMode: "public",
    channelType: "educational",
    coverGradient: "accent",
    tags: ["rust", "programming", "learning", "course"],
    isActive: true,
    subscriberCount: 3450,
    totalPostCount: 90,
    trustScore: 91,
    verificationStatus: "verified",
    allowComments: true,
    allowPolls: true,
    allowMedia: true,
    language: "en",
    createdAt: "2025-12-01",
    lastPostAt: "2026-03-03T06:00:00Z",
  },
  {
    id: "chd2",
    creatorId: "ux2",
    creatorName: "James Park",
    name: "Fitness & Nutrition Hub",
    description: "Evidence-based fitness tips, meal plans, and workout routines from certified trainers.",
    topic: "Health",
    visibilityMode: "public",
    channelType: "community",
    coverGradient: "warm",
    tags: ["fitness", "nutrition", "health", "workout"],
    isActive: true,
    subscriberCount: 7800,
    totalPostCount: 234,
    trustScore: 82,
    verificationStatus: "verified",
    allowComments: true,
    allowPolls: true,
    allowMedia: true,
    language: "en",
    createdAt: "2025-09-15",
    lastPostAt: "2026-03-02T18:30:00Z",
  },
  {
    id: "chd3",
    creatorId: "ux3",
    creatorName: "Marcus Johnson",
    name: "Startup Weekly Digest",
    description: "Curated startup news, funding rounds, and founder interviews. Semi-private with approval for quality control.",
    topic: "Business",
    visibilityMode: "semi_private",
    channelType: "news",
    coverGradient: "cool",
    tags: ["startups", "business", "VC", "funding"],
    isActive: true,
    subscriberCount: 4200,
    totalPostCount: 189,
    trustScore: 87,
    verificationStatus: "premium",
    allowComments: true,
    allowPolls: false,
    allowMedia: true,
    language: "en",
    createdAt: "2025-10-01",
    lastPostAt: "2026-03-01T12:00:00Z",
  },
  {
    id: "chd4",
    creatorId: "ux4",
    creatorName: "Ava Thompson",
    name: "Travel Stories",
    description: "Real travel experiences, hidden gems, and budget tips from verified travelers worldwide.",
    topic: "Travel",
    visibilityMode: "public",
    channelType: "community",
    coverGradient: "secondary",
    tags: ["travel", "adventure", "photography", "budget"],
    isActive: true,
    subscriberCount: 15600,
    totalPostCount: 512,
    trustScore: 79,
    verificationStatus: "verified",
    allowComments: true,
    allowPolls: true,
    allowMedia: true,
    language: "en",
    createdAt: "2025-05-20",
    lastPostAt: "2026-03-03T09:45:00Z",
  },
  {
    id: "chd5",
    creatorId: "ux5",
    creatorName: "Noah Kim",
    name: "Privacy & Security Alerts",
    description: "Invite-only channel for critical security advisories and privacy tool recommendations.",
    topic: "Security",
    visibilityMode: "private",
    channelType: "announcement",
    coverGradient: "primary",
    tags: ["security", "privacy", "infosec"],
    isActive: true,
    subscriberCount: 890,
    totalPostCount: 45,
    trustScore: 95,
    verificationStatus: "premium",
    allowComments: false,
    allowPolls: false,
    allowMedia: true,
    language: "en",
    createdAt: "2025-11-01",
    lastPostAt: "2026-03-02T20:00:00Z",
  },
];

export const mockChannelPosts: Record<string, ChannelPost[]> = {
  ch1: [
    {
      id: "p1",
      channelId: "ch1",
      creatorId: "u1",
      creatorName: "Sarah Chen",
      content: "OpenAI just announced GPT-5 with native multimodal reasoning. Early benchmarks show 40% improvement in complex reasoning tasks. Full analysis thread below.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-03T10:30:00Z",
      likeCount: 342,
      commentCount: 89,
      shareCount: 156,
      isPinned: true,
      isFeatured: true,
      isSponsored: false,
      allowReplies: true,
      isLikedByUser: true,
    },
    {
      id: "p2",
      channelId: "ch1",
      creatorId: "u1",
      creatorName: "Sarah Chen",
      content: "Apple Vision Pro 2 leak: lighter design, M4 chip, and a rumored $2,499 starting price. Shipping Q4 2026.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-03T08:15:00Z",
      likeCount: 218,
      commentCount: 67,
      shareCount: 94,
      isPinned: false,
      isFeatured: false,
      isSponsored: false,
      allowReplies: true,
    },
    {
      id: "p3",
      channelId: "ch1",
      creatorId: "u1",
      creatorName: "Sarah Chen",
      content: "Which AI development excites you most in 2026?",
      contentType: "poll",
      publishStatus: "published",
      publishedAt: "2026-03-02T16:00:00Z",
      likeCount: 156,
      commentCount: 34,
      shareCount: 22,
      isPinned: false,
      isFeatured: false,
      isSponsored: false,
      allowReplies: true,
      poll: {
        question: "Which AI development excites you most in 2026?",
        options: [
          { label: "Autonomous AI agents", votes: 1240 },
          { label: "Multimodal reasoning", votes: 980 },
          { label: "Open-source models", votes: 1560 },
          { label: "AI in healthcare", votes: 870 },
        ],
        multiSelect: false,
        expiresAt: "2026-03-09T16:00:00Z",
        totalVotes: 4650,
        userVoted: 2,
      },
    },
    {
      id: "p4",
      channelId: "ch1",
      creatorId: "u1",
      creatorName: "Sarah Chen",
      content: "Sponsored: Check out the new developer tools from Vercel - v0 now supports full-stack app generation with database integration.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-02T12:00:00Z",
      likeCount: 89,
      commentCount: 12,
      shareCount: 34,
      isPinned: false,
      isFeatured: false,
      isSponsored: true,
      allowReplies: true,
    },
    {
      id: "p5",
      channelId: "ch1",
      creatorId: "u1",
      creatorName: "Sarah Chen",
      content: "EU AI Act enforcement begins next week. Here's what developers and companies need to know about compliance requirements, risk categories, and timeline.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-01T14:30:00Z",
      likeCount: 456,
      commentCount: 123,
      shareCount: 234,
      isPinned: false,
      isFeatured: true,
      isSponsored: false,
      allowReplies: true,
      isLikedByUser: true,
    },
  ],
  ch2: [
    {
      id: "p6",
      channelId: "ch2",
      creatorId: "u8",
      creatorName: "Mia Davis",
      content: "New lesson: Building Adaptive Color Systems. Learn how to create theme tokens that work across light, dark, and high-contrast modes using CSS custom properties.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-02T16:00:00Z",
      likeCount: 189,
      commentCount: 45,
      shareCount: 67,
      isPinned: true,
      isFeatured: false,
      isSponsored: false,
      allowReplies: true,
    },
    {
      id: "p7",
      channelId: "ch2",
      creatorId: "u8",
      creatorName: "Mia Davis",
      content: "The debate between Tailwind and CSS-in-JS continues, but here's what actually matters for design systems: consistency, maintainability, and developer experience.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-01T10:00:00Z",
      likeCount: 234,
      commentCount: 78,
      shareCount: 45,
      isPinned: false,
      isFeatured: false,
      isSponsored: false,
      allowReplies: true,
      isLikedByUser: true,
    },
  ],
  ch3: [
    {
      id: "p8",
      channelId: "ch3",
      creatorId: "u10",
      creatorName: "Raj Patel",
      content: "ETH just broke $8,000 resistance. On-chain data shows strong accumulation by institutional wallets. Key levels to watch: $8,200 resistance, $7,600 support.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-03-03T08:15:00Z",
      likeCount: 567,
      commentCount: 234,
      shareCount: 189,
      isPinned: false,
      isFeatured: true,
      isSponsored: false,
      allowReplies: true,
    },
  ],
  ch4: [
    {
      id: "p9",
      channelId: "ch4",
      creatorId: "u_me",
      creatorName: "You",
      content: "v2.4.0 Release Notes: Broadcast Channels are now live! Create channels, post content, and build communities. Full documentation available in the help center.",
      contentType: "text",
      publishStatus: "published",
      publishedAt: "2026-02-28T14:00:00Z",
      likeCount: 123,
      commentCount: 34,
      shareCount: 56,
      isPinned: true,
      isFeatured: false,
      isSponsored: false,
      allowReplies: false,
    },
  ],
};

export const mockAnalytics: ChannelAnalyticsData = {
  totalSubscribers: 1845,
  newSubscribers: 127,
  unsubscribeCount: 12,
  viewCount: 45200,
  likeCount: 2340,
  commentCount: 456,
  shareCount: 890,
  postCount: 67,
  averageEngagement: 8.4,
  totalRevenue: 0,
  subscriberGrowth: [
    { date: "Feb 1", count: 1580 },
    { date: "Feb 8", count: 1620 },
    { date: "Feb 15", count: 1690 },
    { date: "Feb 22", count: 1730 },
    { date: "Mar 1", count: 1810 },
    { date: "Mar 3", count: 1845 },
  ],
  engagementByDay: [
    { date: "Mon", likes: 380, comments: 65, shares: 120 },
    { date: "Tue", likes: 420, comments: 78, shares: 145 },
    { date: "Wed", likes: 350, comments: 55, shares: 98 },
    { date: "Thu", likes: 410, comments: 82, shares: 155 },
    { date: "Fri", likes: 290, comments: 48, shares: 87 },
    { date: "Sat", likes: 240, comments: 62, shares: 142 },
    { date: "Sun", likes: 250, comments: 66, shares: 143 },
  ],
  topPosts: [
    { id: "p9", title: "v2.4.0 Release Notes", engagement: 213 },
    { id: "p10", title: "Roadmap Q2 2026", engagement: 189 },
    { id: "p11", title: "Security Update v2.3.1", engagement: 156 },
  ],
  tierBreakdown: [
    { tier: "Free", count: 1520, percentage: 82 },
    { tier: "Bronze", count: 180, percentage: 10 },
    { tier: "Silver", count: 110, percentage: 6 },
    { tier: "Gold", count: 35, percentage: 2 },
  ],
  topRegions: [
    { region: "United States", count: 680 },
    { region: "United Kingdom", count: 245 },
    { region: "Germany", count: 189 },
    { region: "India", count: 156 },
    { region: "Canada", count: 134 },
  ],
};

// Helper to format subscriber counts
export function formatCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toString();
}

// Helper to get time ago string
export function timeAgo(dateStr: string): string {
  const now = new Date("2026-03-03T12:00:00Z");
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  if (diffMins < 60) return `${diffMins}m`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h`;
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 7) return `${diffDays}d`;
  return `${Math.floor(diffDays / 7)}w`;
}
