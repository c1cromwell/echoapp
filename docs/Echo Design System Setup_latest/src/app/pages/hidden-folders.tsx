import React, { useState } from "react";
import { useNavigate } from "react-router";
import { motion, AnimatePresence } from "motion/react";
import {
  ArrowLeft, Lock, Plus, FolderLock, Shield, ShieldAlert, ShieldOff,
  ChevronRight, Trash2, Pencil, Settings, Eye, EyeOff, Clock,
  Bell, BellOff, X, Check, TriangleAlert,
} from "lucide-react";
import {
  useHiddenFolders,
  tierDefaults,
  type SecurityTier,
  type HiddenFolder,
} from "../components/hidden-folder-data";

const tierIcons: Record<SecurityTier, React.ReactNode> = {
  standard: <Shield size={18} />,
  elevated: <ShieldAlert size={18} />,
  maximum: <ShieldOff size={18} />,
};

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`;
  return `${(bytes / 1073741824).toFixed(1)} GB`;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

export function HiddenFoldersPage() {
  const navigate = useNavigate();
  const { folders, lock, createFolder, deleteFolder, renameFolder, isDuressMode } = useHiddenFolders();
  const [showCreate, setShowCreate] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");

  // Create form state
  const [newName, setNewName] = useState("");
  const [newTier, setNewTier] = useState<SecurityTier>("elevated");

  const handleCreate = () => {
    if (!newName.trim()) return;
    if (folders.length >= 10) return;
    createFolder(newName.trim(), newTier);
    setNewName("");
    setNewTier("elevated");
    setShowCreate(false);
  };

  const handleDelete = (id: string) => {
    deleteFolder(id);
    setShowDeleteConfirm(null);
  };

  const handleRename = (id: string) => {
    if (editName.trim()) {
      renameFolder(id, editName.trim());
    }
    setEditingId(null);
    setEditName("");
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="flex items-center gap-3 px-4 shrink-0"
        style={{
          height: 96,
          paddingTop: 44,
          backgroundColor: "var(--echo-bg-elevated)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        <button
          onClick={() => {
            lock();
            navigate("/");
          }}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <div className="flex-1 flex items-center gap-2">
          <FolderLock size={20} style={{ color: "var(--echo-primary)" }} />
          <span
            style={{
              fontSize: 18,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              lineHeight: "24px",
            }}
          >
            Hidden Folders
          </span>
          {isDuressMode && (
            <span
              className="rounded-full px-2 py-0.5"
              style={{
                fontSize: 10,
                fontWeight: 600,
                color: "var(--echo-warning)",
                backgroundColor: "var(--echo-warning-light)",
              }}
            >
              DURESS
            </span>
          )}
        </div>
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

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {/* Security info banner */}
        <div
          className="mx-4 mt-4 mb-3 rounded-2xl p-3 flex items-start gap-3"
          style={{
            background: "linear-gradient(135deg, var(--echo-primary-subtle), var(--echo-accent-subtle))",
            border: "1px solid var(--echo-primary-lighter)",
          }}
        >
          <div
            className="flex items-center justify-center rounded-xl shrink-0 mt-0.5"
            style={{
              width: 36,
              height: 36,
              background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            }}
          >
            <EyeOff size={18} color="white" />
          </div>
          <div className="flex-1">
            <div
              style={{
                fontSize: 13,
                fontWeight: 600,
                color: "var(--echo-text-primary)",
                lineHeight: "18px",
                marginBottom: 2,
              }}
            >
              AES-256-GCM Encrypted
            </div>
            <div
              style={{
                fontSize: 12,
                color: "var(--echo-text-secondary)",
                lineHeight: "16px",
              }}
            >
              Layer 2 biometric-bound encryption. Content is invisible in main chat, search, and notifications.
            </div>
          </div>
        </div>

        {/* Folders list */}
        <div className="px-4">
          {folders.map((folder, index) => (
            <motion.div
              key={folder.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
            >
              {showDeleteConfirm === folder.id ? (
                /* Delete confirmation inline */
                <div
                  className="rounded-2xl p-4 mb-3"
                  style={{
                    backgroundColor: "var(--echo-error-light)",
                    border: "1px solid var(--echo-error)",
                  }}
                >
                  <div className="flex items-center gap-2 mb-2">
                    <TriangleAlert size={16} style={{ color: "var(--echo-error)" }} />
                    <span style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-error)" }}>
                      Delete "{folder.name}"?
                    </span>
                  </div>
                  <p
                    style={{
                      fontSize: 12,
                      color: "var(--echo-text-secondary)",
                      lineHeight: "16px",
                      margin: "0 0 12px 0",
                    }}
                  >
                    All {folder.conversationCount} conversation(s) will be permanently deleted.
                    This cannot be undone.
                  </p>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setShowDeleteConfirm(null)}
                      className="flex-1 rounded-xl py-2 border-none cursor-pointer"
                      style={{
                        backgroundColor: "var(--echo-bg-elevated)",
                        fontSize: 13,
                        fontWeight: 600,
                        color: "var(--echo-text-primary)",
                      }}
                    >
                      Cancel
                    </button>
                    <button
                      onClick={() => handleDelete(folder.id)}
                      className="flex-1 rounded-xl py-2 border-none cursor-pointer"
                      style={{
                        backgroundColor: "var(--echo-error)",
                        fontSize: 13,
                        fontWeight: 600,
                        color: "white",
                      }}
                    >
                      Delete Forever
                    </button>
                  </div>
                </div>
              ) : (
                /* Folder card */
                <div
                  className="rounded-2xl mb-3 overflow-hidden"
                  style={{
                    backgroundColor: "var(--echo-bg-elevated)",
                    boxShadow: "var(--echo-shadow-sm)",
                    border: "1px solid var(--echo-border-light)",
                  }}
                >
                  {/* Main row — tappable */}
                  <button
                    onClick={() => navigate(`/hidden-folders/${folder.id}`)}
                    className="flex items-center gap-3 w-full px-4 py-3.5 bg-transparent border-none cursor-pointer text-left"
                  >
                    {/* Tier icon */}
                    <div
                      className="flex items-center justify-center rounded-xl shrink-0"
                      style={{
                        width: 44,
                        height: 44,
                        backgroundColor:
                          folder.securityTier === "maximum"
                            ? "var(--echo-error-light)"
                            : folder.securityTier === "elevated"
                            ? "var(--echo-warning-light)"
                            : "var(--echo-success-light)",
                      }}
                    >
                      <span style={{ color: tierDefaults[folder.securityTier].color }}>
                        {tierIcons[folder.securityTier]}
                      </span>
                    </div>

                    <div className="flex-1 min-w-0">
                      {editingId === folder.id ? (
                        <div className="flex items-center gap-2">
                          <input
                            value={editName}
                            onChange={(e) => setEditName(e.target.value)}
                            onKeyDown={(e) => e.key === "Enter" && handleRename(folder.id)}
                            onBlur={() => handleRename(folder.id)}
                            autoFocus
                            className="flex-1 bg-transparent border-none outline-none px-0"
                            style={{
                              fontSize: 16,
                              fontWeight: 600,
                              color: "var(--echo-text-primary)",
                            }}
                            onClick={(e) => e.stopPropagation()}
                          />
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleRename(folder.id);
                            }}
                            className="bg-transparent border-none cursor-pointer p-0"
                          >
                            <Check size={16} style={{ color: "var(--echo-success)" }} />
                          </button>
                        </div>
                      ) : (
                        <>
                          <div className="flex items-center gap-1.5">
                            <span
                              style={{
                                fontSize: 16,
                                fontWeight: 600,
                                color: "var(--echo-text-primary)",
                                lineHeight: "22px",
                              }}
                            >
                              {folder.name}
                            </span>
                            {folder.isDecoy && (
                              <span
                                className="rounded-full px-1.5 py-0.5"
                                style={{
                                  fontSize: 9,
                                  fontWeight: 700,
                                  color: "var(--echo-text-tertiary)",
                                  backgroundColor: "var(--echo-bg-surface)",
                                }}
                              >
                                DECOY
                              </span>
                            )}
                          </div>
                          <div className="flex items-center gap-2 mt-0.5">
                            <span
                              style={{
                                fontSize: 13,
                                color: "var(--echo-text-tertiary)",
                                lineHeight: "18px",
                              }}
                            >
                              {folder.conversationCount} conversation{folder.conversationCount !== 1 ? "s" : ""}
                            </span>
                            <span style={{ fontSize: 10, color: "var(--echo-border)" }}>|</span>
                            <span
                              style={{
                                fontSize: 12,
                                color: "var(--echo-text-tertiary)",
                              }}
                            >
                              {formatDate(folder.lastAccessedAt)}
                            </span>
                          </div>
                        </>
                      )}
                    </div>

                    <ChevronRight size={18} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>

                  {/* Action bar */}
                  <div
                    className="flex items-center border-t px-3 py-1.5"
                    style={{ borderColor: "var(--echo-border-light)" }}
                  >
                    <div className="flex items-center gap-1 flex-1">
                      <span
                        className="rounded-full px-2 py-0.5 flex items-center gap-1"
                        style={{
                          fontSize: 10,
                          fontWeight: 600,
                          color: tierDefaults[folder.securityTier].color,
                          backgroundColor:
                            folder.securityTier === "maximum"
                              ? "var(--echo-error-light)"
                              : folder.securityTier === "elevated"
                              ? "var(--echo-warning-light)"
                              : "var(--echo-success-light)",
                        }}
                      >
                        {tierDefaults[folder.securityTier].label}
                      </span>
                      <span
                        className="rounded-full px-2 py-0.5 flex items-center gap-1"
                        style={{
                          fontSize: 10,
                          fontWeight: 500,
                          color: "var(--echo-text-tertiary)",
                          backgroundColor: "var(--echo-bg-surface)",
                        }}
                      >
                        {formatBytes(folder.mediaUsedBytes)} / {formatBytes(folder.mediaQuotaBytes)}
                      </span>
                    </div>
                    <div className="flex items-center gap-0.5">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setEditingId(folder.id);
                          setEditName(folder.name);
                        }}
                        className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1.5 rounded-lg hover:bg-[var(--echo-bg-surface)]"
                      >
                        <Pencil size={14} style={{ color: "var(--echo-text-tertiary)" }} />
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setShowDeleteConfirm(folder.id);
                        }}
                        className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1.5 rounded-lg hover:bg-[var(--echo-error-light)]"
                      >
                        <Trash2 size={14} style={{ color: "var(--echo-error)" }} />
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </motion.div>
          ))}

          {/* Empty state */}
          {folders.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16">
              <div
                className="flex items-center justify-center rounded-full mb-4"
                style={{
                  width: 72,
                  height: 72,
                  backgroundColor: "var(--echo-primary-subtle)",
                }}
              >
                <FolderLock size={32} style={{ color: "var(--echo-primary)" }} />
              </div>
              <div
                style={{
                  fontSize: 16,
                  fontWeight: 600,
                  color: "var(--echo-text-primary)",
                  marginBottom: 4,
                }}
              >
                No hidden folders
              </div>
              <div
                style={{
                  fontSize: 13,
                  color: "var(--echo-text-tertiary)",
                  textAlign: "center",
                  maxWidth: 240,
                }}
              >
                Create a hidden folder to securely store sensitive conversations with biometric protection.
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Create button */}
      {folders.length < 10 && (
        <div className="px-4 pb-8 pt-3 shrink-0">
          <button
            onClick={() => setShowCreate(true)}
            className="w-full flex items-center justify-center gap-2 rounded-2xl border-none cursor-pointer py-3.5"
            style={{
              background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              boxShadow: "var(--echo-shadow-primary)",
            }}
          >
            <Plus size={20} color="white" />
            <span style={{ fontSize: 15, fontWeight: 600, color: "white" }}>
              Create Hidden Folder
            </span>
          </button>
          <div
            className="text-center mt-2"
            style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}
          >
            {folders.length}/10 folders used
          </div>
        </div>
      )}

      {/* Create Modal */}
      <AnimatePresence>
        {showCreate && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50"
              style={{ backgroundColor: "rgba(0,0,0,0.4)" }}
              onClick={() => setShowCreate(false)}
            />
            <motion.div
              initial={{ y: "100%", opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: "100%", opacity: 0 }}
              transition={{ type: "spring", damping: 28, stiffness: 350 }}
              className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
              style={{ backgroundColor: "var(--echo-bg-elevated)" }}
            >
              <div className="flex justify-center pt-3 pb-1">
                <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
              </div>

              <div className="px-5 pb-8">
                <div className="flex items-center justify-between mb-5">
                  <div
                    style={{
                      fontSize: 18,
                      fontWeight: 700,
                      color: "var(--echo-text-primary)",
                    }}
                  >
                    New Hidden Folder
                  </div>
                  <button
                    onClick={() => setShowCreate(false)}
                    className="bg-transparent border-none cursor-pointer p-1"
                  >
                    <X size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                </div>

                {/* Name input */}
                <div className="mb-5">
                  <label
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: "var(--echo-text-secondary)",
                      display: "block",
                      marginBottom: 6,
                    }}
                  >
                    Folder Name
                  </label>
                  <input
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                    placeholder="e.g., Personal, Legal, Finance"
                    className="w-full rounded-xl px-4 py-3 border-none outline-none"
                    style={{
                      fontSize: 15,
                      color: "var(--echo-text-primary)",
                      backgroundColor: "var(--echo-bg-surface)",
                    }}
                  />
                </div>

                {/* Security Tier */}
                <div className="mb-6">
                  <label
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: "var(--echo-text-secondary)",
                      display: "block",
                      marginBottom: 8,
                    }}
                  >
                    Security Tier
                  </label>
                  <div className="flex flex-col gap-2">
                    {(["standard", "elevated", "maximum"] as SecurityTier[]).map((tier) => {
                      const info = tierDefaults[tier];
                      const isSelected = newTier === tier;
                      return (
                        <button
                          key={tier}
                          onClick={() => setNewTier(tier)}
                          className="flex items-center gap-3 rounded-xl px-4 py-3 border-2 cursor-pointer text-left transition-all"
                          style={{
                            backgroundColor: isSelected
                              ? tier === "maximum"
                                ? "var(--echo-error-light)"
                                : tier === "elevated"
                                ? "var(--echo-warning-light)"
                                : "var(--echo-success-light)"
                              : "var(--echo-bg-surface)",
                            borderColor: isSelected ? info.color : "transparent",
                          }}
                        >
                          <span style={{ color: info.color }}>{tierIcons[tier]}</span>
                          <div className="flex-1">
                            <div
                              style={{
                                fontSize: 14,
                                fontWeight: 600,
                                color: "var(--echo-text-primary)",
                              }}
                            >
                              {info.label}
                            </div>
                            <div
                              style={{
                                fontSize: 12,
                                color: "var(--echo-text-tertiary)",
                              }}
                            >
                              {info.description}
                            </div>
                          </div>
                          {isSelected && (
                            <div
                              className="flex items-center justify-center rounded-full"
                              style={{
                                width: 22,
                                height: 22,
                                backgroundColor: info.color,
                              }}
                            >
                              <Check size={14} color="white" />
                            </div>
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>

                {/* Create button */}
                <button
                  onClick={handleCreate}
                  disabled={!newName.trim()}
                  className="w-full flex items-center justify-center gap-2 rounded-xl border-none cursor-pointer py-3.5 transition-opacity"
                  style={{
                    background: newName.trim()
                      ? "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))"
                      : "var(--echo-bg-surface)",
                    opacity: newName.trim() ? 1 : 0.5,
                  }}
                >
                  <FolderLock size={18} color={newName.trim() ? "white" : "var(--echo-text-tertiary)"} />
                  <span
                    style={{
                      fontSize: 15,
                      fontWeight: 600,
                      color: newName.trim() ? "white" : "var(--echo-text-tertiary)",
                    }}
                  >
                    Create Folder
                  </span>
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}