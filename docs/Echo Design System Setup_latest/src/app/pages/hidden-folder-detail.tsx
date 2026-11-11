import React, { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { motion, AnimatePresence } from "motion/react";
import {
  ArrowLeft, Lock, FolderLock, Shield, ShieldAlert, ShieldOff,
  Settings, Eye, EyeOff, Timer, VolumeX, Bell, BellOff,
  ChevronRight, ArrowUpRight, TriangleAlert, X, Check,
} from "lucide-react";
import { EchoAvatar } from "../components/echo-avatar";
import {
  useHiddenFolders,
  tierDefaults,
  type SecurityTier,
  type NotificationMode,
} from "../components/hidden-folder-data";

export function HiddenFolderDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { folders, lock, updateFolderSettings, moveConversationOut } = useHiddenFolders();
  const folder = folders.find((f) => f.id === id);
  const [showSettings, setShowSettings] = useState(false);
  const [showMoveConfirm, setShowMoveConfirm] = useState<string | null>(null);

  if (!folder) {
    return (
      <div className="flex flex-col h-full items-center justify-center">
        <FolderLock size={48} style={{ color: "var(--echo-text-tertiary)", marginBottom: 12 }} />
        <div style={{ fontSize: 16, color: "var(--echo-text-tertiary)" }}>Folder not found</div>
        <button
          onClick={() => navigate("/hidden-folders")}
          className="mt-4 px-4 py-2 rounded-xl border-none cursor-pointer"
          style={{
            backgroundColor: "var(--echo-primary)",
            fontSize: 14,
            fontWeight: 600,
            color: "white",
          }}
        >
          Back to folders
        </button>
      </div>
    );
  }

  const tierInfo = tierDefaults[folder.securityTier];

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="shrink-0"
        style={{
          backgroundColor: "var(--echo-bg-elevated)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        <div className="flex items-center gap-3 px-4" style={{ height: 96, paddingTop: 44 }}>
          <button
            onClick={() => navigate("/hidden-folders")}
            className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          >
            <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
          </button>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span
                style={{
                  fontSize: 18,
                  fontWeight: 700,
                  color: "var(--echo-text-primary)",
                  lineHeight: "24px",
                }}
              >
                {folder.name}
              </span>
              <span
                className="rounded-full px-2 py-0.5"
                style={{
                  fontSize: 10,
                  fontWeight: 600,
                  color: tierInfo.color,
                  backgroundColor:
                    folder.securityTier === "maximum"
                      ? "var(--echo-error-light)"
                      : folder.securityTier === "elevated"
                      ? "var(--echo-warning-light)"
                      : "var(--echo-success-light)",
                }}
              >
                {tierInfo.label}
              </span>
            </div>
            <div
              style={{
                fontSize: 12,
                color: "var(--echo-text-tertiary)",
                lineHeight: "16px",
              }}
            >
              {folder.conversationCount} conversation{folder.conversationCount !== 1 ? "s" : ""} &middot; Layer 2 encrypted
            </div>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setShowSettings(true)}
              className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2 rounded-full hover:bg-[var(--echo-bg-surface)]"
            >
              <Settings size={20} style={{ color: "var(--echo-text-secondary)" }} />
            </button>
            <button
              onClick={() => {
                lock();
                navigate("/");
              }}
              className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2 rounded-full hover:bg-[var(--echo-bg-surface)]"
            >
              <Lock size={20} style={{ color: "var(--echo-text-secondary)" }} />
            </button>
          </div>
        </div>
      </div>

      {/* Conversations */}
      <div className="flex-1 overflow-y-auto">
        {/* Security notice */}
        <div
          className="mx-4 mt-3 mb-2 rounded-xl px-3 py-2 flex items-center gap-2"
          style={{
            backgroundColor: "var(--echo-bg-surface)",
            border: "1px solid var(--echo-border-light)",
          }}
        >
          <EyeOff size={14} style={{ color: "var(--echo-text-tertiary)" }} />
          <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", flex: 1 }}>
            Forward & export blocked &middot; Excluded from search &middot; Notifications {folder.notificationMode}
          </span>
        </div>

        {folder.conversations.map((conv) => (
          <div key={conv.id}>
            {showMoveConfirm === conv.id ? (
              <div
                className="mx-4 my-2 rounded-xl p-3"
                style={{
                  backgroundColor: "var(--echo-warning-light)",
                  border: "1px solid var(--echo-warning)",
                }}
              >
                <div
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-primary)",
                    marginBottom: 4,
                  }}
                >
                  Move to main chats?
                </div>
                <div
                  style={{
                    fontSize: 12,
                    color: "var(--echo-text-secondary)",
                    marginBottom: 10,
                  }}
                >
                  This conversation will be re-encrypted and visible in your main chat list.
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => setShowMoveConfirm(null)}
                    className="flex-1 rounded-lg py-2 border-none cursor-pointer"
                    style={{
                      backgroundColor: "var(--echo-bg-elevated)",
                      fontSize: 12,
                      fontWeight: 600,
                      color: "var(--echo-text-primary)",
                    }}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                      moveConversationOut(conv.id, folder.id);
                      setShowMoveConfirm(null);
                    }}
                    className="flex-1 rounded-lg py-2 border-none cursor-pointer"
                    style={{
                      backgroundColor: "var(--echo-warning)",
                      fontSize: 12,
                      fontWeight: 600,
                      color: "white",
                    }}
                  >
                    Move Out
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => navigate(`/hidden-folders/${folder.id}/chat/${conv.id}`)}
                className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer transition-colors hover:bg-[var(--echo-bg-surface)] text-left"
              >
                <EchoAvatar name={conv.name} size="md" online={conv.online} gradient={conv.gradient} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-0.5">
                    <span
                      className="truncate"
                      style={{
                        fontSize: 16,
                        fontWeight: conv.unread > 0 ? 600 : 400,
                        color: "var(--echo-text-primary)",
                        lineHeight: "22px",
                      }}
                    >
                      {conv.name}
                    </span>
                    <span
                      className="shrink-0 ml-2"
                      style={{
                        fontSize: 13,
                        color: conv.unread > 0 ? "var(--echo-primary)" : "var(--echo-text-tertiary)",
                      }}
                    >
                      {conv.time}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span
                      className="truncate"
                      style={{
                        fontSize: 14,
                        fontWeight: conv.unread > 0 ? 500 : 400,
                        color: conv.unread > 0 ? "var(--echo-text-secondary)" : "var(--echo-text-tertiary)",
                        lineHeight: "20px",
                      }}
                    >
                      {conv.lastMessage}
                    </span>
                    <div className="flex items-center gap-1.5 shrink-0 ml-1">
                      {conv.unread > 0 && (
                        <span
                          className="flex items-center justify-center rounded-full"
                          style={{
                            minWidth: 20,
                            height: 20,
                            padding: "0 6px",
                            backgroundColor: "var(--echo-primary)",
                            fontSize: 11,
                            fontWeight: 700,
                            color: "white",
                          }}
                        >
                          {conv.unread}
                        </span>
                      )}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setShowMoveConfirm(conv.id);
                        }}
                        className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1 rounded-lg hover:bg-[var(--echo-bg-surface)]"
                        title="Move to main chats"
                      >
                        <ArrowUpRight size={14} style={{ color: "var(--echo-text-tertiary)" }} />
                      </button>
                    </div>
                  </div>
                </div>
              </button>
            )}
          </div>
        ))}

        {folder.conversations.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16">
            <EyeOff size={40} style={{ color: "var(--echo-text-tertiary)", marginBottom: 12 }} />
            <div style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-tertiary)" }}>
              No hidden conversations
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--echo-text-tertiary)",
                textAlign: "center",
                maxWidth: 220,
                marginTop: 4,
              }}
            >
              Long-press a conversation in your main chat list to move it here.
            </div>
          </div>
        )}
      </div>

      {/* Settings Sheet */}
      <AnimatePresence>
        {showSettings && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50"
              style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
              onClick={() => setShowSettings(false)}
            />
            <motion.div
              initial={{ y: "100%", opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: "100%", opacity: 0 }}
              transition={{ type: "spring", damping: 28, stiffness: 350 }}
              className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
              style={{ backgroundColor: "var(--echo-bg-elevated)", maxHeight: "80vh", overflow: "auto" }}
            >
              <div className="flex justify-center pt-3 pb-1">
                <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
              </div>

              <div className="px-5 pb-8">
                <div className="flex items-center justify-between mb-5">
                  <div style={{ fontSize: 18, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                    Folder Settings
                  </div>
                  <button
                    onClick={() => setShowSettings(false)}
                    className="bg-transparent border-none cursor-pointer p-1"
                  >
                    <X size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                </div>

                {/* Security Tier Display */}
                <div
                  className="rounded-xl p-3 mb-4 flex items-center gap-3"
                  style={{
                    backgroundColor:
                      folder.securityTier === "maximum"
                        ? "var(--echo-error-light)"
                        : folder.securityTier === "elevated"
                        ? "var(--echo-warning-light)"
                        : "var(--echo-success-light)",
                    border: `1px solid ${tierInfo.color}`,
                  }}
                >
                  <span style={{ color: tierInfo.color }}>
                    {folder.securityTier === "maximum" ? (
                      <ShieldOff size={22} />
                    ) : folder.securityTier === "elevated" ? (
                      <ShieldAlert size={22} />
                    ) : (
                      <Shield size={22} />
                    )}
                  </span>
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                      {tierInfo.label} Security
                    </div>
                    <div style={{ fontSize: 12, color: "var(--echo-text-secondary)" }}>
                      {tierInfo.description}
                    </div>
                  </div>
                </div>

                {/* Settings rows */}
                <div className="flex flex-col gap-0.5">
                  {/* Auto-lock */}
                  <div className="flex items-center justify-between py-3">
                    <div className="flex items-center gap-3">
                      <Timer size={18} style={{ color: "var(--echo-text-secondary)" }} />
                      <div>
                        <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                          Auto-lock
                        </div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {folder.autoLockTimeout === 0
                            ? "Immediate on background"
                            : `${folder.autoLockTimeout}s inactive`}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Notifications */}
                  <div className="flex items-center justify-between py-3">
                    <div className="flex items-center gap-3">
                      {folder.notificationMode === "suppressed" ? (
                        <BellOff size={18} style={{ color: "var(--echo-text-secondary)" }} />
                      ) : (
                        <Bell size={18} style={{ color: "var(--echo-text-secondary)" }} />
                      )}
                      <div>
                        <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                          Notifications
                        </div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {folder.notificationMode === "suppressed"
                            ? "Fully suppressed"
                            : folder.notificationMode === "redacted"
                            ? "Redacted (no details)"
                            : "Only when unlocked"}
                        </div>
                      </div>
                    </div>
                    <select
                      value={folder.notificationMode}
                      onChange={(e) =>
                        updateFolderSettings(folder.id, {
                          notificationMode: e.target.value as NotificationMode,
                        })
                      }
                      className="rounded-lg px-2 py-1 border-none outline-none cursor-pointer"
                      style={{
                        fontSize: 12,
                        backgroundColor: "var(--echo-bg-surface)",
                        color: "var(--echo-text-secondary)",
                      }}
                    >
                      <option value="suppressed">Suppressed</option>
                      <option value="redacted">Redacted</option>
                      <option value="unlockedOnly">When Unlocked</option>
                    </select>
                  </div>

                  {/* Screenshot Protection */}
                  <div className="flex items-center justify-between py-3">
                    <div className="flex items-center gap-3">
                      <Eye size={18} style={{ color: "var(--echo-text-secondary)" }} />
                      <div>
                        <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                          Screenshot Protection
                        </div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          Block screenshots & screen recording
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={() =>
                        updateFolderSettings(folder.id, {
                          screenshotProtection: !folder.screenshotProtection,
                        })
                      }
                      className="rounded-full border-none cursor-pointer"
                      style={{
                        width: 44,
                        height: 24,
                        backgroundColor: folder.screenshotProtection
                          ? "var(--echo-primary)"
                          : "var(--echo-border)",
                        position: "relative",
                        transition: "background-color 0.2s",
                      }}
                    >
                      <div
                        className="rounded-full"
                        style={{
                          width: 20,
                          height: 20,
                          backgroundColor: "white",
                          position: "absolute",
                          top: 2,
                          left: folder.screenshotProtection ? 22 : 2,
                          transition: "left 0.2s",
                        }}
                      />
                    </button>
                  </div>

                  {/* Wipe Threshold */}
                  {folder.wipeLockoutThreshold > 0 && (
                    <div className="flex items-center justify-between py-3">
                      <div className="flex items-center gap-3">
                        <TriangleAlert size={18} style={{ color: "var(--echo-error)" }} />
                        <div>
                          <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                            Wipe on Lockout
                          </div>
                          <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                            After {folder.wipeLockoutThreshold} failed attempts
                          </div>
                        </div>
                      </div>
                      <span
                        className="rounded-full px-2 py-0.5"
                        style={{
                          fontSize: 11,
                          fontWeight: 600,
                          color: "var(--echo-error)",
                          backgroundColor: "var(--echo-error-light)",
                        }}
                      >
                        Enabled
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}