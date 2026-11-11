// Persona types and mock data matching the Go backend models

export type PersonaCategory =
  | "professional"
  | "personal"
  | "family"
  | "gaming"
  | "dating"
  | "creative"
  | "anonymous"
  | "custom";

export type VisibilityLevel = "everyone" | "contacts" | "nobody";
export type NotificationLevel = "all" | "mentions" | "none";

export interface AccessPermissions {
  canView: boolean;
  canMessage: boolean;
  canCall: boolean;
  canSeeOtherPersonas: boolean;
}

export interface AccessGrant {
  grantId: string;
  personaId: string;
  contactId: string;
  contactName: string;
  permissions: AccessPermissions;
  expiresAt?: string;
  grantedAt: string;
}

export interface PrivacySettings {
  lastSeenVisible: boolean;
  onlineStatusVisible: boolean;
  profilePictureVisible: VisibilityLevel;
  bioVisible: VisibilityLevel;
  statusVisible: VisibilityLevel;
  readReceipts: boolean;
  typingIndicator: boolean;
  searchable: boolean;
  suggestable: boolean;
  contactSharing: boolean;
  crossPersonaLinking: boolean;
  trustScoreSharing: boolean;
}

export interface NotificationSettings {
  enabled: boolean;
  messageLevel: NotificationLevel;
  callLevel: NotificationLevel;
  groupLevel: NotificationLevel;
  sound: boolean;
  vibration: boolean;
  previewContent: boolean;
  showSender: boolean;
  showPersonaName: boolean;
}

export interface FeatureSettings {
  voiceCalls: boolean;
  videoCalls: boolean;
  fileSharing: boolean;
  locationSharing: boolean;
  disappearingMessages: boolean;
  scheduledMessages: boolean;
  maxGroupSize: number;
  maxFileSize: number; // MB
}

export interface Persona {
  personaId: string;
  userId: string;
  displayName: string;
  username: string;
  avatar: string;
  bio: string;
  category: PersonaCategory;
  customCategory?: string;
  isActive: boolean;
  isDefault: boolean;
  createdAt: string;
  gradient: "primary" | "secondary" | "warm" | "cool";
  privacy: PrivacySettings;
  notifications: NotificationSettings;
  features: FeatureSettings;
  grants: AccessGrant[];
  derivationPath: string;
}

export interface MasterIdentity {
  userId: string;
  did: string;
  trustLevel: "unverified" | "newcomer" | "member" | "trusted" | "verified";
  trustScore: number;
  phoneVerified: boolean;
  emailVerified: boolean;
  kycVerified: boolean;
  personaCount: number;
  maxPersonas: number;
}

export const categoryInfo: Record<PersonaCategory, { emoji: string; label: string; description: string; color: string }> = {
  professional: { emoji: "💼", label: "Professional", description: "Work & career context", color: "#3B82F6" },
  personal: { emoji: "🏠", label: "Personal", description: "Private life & friends", color: "#10B981" },
  family: { emoji: "👨‍👩‍👧‍👦", label: "Family", description: "Family interactions", color: "#F59E0B" },
  gaming: { emoji: "🎮", label: "Gaming", description: "Gaming identity", color: "#8B5CF6" },
  dating: { emoji: "💕", label: "Dating", description: "Dating & relationships", color: "#F43F5E" },
  creative: { emoji: "🎨", label: "Creative", description: "Artist & creator identity", color: "#06B6D4" },
  anonymous: { emoji: "🕶️", label: "Anonymous", description: "Privacy-focused identity", color: "#64748B" },
  custom: { emoji: "✨", label: "Custom", description: "User-defined category", color: "#A855F7" },
};

export const trustLevelLimits: Record<string, { maxPersonas: number; customCategories: boolean }> = {
  unverified: { maxPersonas: 2, customCategories: false },
  newcomer: { maxPersonas: 3, customCategories: false },
  member: { maxPersonas: 5, customCategories: true },
  trusted: { maxPersonas: 7, customCategories: true },
  verified: { maxPersonas: 10, customCategories: true },
};

// Default settings
export const defaultPrivacy: PrivacySettings = {
  lastSeenVisible: true,
  onlineStatusVisible: true,
  profilePictureVisible: "everyone",
  bioVisible: "everyone",
  statusVisible: "contacts",
  readReceipts: true,
  typingIndicator: true,
  searchable: true,
  suggestable: true,
  contactSharing: false,
  crossPersonaLinking: false,
  trustScoreSharing: false,
};

export const defaultNotifications: NotificationSettings = {
  enabled: true,
  messageLevel: "all",
  callLevel: "all",
  groupLevel: "mentions",
  sound: true,
  vibration: true,
  previewContent: true,
  showSender: true,
  showPersonaName: true,
};

export const defaultFeatures: FeatureSettings = {
  voiceCalls: true,
  videoCalls: true,
  fileSharing: true,
  locationSharing: false,
  disappearingMessages: false,
  scheduledMessages: false,
  maxGroupSize: 50,
  maxFileSize: 25,
};

// Mock master identity
export const mockMasterIdentity: MasterIdentity = {
  userId: "user_alex",
  did: "did:echo:alex_morgan_2024",
  trustLevel: "trusted",
  trustScore: 87,
  phoneVerified: true,
  emailVerified: true,
  kycVerified: false,
  personaCount: 4,
  maxPersonas: 7,
};

// Mock personas
export const mockPersonas: Persona[] = [
  {
    personaId: "p1",
    userId: "user_alex",
    displayName: "Alex Morgan",
    username: "alexmorgan",
    avatar: "",
    bio: "Design-obsessed product thinker. Building beautiful experiences.",
    category: "professional",
    isActive: true,
    isDefault: true,
    createdAt: "2024-03-15",
    gradient: "primary",
    derivationPath: "m/867530'/0'/0'",
    privacy: { ...defaultPrivacy },
    notifications: { ...defaultNotifications },
    features: { ...defaultFeatures },
    grants: [
      {
        grantId: "gr1",
        personaId: "p1",
        contactId: "u1",
        contactName: "Sarah Chen",
        permissions: { canView: true, canMessage: true, canCall: true, canSeeOtherPersonas: false },
        grantedAt: "2024-04-01",
      },
      {
        grantId: "gr2",
        personaId: "p1",
        contactId: "u2",
        contactName: "Marcus Johnson",
        permissions: { canView: true, canMessage: true, canCall: true, canSeeOtherPersonas: true },
        grantedAt: "2024-04-05",
      },
    ],
  },
  {
    personaId: "p2",
    userId: "user_alex",
    displayName: "Lex",
    username: "lex_creates",
    avatar: "",
    bio: "Digital artist & illustrator. Commissions open.",
    category: "creative",
    isActive: false,
    isDefault: false,
    createdAt: "2024-06-10",
    gradient: "cool",
    derivationPath: "m/867530'/1'/0'",
    privacy: {
      ...defaultPrivacy,
      lastSeenVisible: false,
      searchable: true,
      crossPersonaLinking: false,
    },
    notifications: {
      ...defaultNotifications,
      callLevel: "none",
      groupLevel: "none",
    },
    features: {
      ...defaultFeatures,
      voiceCalls: false,
      videoCalls: false,
      fileSharing: true,
    },
    grants: [
      {
        grantId: "gr3",
        personaId: "p2",
        contactId: "u8",
        contactName: "Mia Davis",
        permissions: { canView: true, canMessage: true, canCall: false, canSeeOtherPersonas: false },
        grantedAt: "2024-07-01",
      },
    ],
  },
  {
    personaId: "p3",
    userId: "user_alex",
    displayName: "ShadowFox",
    username: "shadowfox42",
    avatar: "",
    bio: "Competitive gamer. FPS enthusiast. Always looking for a squad.",
    category: "gaming",
    isActive: false,
    isDefault: false,
    createdAt: "2024-08-20",
    gradient: "warm",
    derivationPath: "m/867530'/2'/0'",
    privacy: {
      ...defaultPrivacy,
      lastSeenVisible: false,
      onlineStatusVisible: false,
      profilePictureVisible: "contacts",
      bioVisible: "contacts",
      searchable: false,
      crossPersonaLinking: false,
      trustScoreSharing: false,
    },
    notifications: {
      ...defaultNotifications,
      messageLevel: "mentions",
      previewContent: false,
    },
    features: {
      ...defaultFeatures,
      voiceCalls: true,
      videoCalls: true,
      locationSharing: false,
      disappearingMessages: true,
    },
    grants: [],
  },
  {
    personaId: "p4",
    userId: "user_alex",
    displayName: "Ghost",
    username: "ghost_anon",
    avatar: "",
    bio: "",
    category: "anonymous",
    isActive: false,
    isDefault: false,
    createdAt: "2025-01-05",
    gradient: "secondary",
    derivationPath: "m/867530'/3'/0'",
    privacy: {
      lastSeenVisible: false,
      onlineStatusVisible: false,
      profilePictureVisible: "nobody",
      bioVisible: "nobody",
      statusVisible: "nobody",
      readReceipts: false,
      typingIndicator: false,
      searchable: false,
      suggestable: false,
      contactSharing: false,
      crossPersonaLinking: false,
      trustScoreSharing: false,
    },
    notifications: {
      ...defaultNotifications,
      enabled: false,
    },
    features: {
      voiceCalls: false,
      videoCalls: false,
      fileSharing: false,
      locationSharing: false,
      disappearingMessages: true,
      scheduledMessages: false,
      maxGroupSize: 10,
      maxFileSize: 5,
    },
    grants: [],
  },
];
