import React, { useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
  ArrowLeft, MoreHorizontal, Shield, Eye, EyeOff, Users,
  Phone, Send, Lock, UserPlus, UserMinus, ShieldCheck,
  Key,
} from "lucide-react";
import { EchoAvatar } from "../components/echo-avatar";
import {
  mockPersonas,
  categoryInfo,
  type VisibilityLevel,
} from "../components/persona-data";
import { motion } from "motion/react";

const tabs = ["Profile", "Privacy", "Access", "Features"];

function Toggle({ value, onChange }: { value: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      onClick={() => onChange(!value)}
      className="relative rounded-full border-none cursor-pointer shrink-0 transition-all"
      style={{
        width: 48,
        height: 28,
        backgroundColor: value ? "var(--echo-primary)" : "var(--echo-border)",
      }}
    >
      <div
        className="absolute top-1 rounded-full transition-all"
        style={{
          width: 20,
          height: 20,
          backgroundColor: "white",
          left: value ? 26 : 2,
          boxShadow: "var(--echo-shadow-sm)",
        }}
      />
    </button>
  );
}

function SettingRow({
  label,
  desc,
  value,
  onChange,
}: {
  label: string;
  desc?: string;
  value: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between py-3">
      <div>
        <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>{label}</div>
        {desc && <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 1 }}>{desc}</div>}
      </div>
      <Toggle value={value} onChange={onChange} />
    </div>
  );
}

const visibilityLabels: Record<VisibilityLevel, string> = {
  everyone: "Everyone",
  contacts: "Contacts Only",
  nobody: "Nobody",
};

export function PersonaDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [activeTab, setActiveTab] = useState(0);
  const [showGrantForm, setShowGrantForm] = useState(false);
  const [newGrantContact, setNewGrantContact] = useState("");

  const persona = mockPersonas.find((p) => p.personaId === id);

  if (!persona) {
    return (
      <div className="flex flex-col items-center justify-center h-full px-6">
        <p style={{ fontSize: 16, color: "var(--echo-text-tertiary)" }}>Persona not found</p>
        <button
          onClick={() => navigate(-1)}
          className="mt-4 bg-transparent border-none cursor-pointer"
          style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-primary)" }}
        >
          Go back
        </button>
      </div>
    );
  }

  const cat = categoryInfo[persona.category];

  // Local state for editable settings
  const [privacy, setPrivacy] = useState(persona.privacy);
  const [notifications, setNotifications] = useState(persona.notifications);
  const [features, setFeatures] = useState(persona.features);

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="shrink-0"
        style={{
          background: `linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))`,
          paddingTop: 44,
        }}
      >
        <div className="flex items-center justify-between px-4 py-2">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center justify-center bg-transparent border-none cursor-pointer rounded-full"
            style={{ width: 44, height: 44 }}
          >
            <ArrowLeft size={24} color="white" />
          </button>
          <button
            className="flex items-center justify-center bg-transparent border-none cursor-pointer rounded-full"
            style={{ width: 44, height: 44 }}
          >
            <MoreHorizontal size={24} color="white" />
          </button>
        </div>

        <div className="flex flex-col items-center pb-5 px-6">
          <div className="relative mb-3">
            <EchoAvatar name={persona.displayName} size="xl" gradient={persona.gradient} />
            <div
              className="absolute -bottom-1 -right-1 flex items-center justify-center rounded-full"
              style={{
                width: 28,
                height: 28,
                backgroundColor: cat.color,
                border: "3px solid white",
                fontSize: 14,
              }}
            >
              {cat.emoji}
            </div>
          </div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: "white", margin: "0 0 2px" }}>
            {persona.displayName}
          </h1>
          <div style={{ fontSize: 14, color: "rgba(255,255,255,0.7)" }}>
            @{persona.username} · {cat.label}
          </div>
          {persona.isActive && (
            <span
              className="mt-2 rounded-full px-3 py-1"
              style={{ fontSize: 11, fontWeight: 700, color: "white", backgroundColor: "rgba(255,255,255,0.25)" }}
            >
              ACTIVE PERSONA
            </span>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div
        className="flex shrink-0"
        style={{ borderBottom: "1px solid var(--echo-border-light)", backgroundColor: "var(--echo-bg-elevated)" }}
      >
        {tabs.map((tab, i) => (
          <button
            key={tab}
            onClick={() => setActiveTab(i)}
            className="flex-1 flex items-center justify-center bg-transparent border-none cursor-pointer py-3 transition-all"
            style={{
              fontSize: 13,
              fontWeight: activeTab === i ? 600 : 400,
              color: activeTab === i ? "var(--echo-primary)" : "var(--echo-text-tertiary)",
              borderBottom: activeTab === i ? "2px solid var(--echo-primary)" : "2px solid transparent",
            }}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto">
        {/* PROFILE */}
        {activeTab === 0 && (
          <div className="px-6 py-4">
            {persona.bio && (
              <div className="mb-6">
                <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "0 0 8px" }}>
                  Bio
                </h3>
                <p style={{ fontSize: 15, color: "var(--echo-text-secondary)", lineHeight: "22px", margin: 0 }}>
                  {persona.bio}
                </p>
              </div>
            )}

            {/* Key Derivation */}
            <div className="mb-6">
              <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "0 0 8px" }}>
                Cryptographic Identity
              </h3>
              <div className="rounded-2xl p-4" style={{ backgroundColor: "var(--echo-bg-surface)" }}>
                <div className="flex items-center gap-3 mb-3">
                  <Key size={18} style={{ color: "var(--echo-primary)" }} />
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                      Derivation Path
                    </div>
                    <div style={{ fontSize: 13, fontFamily: "monospace", color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                      {persona.derivationPath}
                    </div>
                  </div>
                </div>
                <div className="flex gap-2">
                  <div className="flex-1 rounded-lg p-2" style={{ backgroundColor: "var(--echo-bg-elevated)" }}>
                    <div style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>Signing Key</div>
                    <div style={{ fontSize: 11, fontFamily: "monospace", color: "var(--echo-success)", marginTop: 2 }}>Active</div>
                  </div>
                  <div className="flex-1 rounded-lg p-2" style={{ backgroundColor: "var(--echo-bg-elevated)" }}>
                    <div style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>Encryption Key</div>
                    <div style={{ fontSize: 11, fontFamily: "monospace", color: "var(--echo-success)", marginTop: 2 }}>Active</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Metadata */}
            <div
              className="rounded-2xl overflow-hidden"
              style={{ backgroundColor: "var(--echo-bg-elevated)", boxShadow: "var(--echo-shadow-sm)" }}
            >
              {[
                { label: "Created", value: new Date(persona.createdAt).toLocaleDateString("en-US", { month: "long", day: "numeric", year: "numeric" }) },
                { label: "Category", value: `${cat.emoji} ${cat.label}` },
                { label: "Default", value: persona.isDefault ? "Yes" : "No" },
                { label: "Access Grants", value: `${persona.grants.length} contacts` },
              ].map((item, i) => (
                <div
                  key={item.label}
                  className="flex items-center justify-between px-4 py-3"
                  style={{ borderBottom: i < 3 ? "1px solid var(--echo-border-light)" : "none" }}
                >
                  <span style={{ fontSize: 14, color: "var(--echo-text-secondary)" }}>{item.label}</span>
                  <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>{item.value}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* PRIVACY */}
        {activeTab === 1 && (
          <div className="px-6 py-4">
            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "0 0 12px" }}>
              Visibility Controls
            </h3>
            {(["profilePictureVisible", "bioVisible", "statusVisible"] as const).map((field) => (
              <div key={field} className="flex items-center justify-between py-3" style={{ borderBottom: "1px solid var(--echo-border-light)" }}>
                <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                  {field === "profilePictureVisible" ? "Profile Picture" : field === "bioVisible" ? "Bio" : "Status"}
                </span>
                <div className="flex gap-1">
                  {(["everyone", "contacts", "nobody"] as VisibilityLevel[]).map((level) => (
                    <button
                      key={level}
                      onClick={() => setPrivacy({ ...privacy, [field]: level })}
                      className="rounded-full px-2.5 py-1 border-none cursor-pointer transition-all"
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        color: privacy[field] === level ? "white" : "var(--echo-text-tertiary)",
                        backgroundColor: privacy[field] === level ? "var(--echo-primary)" : "var(--echo-bg-surface)",
                      }}
                    >
                      {visibilityLabels[level]}
                    </button>
                  ))}
                </div>
              </div>
            ))}

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 8px" }}>
              Communication
            </h3>
            <SettingRow label="Last Seen" desc="Show when you were last active" value={privacy.lastSeenVisible} onChange={(v) => setPrivacy({ ...privacy, lastSeenVisible: v })} />
            <SettingRow label="Online Status" desc="Show green dot when active" value={privacy.onlineStatusVisible} onChange={(v) => setPrivacy({ ...privacy, onlineStatusVisible: v })} />
            <SettingRow label="Read Receipts" value={privacy.readReceipts} onChange={(v) => setPrivacy({ ...privacy, readReceipts: v })} />
            <SettingRow label="Typing Indicator" value={privacy.typingIndicator} onChange={(v) => setPrivacy({ ...privacy, typingIndicator: v })} />

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 8px" }}>
              Discovery
            </h3>
            <SettingRow label="Searchable" desc="Appear in search results" value={privacy.searchable} onChange={(v) => setPrivacy({ ...privacy, searchable: v })} />
            <SettingRow label="Suggestable" desc="Suggest to people you may know" value={privacy.suggestable} onChange={(v) => setPrivacy({ ...privacy, suggestable: v })} />
            <SettingRow label="Contact Sharing" desc="Allow contacts to share your persona" value={privacy.contactSharing} onChange={(v) => setPrivacy({ ...privacy, contactSharing: v })} />

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 8px" }}>
              Cross-Persona
            </h3>
            <SettingRow label="Cross-Persona Linking" desc="Allow others to discover your other personas" value={privacy.crossPersonaLinking} onChange={(v) => setPrivacy({ ...privacy, crossPersonaLinking: v })} />
            <SettingRow label="Trust Score Sharing" desc="Share your trust score across personas" value={privacy.trustScoreSharing} onChange={(v) => setPrivacy({ ...privacy, trustScoreSharing: v })} />
          </div>
        )}

        {/* ACCESS GRANTS */}
        {activeTab === 2 && (
          <div className="px-6 py-4">
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: 0 }}>
                Access Grants ({persona.grants.length})
              </h3>
              <button
                onClick={() => setShowGrantForm(!showGrantForm)}
                className="flex items-center gap-1.5 rounded-full px-3 py-1.5 border-none cursor-pointer"
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: "var(--echo-primary)",
                  backgroundColor: "var(--echo-primary-subtle)",
                }}
              >
                <UserPlus size={14} />
                Grant Access
              </button>
            </div>

            {showGrantForm && (
              <motion.div
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                className="rounded-2xl p-4 mb-4"
                style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-border)" }}
              >
                <label style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                  Contact Name
                </label>
                <input
                  value={newGrantContact}
                  onChange={(e) => setNewGrantContact(e.target.value)}
                  placeholder="Search contacts..."
                  className="w-full rounded-xl px-4 outline-none mb-3"
                  style={{
                    height: 44,
                    backgroundColor: "var(--echo-bg-elevated)",
                    border: "2px solid var(--echo-border)",
                    fontSize: 14,
                    color: "var(--echo-text-primary)",
                  }}
                />
                <div className="flex flex-col gap-2 mb-3">
                  <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-text-tertiary)" }}>Permissions</span>
                  {["Can View", "Can Message", "Can Call", "Can See Other Personas"].map((perm) => (
                    <div key={perm} className="flex items-center justify-between">
                      <span style={{ fontSize: 13, color: "var(--echo-text-secondary)" }}>{perm}</span>
                      <Toggle value={perm !== "Can See Other Personas"} onChange={() => {}} />
                    </div>
                  ))}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => setShowGrantForm(false)}
                    className="flex-1 rounded-lg border-none cursor-pointer py-2"
                    style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)", backgroundColor: "var(--echo-bg-elevated)" }}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => setShowGrantForm(false)}
                    className="flex-1 rounded-lg border-none cursor-pointer py-2"
                    style={{ fontSize: 13, fontWeight: 600, color: "white", backgroundColor: "var(--echo-primary)" }}
                  >
                    Grant
                  </button>
                </div>
              </motion.div>
            )}

            {persona.grants.length === 0 ? (
              <div className="flex flex-col items-center py-12">
                <Lock size={32} style={{ color: "var(--echo-text-tertiary)", marginBottom: 8 }} />
                <p style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-secondary)", margin: "0 0 4px" }}>
                  No access grants
                </p>
                <p style={{ fontSize: 13, color: "var(--echo-text-tertiary)", textAlign: "center" }}>
                  This persona is completely private. Grant access to specific contacts.
                </p>
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {persona.grants.map((grant) => (
                  <div
                    key={grant.grantId}
                    className="rounded-2xl p-4"
                    style={{ backgroundColor: "var(--echo-bg-elevated)", border: "1px solid var(--echo-border)" }}
                  >
                    <div className="flex items-center gap-3 mb-3">
                      <EchoAvatar name={grant.contactName} size="sm" gradient="primary" />
                      <div className="flex-1">
                        <div style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                          {grant.contactName}
                        </div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          Granted {new Date(grant.grantedAt).toLocaleDateString()}
                          {grant.expiresAt && ` · Expires ${new Date(grant.expiresAt).toLocaleDateString()}`}
                        </div>
                      </div>
                      <button
                        className="flex items-center gap-1 rounded-full px-2.5 py-1 border-none cursor-pointer"
                        style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-error)", backgroundColor: "var(--echo-error-light)" }}
                      >
                        <UserMinus size={12} />
                        Revoke
                      </button>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {[
                        { label: "View", enabled: grant.permissions.canView, icon: Eye },
                        { label: "Message", enabled: grant.permissions.canMessage, icon: Send },
                        { label: "Call", enabled: grant.permissions.canCall, icon: Phone },
                        { label: "See Others", enabled: grant.permissions.canSeeOtherPersonas, icon: Users },
                      ].map((perm) => {
                        const Icon = perm.icon;
                        return (
                          <span
                            key={perm.label}
                            className="flex items-center gap-1 rounded-full px-2 py-0.5"
                            style={{
                              fontSize: 11,
                              fontWeight: 500,
                              color: perm.enabled ? "var(--echo-success)" : "var(--echo-text-tertiary)",
                              backgroundColor: perm.enabled ? "var(--echo-success-light)" : "var(--echo-bg-surface)",
                            }}
                          >
                            <Icon size={10} />
                            {perm.label}
                          </span>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* FEATURES */}
        {activeTab === 3 && (
          <div className="px-6 py-4">
            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "0 0 12px" }}>
              Communication
            </h3>
            <SettingRow label="Voice Calls" desc="Allow voice calls with this persona" value={features.voiceCalls} onChange={(v) => setFeatures({ ...features, voiceCalls: v })} />
            <SettingRow label="Video Calls" desc="Allow video calls" value={features.videoCalls} onChange={(v) => setFeatures({ ...features, videoCalls: v })} />

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 12px" }}>
              Content
            </h3>
            <SettingRow label="File Sharing" desc="Send and receive files" value={features.fileSharing} onChange={(v) => setFeatures({ ...features, fileSharing: v })} />
            <SettingRow label="Location Sharing" desc="Share your location" value={features.locationSharing} onChange={(v) => setFeatures({ ...features, locationSharing: v })} />

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 12px" }}>
              Security
            </h3>
            <SettingRow label="Disappearing Messages" desc="Auto-delete messages after reading" value={features.disappearingMessages} onChange={(v) => setFeatures({ ...features, disappearingMessages: v })} />
            <SettingRow label="Scheduled Messages" desc="Send messages at a scheduled time" value={features.scheduledMessages} onChange={(v) => setFeatures({ ...features, scheduledMessages: v })} />

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 12px" }}>
              Limits
            </h3>
            <div className="flex items-center justify-between py-3" style={{ borderBottom: "1px solid var(--echo-border-light)" }}>
              <div>
                <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>Max Group Size</div>
                <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Maximum groups this persona can create</div>
              </div>
              <span style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-primary)" }}>{features.maxGroupSize}</span>
            </div>
            <div className="flex items-center justify-between py-3">
              <div>
                <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>Max File Size</div>
                <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Maximum file upload size</div>
              </div>
              <span style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-primary)" }}>{features.maxFileSize} MB</span>
            </div>

            <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", margin: "20px 0 12px" }}>
              Notifications
            </h3>
            <SettingRow label="Notifications Enabled" value={notifications.enabled} onChange={(v) => setNotifications({ ...notifications, enabled: v })} />
            {notifications.enabled && (
              <>
                <SettingRow label="Sound" value={notifications.sound} onChange={(v) => setNotifications({ ...notifications, sound: v })} />
                <SettingRow label="Vibration" value={notifications.vibration} onChange={(v) => setNotifications({ ...notifications, vibration: v })} />
                <SettingRow label="Preview Content" desc="Show message content in notifications" value={notifications.previewContent} onChange={(v) => setNotifications({ ...notifications, previewContent: v })} />
                <SettingRow label="Show Sender" value={notifications.showSender} onChange={(v) => setNotifications({ ...notifications, showSender: v })} />
                <SettingRow label="Show Persona Name" desc="Include persona name in notification" value={notifications.showPersonaName} onChange={(v) => setNotifications({ ...notifications, showPersonaName: v })} />
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
}