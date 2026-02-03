/**
 * Echo App - Core TypeScript Types
 * Auto-generate from OpenAPI spec using: npm run generate-api
 * These types are provided as reference; generated types should be preferred.
 */

// =============================================================================
// Common Types
// =============================================================================

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface PaginatedResponse<T> {
  items: T[];
  next_page_token?: string;
  total_count?: number;
}

// =============================================================================
// Authentication
// =============================================================================

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  user: UserProfile;
}

export interface DeviceInfo {
  device_id?: string;
  platform: 'ios' | 'android';
  os_version?: string;
  app_version?: string;
  device_name?: string;
}

// =============================================================================
// Identity & Profile
// =============================================================================

export interface DIDDocument {
  id: string;
  controller?: string;
  verificationMethod: VerificationMethod[];
  authentication?: string[];
  service?: DIDService[];
}

export interface VerificationMethod {
  id: string;
  type: string;
  controller: string;
  publicKeyMultibase?: string;
}

export interface DIDService {
  id: string;
  type: string;
  serviceEndpoint: string;
}

export interface UserProfile {
  did: string;
  display_name: string;
  username?: string;
  bio?: string;
  avatar_url?: string;
  status?: string;
  trust_score: number;
  verification_badges: VerificationBadge[];
  created_at: string;
  last_seen?: string;
  online?: boolean;
}

export interface Persona {
  id: string;
  display_name: string;
  category: 'professional' | 'personal' | 'family' | 'gaming' | 'custom';
  custom_label?: string;
  avatar_url?: string;
  bio?: string;
  is_default?: boolean;
  created_at: string;
}

export interface TrustScore {
  score: number;
  level: 'new' | 'basic' | 'trusted' | 'verified' | 'highly_trusted';
  breakdown: {
    verification_score: number;
    interaction_score: number;
    longevity_score: number;
    report_penalty: number;
  };
  multiplier: number;
  last_updated: string;
}

export interface VerificationBadge {
  type: 'identity_verified' | 'phone_verified' | 'email_verified' | 'professional' | 'government' | 'financial_institution';
  display_name: string;
  icon_url?: string;
  verified_at: string;
  issuer?: string;
  expires_at?: string;
}

export interface VerifiableCredential {
  id: string;
  type: string[];
  issuer: string;
  issuance_date: string;
  expiration_date?: string;
  credential_subject: Record<string, unknown>;
  proof?: Record<string, unknown>;
}

// =============================================================================
// Contacts
// =============================================================================

export type TrustCircle = 'inner_circle' | 'trusted' | 'acquaintance';

export interface Contact {
  did: string;
  display_name: string;
  nickname?: string;
  avatar_url?: string;
  trust_score?: number;
  trust_circle: TrustCircle;
  verification_badges: VerificationBadge[];
  mutual_contacts_count?: number;
  blocked: boolean;
  muted: boolean;
  online?: boolean;
  last_seen?: string;
  added_at: string;
}

export interface ContactRequest {
  id: string;
  from_did: string;
  to_did: string;
  from_user?: UserProfile;
  message?: string;
  status: 'pending' | 'accepted' | 'rejected' | 'expired';
  created_at: string;
  expires_at?: string;
}

// =============================================================================
// Conversations & Messaging
// =============================================================================

export type ConversationType = 'direct' | 'group' | 'channel';
export type ParticipantRole = 'owner' | 'admin' | 'moderator' | 'member';
export type MessageContentType = 'text' | 'image' | 'video' | 'audio' | 'file' | 'location' | 'poll' | 'system';

export interface ConversationSummary {
  id: string;
  type: ConversationType;
  name?: string;
  avatar_url?: string;
  last_message?: MessagePreview;
  unread_count: number;
  muted: boolean;
  pinned: boolean;
  archived: boolean;
  hidden: boolean;
  updated_at: string;
}

export interface Conversation {
  id: string;
  type: 'direct' | 'group';
  name?: string;
  description?: string;
  avatar_url?: string;
  participants: Participant[];
  participant_count: number;
  settings: ConversationSettings;
  my_role: ParticipantRole;
  created_at: string;
  updated_at: string;
}

export interface ConversationSettings {
  disappearing_messages?: {
    enabled: boolean;
    duration_seconds: number;
  };
  min_trust_score_to_join?: number;
  allow_screenshots?: boolean;
  provable_mode?: boolean;
}

export interface Participant {
  did: string;
  display_name?: string;
  avatar_url?: string;
  role: ParticipantRole;
  trust_score?: number;
  verification_badges: VerificationBadge[];
  joined_at: string;
  online?: boolean;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_did: string;
  sender?: UserProfile;
  content: string;
  content_type: MessageContentType;
  attachments: Attachment[];
  reply_to?: {
    message_id: string;
    preview: string;
    sender_did: string;
  };
  forwarded_from?: {
    original_sender_did: string;
    original_conversation_id: string;
  };
  reactions: ReactionSummary[];
  mentions: string[];
  edited: boolean;
  edited_at?: string;
  deleted: boolean;
  silent: boolean;
  blockchain_proof?: {
    tx_hash: string;
    anchored_at: string;
  };
  created_at: string;
  expires_at?: string;
  read_by?: Array<{ did: string; read_at: string }>;
}

export interface MessagePreview {
  id: string;
  content: string;
  content_type: MessageContentType;
  sender_did: string;
  sender_name: string;
  created_at: string;
}

export interface SendMessageRequest {
  content: string;
  content_type?: 'text' | 'location';
  attachments?: AttachmentInput[];
  reply_to_id?: string;
  mentions?: string[];
  silent?: boolean;
  provable?: boolean;
  disappears_after?: number;
}

export interface AttachmentInput {
  file_id: string;
  caption?: string;
}

export interface Attachment {
  id: string;
  type: 'image' | 'video' | 'audio' | 'document' | 'other';
  file_id: string;
  filename?: string;
  mime_type?: string;
  size_bytes?: number;
  thumbnail_url?: string;
  download_url?: string;
  caption?: string;
  duration_seconds?: number;
  dimensions?: { width: number; height: number };
}

export interface Reaction {
  emoji: string;
  user_did: string;
  user?: UserProfile;
  created_at: string;
}

export interface ReactionSummary {
  emoji: string;
  count: number;
  includes_me: boolean;
}

export interface MessageProof {
  message_id: string;
  hash: string;
  tx_hash: string;
  block_hash?: string;
  anchored_at: string;
  verification_url?: string;
  zk_proof?: string;
}

export interface ScheduledMessage {
  id: string;
  conversation_id: string;
  content: string;
  content_type?: MessageContentType;
  attachments?: AttachmentInput[];
  scheduled_at: string;
  timezone?: string;
  created_at: string;
}

export interface Poll {
  id: string;
  question: string;
  options: Array<{
    index: number;
    text: string;
    vote_count: number;
    voters?: string[];
  }>;
  total_votes: number;
  allows_multiple_answers: boolean;
  is_anonymous: boolean;
  is_closed: boolean;
  closes_at?: string;
  my_votes: number[];
  created_at: string;
}

// =============================================================================
// Groups & Channels
// =============================================================================

export interface GroupSummary {
  id: string;
  name: string;
  description?: string;
  avatar_url?: string;
  member_count: number;
  visibility: 'public' | 'private';
  verification_score?: number;
  my_role?: ParticipantRole;
  category?: string;
}

export interface Group extends GroupSummary {
  owner_did: string;
  online_count?: number;
  join_approval_required: boolean;
  min_trust_score?: number;
  invite_link?: string;
  tags?: string[];
  created_at: string;
}

export interface CreateGroupRequest {
  name: string;
  description?: string;
  avatar?: string;
  visibility?: 'public' | 'private';
  join_approval_required?: boolean;
  min_trust_score?: number;
  initial_members?: string[];
  category?: string;
  tags?: string[];
}

export interface GroupInvite {
  code: string;
  link: string;
  max_uses?: number;
  uses: number;
  requires_approval: boolean;
  expires_at?: string;
  created_by: string;
  created_at: string;
}

export interface ChannelSummary {
  id: string;
  name: string;
  description?: string;
  avatar_url?: string;
  subscriber_count: number;
  is_subscribed: boolean;
  my_role?: 'owner' | 'admin' | 'subscriber';
}

export interface Channel extends ChannelSummary {
  owner_did: string;
  visibility: 'public' | 'private';
  category?: string;
  created_at: string;
}

export interface ChannelPost {
  id: string;
  content: string;
  attachments: Attachment[];
  view_count: number;
  reaction_count: number;
  comment_count: number;
  created_at: string;
  edited_at?: string;
}

// =============================================================================
// Calls
// =============================================================================

export interface Call {
  id: string;
  conversation_id: string;
  type: 'voice' | 'video';
  status: 'ringing' | 'active' | 'ended' | 'missed';
  initiator_did: string;
  participants: Array<{
    did: string;
    display_name?: string;
    status: 'ringing' | 'connected' | 'disconnected';
    video_enabled: boolean;
    audio_enabled: boolean;
    screen_sharing: boolean;
  }>;
  started_at?: string;
  ended_at?: string;
  duration_seconds?: number;
}

export interface IceServer {
  urls: string[];
  username?: string;
  credential?: string;
}

// =============================================================================
// Files
// =============================================================================

export interface FileUploadResponse {
  file_id: string;
  filename: string;
  mime_type?: string;
  size_bytes: number;
  ipfs_hash?: string;
  blockchain_tx?: string;
  thumbnail_url?: string;
}

export interface FileMetadata {
  id: string;
  filename: string;
  mime_type?: string;
  size_bytes: number;
  ipfs_hash?: string;
  owner_did: string;
  created_at: string;
  expires_at?: string;
}

export interface FileProof {
  file_id: string;
  hash: string;
  ipfs_hash?: string;
  blockchain_tx?: string;
  anchored_at: string;
  verification_url?: string;
}

// =============================================================================
// Tokens & Rewards
// =============================================================================

export interface TokenBalance {
  available: number;
  staked: number;
  pending_rewards: number;
  total: number;
}

export interface RewardsTracker {
  total_earned: number;
  period_earned: number;
  breakdown: {
    messaging: number;
    payment_rail: number;
    referrals: number;
    staking: number;
    achievements: number;
  };
  daily_earnings: Array<{ date: string; amount: number }>;
  multiplier: number;
  daily_cap: number;
  daily_remaining: number;
}

export interface RewardTransaction {
  id: string;
  type: 'messaging' | 'payment_rail' | 'referral' | 'staking' | 'achievement' | 'bonus';
  amount: number;
  description?: string;
  blockchain_tx?: string;
  created_at: string;
}

export interface Achievement {
  id: string;
  name: string;
  description?: string;
  icon_url?: string;
  reward_amount: number;
  progress: number;
  target?: number;
  current?: number;
  completed_at?: string;
  multiplier_bonus?: number;
}

export interface StakingInfo {
  staked_amount: number;
  apy: number;
  pending_rewards: number;
  validator_node?: string;
  staked_at?: string;
  unlock_at?: string;
}

export interface ReferralInfo {
  referral_code: string;
  referral_link: string;
  total_referrals: number;
  verified_referrals: number;
  total_earned: number;
  tiers: Array<{
    tier: number;
    count: number;
    earnings: number;
  }>;
}

// =============================================================================
// Notifications
// =============================================================================

export type NotificationType = 'message' | 'mention' | 'reaction' | 'contact_request' | 'group_invite' | 'call_missed' | 'reward' | 'system';

export interface Notification {
  id: string;
  type: NotificationType;
  title?: string;
  body?: string;
  data?: Record<string, unknown>;
  read: boolean;
  created_at: string;
}

export interface NotificationSettings {
  push_enabled: boolean;
  messages: boolean;
  mentions: boolean;
  reactions: boolean;
  contact_requests: boolean;
  group_invites: boolean;
  calls: boolean;
  rewards: boolean;
  quiet_hours?: {
    enabled: boolean;
    start: string;
    end: string;
    timezone: string;
  };
}

export interface Device {
  id: string;
  platform: 'ios' | 'android';
  device_name?: string;
  last_active?: string;
  registered_at: string;
}

// =============================================================================
// Enterprise
// =============================================================================

export type OrganizationType = 'corporation' | 'financial_institution' | 'government' | 'nonprofit';
export type VerificationTier = 'basic' | 'regulated' | 'government';
export type EmployeeRole = 'admin' | 'manager' | 'representative' | 'support';

export interface Organization {
  id: string;
  did?: string;
  legal_name: string;
  display_name?: string;
  type: OrganizationType;
  logo_url?: string;
  verification_status: 'pending' | 'verified' | 'suspended';
  verification_tier?: VerificationTier;
  employee_count?: number;
  created_at: string;
}

export interface Employee {
  did: string;
  display_name?: string;
  role: EmployeeRole;
  department?: string;
  title?: string;
  verified: boolean;
  added_at: string;
}

// =============================================================================
// Bots
// =============================================================================

export interface Bot {
  id: string;
  display_name: string;
  username?: string;
  description?: string;
  avatar_url?: string;
  owner_did: string;
  trust_score?: number;
  webhook_url?: string;
  commands: BotCommand[];
  active: boolean;
  created_at: string;
}

export interface BotListing {
  id: string;
  display_name: string;
  description?: string;
  avatar_url?: string;
  category?: string;
  trust_score?: number;
  user_count?: number;
  rating?: number;
  review_count?: number;
}

export interface BotCommand {
  command: string;
  description: string;
  parameters?: Array<{
    name: string;
    description: string;
    required: boolean;
  }>;
}

export interface BotPermissions {
  can_read_messages: boolean;
  can_send_messages: boolean;
  can_read_members: boolean;
  can_manage_messages: boolean;
  can_initiate_payments: boolean;
  allowed_conversations?: string[];
}

// =============================================================================
// WebSocket Events
// =============================================================================

export type WebSocketEventType =
  | 'message.new'
  | 'message.updated'
  | 'message.deleted'
  | 'message.reaction'
  | 'typing.start'
  | 'typing.stop'
  | 'presence.update'
  | 'call.incoming'
  | 'call.update'
  | 'notification';

export interface WebSocketEvent<T = unknown> {
  type: WebSocketEventType;
  conversation_id?: string;
  data: T;
  timestamp: string;
}

export interface TypingEvent {
  conversation_id: string;
  user_did: string;
  typing: boolean;
}

export interface PresenceEvent {
  user_did: string;
  online: boolean;
  last_seen?: string;
}

export interface CallEvent {
  call: Call;
  event: 'incoming' | 'answered' | 'ended' | 'participant_joined' | 'participant_left';
}
