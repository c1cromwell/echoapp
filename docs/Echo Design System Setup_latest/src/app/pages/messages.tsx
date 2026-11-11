import React, { useState, useRef, useCallback } from "react";
import { useNavigate } from "react-router";
import { Search, Plus, ShieldAlert, X, Users, Compass, FolderLock, ChevronDown, Radio, Pin } from "lucide-react";
import { PinnedSection } from "../components/pinned-section";
import { ConversationItem, type Conversation } from "../components/conversation-item";
import { GroupsList } from "../components/groups-list";
import { mockGroups } from "../components/group-data";
import { ChannelsList } from "../components/channels-list";
import { mockChannels } from "../components/broadcast-channel-data";
import { PersonaSwitcher } from "../components/persona-switcher";
import { BiometricAuthModal } from "../components/biometric-auth-modal";
import { useHiddenFolders } from "../components/hidden-folder-data";
import { motion, AnimatePresence } from "motion/react";
import { ChatFolderFilter, FilterEmptyState, matchesFilter, type FolderFilter } from "../components/chat-folder-filter";
import { SecureContactsOnline } from "../components/secure-contacts-online";

const conversations: Conversation[] = [
  {
    id: "1",
    name: "Sarah Chen",
    lastMessage: "That sounds great! Let me check my schedule",
    time: "2m",
    unread: 3,
    online: true,
    gradient: "primary",
    hasScheduledMessages: true,
    verificationStatus: "identity_full",
  },
  {
    id: "2",
    name: "Marcus Johnson",
    lastMessage: "The presentation is ready for review",
    time: "15m",
    unread: 1,
    online: true,
    gradient: "secondary",
    hasDeliveryFailure: true,
    verificationStatus: "vip_pro",
  },
  {
    id: "3",
    name: "Elena Rodriguez",
    lastMessage: "I'll share the document via the secure channel",
    time: "45m",
    unread: 2,
    gradient: "warm",
    isSilentConversation: true,
    isDisappearing: true,
    verificationStatus: "identity_lite",
  },
  {
    id: "4",
    name: "Product Team",
    lastMessage: "Alex: Pushed the latest changes",
    time: "1h",
    unread: 12,
    type: "group",
    members: [
      { name: "Alex" },
      { name: "Beth" },
      { name: "Chris" },
      { name: "Dana" },
      { name: "Eve" },
      { name: "Frank" },
    ],
  },
  {
    id: "5",
    name: "James Park",
    lastMessage: "See you tomorrow at the gym!",
    time: "3h",
    online: false,
    gradient: "warm",
    verificationStatus: "phone_verified",
  },
  {
    id: "6",
    name: "Design Guild",
    lastMessage: "Mia: New design system is live!",
    time: "1d",
    type: "group",
    muted: true,
    members: [{ name: "Mia" }, { name: "Noah" }, { name: "Olivia" }],
  },
  {
    id: "7",
    name: "Chase Williams",
    lastMessage: "Let's catch up this weekend",
    time: "1d",
    online: true,
    gradient: "cool",
    verificationStatus: "identity_full",
  },
  {
    id: "8",
    name: "Ava Thompson",
    lastMessage: "The flight is booked for next Friday",
    time: "2d",
    gradient: "secondary",
  },
];

export function MessagesPage() {
  const navigate = useNavigate();
  const { authenticate, authenticateDuress, folders } = useHiddenFolders();
  const [searchQuery, setSearchQuery] = useState("");
  const [showSearchBar, setShowSearchBar] = useState(false);
  const [activeSegment, setActiveSegment] = useState<"chats" | "groups" | "pinned" | "channels">("chats");
  const [showNudge, setShowNudge] = useState(() => {
    return !localStorage.getItem("echo-nudge-dismissed");
  });
  const isOnboarded = localStorage.getItem("echo-onboarded") === "true";

  // Hidden folder access state
  const [showBiometric, setShowBiometric] = useState(false);
  const [pullProgress, setPullProgress] = useState(0);
  const [showPullHint, setShowPullHint] = useState(false);
  const pullStartY = useRef<number | null>(null);
  const holdTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [isHolding, setIsHolding] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Long-press on conversation for "Move to Hidden"
  const longPressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { moveConversationToHidden } = useHiddenFolders();
  const [showMoveSheet, setShowMoveSheet] = useState(false);
  const [moveTargetConv, setMoveTargetConv] = useState<string | null>(null);
  // Chat folder filter
  const [folderFilter, setFolderFilter] = useState<FolderFilter>("all");

  // Pull-down gesture handler
  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    if (scrollRef.current && scrollRef.current.scrollTop <= 0) {
      pullStartY.current = e.touches[0].clientY;
    }
  }, []);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (pullStartY.current === null) return;
    const deltaY = e.touches[0].clientY - pullStartY.current;
    if (deltaY > 0 && deltaY < 200) {
      const progress = Math.min(deltaY / 120, 1);
      setPullProgress(progress);
      if (progress >= 1 && !isHolding) {
        setIsHolding(true);
        setShowPullHint(true);
        holdTimer.current = setTimeout(() => {
          setPullProgress(0);
          setIsHolding(false);
          setShowPullHint(false);
          pullStartY.current = null;
          setShowBiometric(true);
        }, 1500);
      }
    }
  }, [isHolding]);

  const handleTouchEnd = useCallback(() => {
    pullStartY.current = null;
    setPullProgress(0);
    setIsHolding(false);
    setShowPullHint(false);
    if (holdTimer.current) {
      clearTimeout(holdTimer.current);
      holdTimer.current = null;
    }
  }, []);

  const handleBiometricSuccess = async () => {
    const success = await authenticate();
    if (success) {
      navigate("/hidden-folders");
    }
    return success;
  };

  // Long press handlers for conversations
  const handleConvTouchStart = (convId: string) => {
    longPressTimer.current = setTimeout(() => {
      setMoveTargetConv(convId);
      setShowMoveSheet(true);
    }, 600);
  };

  const handleConvTouchEnd = () => {
    if (longPressTimer.current) {
      clearTimeout(longPressTimer.current);
      longPressTimer.current = null;
    }
  };

  const handleMoveToHidden = (folderId: string) => {
    if (moveTargetConv) {
      moveConversationToHidden(moveTargetConv, folderId);
      setShowMoveSheet(false);
      setMoveTargetConv(null);
    }
  };

  const filtered = conversations.filter(
    (c) =>
      c.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      c.lastMessage.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Count total hidden conversations
  const hiddenConvCount = folders.reduce((sum, f) => sum + f.conversationCount, 0);

  return (
    <div className="flex flex-col h-full">
      {/* Pull-down indicator */}
      <AnimatePresence>
        {(pullProgress > 0.1 || showPullHint) && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: pullProgress * 60 + 10, opacity: pullProgress }}
            exit={{ height: 0, opacity: 0 }}
            className="flex items-center justify-center overflow-hidden shrink-0"
            style={{ backgroundColor: "var(--echo-primary-subtle)" }}
          >
            <div className="flex items-center gap-2">
              <motion.div
                animate={isHolding ? { rotate: 360 } : { rotate: 0 }}
                transition={isHolding ? { duration: 1.5, ease: "linear" } : {}}
              >
                <FolderLock
                  size={18}
                  style={{
                    color: pullProgress >= 1 ? "var(--echo-accent)" : "var(--echo-text-tertiary)",
                  }}
                />
              </motion.div>
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: pullProgress >= 1 ? "var(--echo-accent)" : "var(--echo-text-tertiary)",
                }}
              >
                {isHolding ? "Hold to access..." : pullProgress >= 0.8 ? "Keep pulling..." : "Pull down for hidden folders"}
              </span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Header */}
      <div className="px-4 pt-14 pb-2">
        {/* Persona switcher row */}
        <div className="flex items-center justify-between mb-2">
          <PersonaSwitcher compact />
        </div>

        <div className="flex items-center justify-between mb-4">
          <h1
            style={{
              fontSize: 28,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              lineHeight: "36px",
              letterSpacing: "-0.5px",
              margin: 0,
            }}
          >
            Messages
          </h1>
          <div className="flex items-center gap-2">
            {/* Search toggle */}
            <button
              onClick={() => {
                setShowSearchBar(!showSearchBar);
                if (showSearchBar) setSearchQuery("");
              }}
              className="flex items-center justify-center rounded-full border-none cursor-pointer"
              style={{
                width: 36,
                height: 36,
                backgroundColor: showSearchBar ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
              }}
              title="Search"
            >
              <Search size={18} style={{ color: showSearchBar ? "var(--echo-primary)" : "var(--echo-text-tertiary)" }} />
            </button>
            {/* Hidden folders quick access */}
            <button
              onClick={() => setShowBiometric(true)}
              className="flex items-center justify-center rounded-full border-none cursor-pointer"
              style={{
                width: 36,
                height: 36,
                backgroundColor: "var(--echo-bg-surface)",
              }}
              title="Hidden Folders"
            >
              <FolderLock size={18} style={{ color: "var(--echo-text-tertiary)" }} />
            </button>
            <button
              onClick={() => activeSegment === "groups" ? navigate("/groups/create") : activeSegment === "channels" ? navigate("/channels/create") : undefined}
              className="flex items-center justify-center rounded-full border-none cursor-pointer"
              style={{
                width: 36,
                height: 36,
                background: `linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))`,
              }}
            >
              <Plus size={20} color="white" strokeWidth={2.5} />
            </button>
          </div>
        </div>

        {/* Inline Search Bar (toggle) */}
        <AnimatePresence>
          {showSearchBar && (
            <motion.div
              initial={{ height: 0, opacity: 0, marginBottom: 0 }}
              animate={{ height: 44, opacity: 1, marginBottom: 12 }}
              exit={{ height: 0, opacity: 0, marginBottom: 0 }}
              transition={{ duration: 0.2 }}
              className="overflow-hidden"
            >
              <div
                className="flex items-center gap-2 rounded-xl px-4"
                style={{
                  height: 44,
                  backgroundColor: "var(--echo-bg-surface)",
                  border: "2px solid var(--echo-primary-lighter)",
                }}
              >
                <Search size={18} style={{ color: "var(--echo-primary)" }} />
                <input
                  type="text"
                  placeholder={activeSegment === "chats" ? "Search messages..." : activeSegment === "groups" ? "Search groups..." : activeSegment === "channels" ? "Search channels..." : "Search pinned..."}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="flex-1 bg-transparent border-none outline-none"
                  style={{
                    fontSize: 16,
                    color: "var(--echo-text-primary)",
                  }}
                  autoFocus
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery("")}
                    className="flex items-center justify-center bg-transparent border-none cursor-pointer p-0"
                  >
                    <X size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                )}
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Segment Toggle — now with Pinned tab */}
        <div
          className="flex rounded-xl p-1 mb-3"
          style={{ backgroundColor: "var(--echo-bg-surface)" }}
        >
          {(["chats", "groups", "pinned", "channels"] as const).map((seg) => (
            <button
              key={seg}
              onClick={() => setActiveSegment(seg)}
              className="flex-1 flex items-center justify-center rounded-lg border-none cursor-pointer py-2 transition-all"
              style={{
                backgroundColor: activeSegment === seg ? "var(--echo-bg-elevated)" : "transparent",
                boxShadow: activeSegment === seg ? "var(--echo-shadow-sm)" : "none",
                fontSize: 13,
                fontWeight: activeSegment === seg ? 600 : 400,
                color: activeSegment === seg ? "var(--echo-text-primary)" : "var(--echo-text-tertiary)",
              }}
            >
              {seg === "chats" ? "Chats" : seg === "groups" ? "Groups" : seg === "pinned" ? "Pinned" : "Channels"}
            </button>
          ))}
        </div>
      </div>

      {/* Secure Contacts Online — show above chats/pinned views */}
      {(activeSegment === "chats" || activeSegment === "pinned") && (
        <SecureContactsOnline />
      )}

      {/* Chat Folder Filter Bar — replaces old "ALL MESSAGES" divider */}
      {activeSegment === "chats" && (
        <ChatFolderFilter
          active={folderFilter}
          onChange={setFolderFilter}
          unreadCounts={{
            all: 0,
            verified: filtered
              .filter((c) => matchesFilter("verified", c.verificationStatus))
              .reduce((sum, c) => sum + (c.unread || 0), 0),
            trusted: filtered
              .filter((c) => matchesFilter("trusted", c.verificationStatus))
              .reduce((sum, c) => sum + (c.unread || 0), 0),
            unverified: filtered
              .filter((c) => matchesFilter("unverified", c.verificationStatus))
              .reduce((sum, c) => sum + (c.unread || 0), 0),
          }}
        />
      )}

      {/* Nudge card for skipped setup */}
      {showNudge && isOnboarded && (
        <div className="px-4 mb-2">
          <div
            className="flex items-center gap-3 rounded-2xl p-3"
            style={{
              background: "linear-gradient(135deg, var(--echo-primary-subtle), var(--echo-accent-subtle))",
              border: "1px solid var(--echo-primary-lighter)",
            }}
          >
            <div
              className="flex items-center justify-center rounded-xl shrink-0"
              style={{
                width: 40,
                height: 40,
                backgroundColor: "var(--echo-accent)",
              }}
            >
              <ShieldAlert size={20} color="white" />
            </div>
            <div className="flex-1 min-w-0">
              <div
                style={{
                  fontSize: 14,
                  fontWeight: 600,
                  color: "var(--echo-text-primary)",
                  lineHeight: "18px",
                }}
              >
                Complete your setup
              </div>
              <div
                style={{
                  fontSize: 12,
                  color: "var(--echo-text-secondary)",
                  lineHeight: "16px",
                  marginTop: 2,
                }}
              >
                Add recovery method to protect your account
              </div>
            </div>
            <button
              onClick={() => {
                setShowNudge(false);
                localStorage.setItem("echo-nudge-dismissed", "true");
              }}
              className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1 shrink-0"
            >
              <X size={16} style={{ color: "var(--echo-text-tertiary)" }} />
            </button>
          </div>
        </div>
      )}

      {/* Content */}
      {activeSegment === "chats" ? (() => {
        const folderFiltered = filtered.filter((c) => {
          if (c.type === "group") {
            if (folderFilter === "all") return true;
            return matchesFilter(folderFilter, c.verificationStatus);
          }
          return matchesFilter(folderFilter, c.verificationStatus);
        });

        if (folderFilter !== "all" && folderFiltered.length === 0) {
          return (
            <div className="flex-1 overflow-y-auto">
              <FilterEmptyState filter={folderFilter} />
            </div>
          );
        }

        return (
          <div
            ref={scrollRef}
            className="flex-1 overflow-y-auto"
            onTouchStart={handleTouchStart}
            onTouchMove={handleTouchMove}
            onTouchEnd={handleTouchEnd}
          >
            <AnimatePresence mode="wait">
              <motion.div
                key={folderFilter}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15 }}
              >
                {/* Recents label */}
                <h2
                  style={{
                    fontSize: 10,
                    fontWeight: 700,
                    letterSpacing: "0.2em",
                    textTransform: "uppercase" as const,
                    color: "var(--echo-text-tertiary)",
                    margin: "4px 0 8px",
                    paddingLeft: 18,
                  }}
                >
                  Recents
                </h2>
                {folderFiltered.map((conv) => (
                  <div
                    key={conv.id}
                    onTouchStart={() => handleConvTouchStart(conv.id)}
                    onTouchEnd={handleConvTouchEnd}
                    onMouseDown={() => handleConvTouchStart(conv.id)}
                    onMouseUp={handleConvTouchEnd}
                    onMouseLeave={handleConvTouchEnd}
                  >
                    <ConversationItem
                      conversation={conv}
                      onClick={() => navigate(`/chat/${conv.id}`)}
                    />
                  </div>
                ))}
              </motion.div>
            </AnimatePresence>
          </div>
        );
      })() : activeSegment === "pinned" ? (
        <div className="flex-1 overflow-y-auto">
          <PinnedSection />
        </div>
      ) : activeSegment === "groups" ? (
        <div className="flex-1 overflow-y-auto">
          {/* Discover banner */}
          <button
            onClick={() => navigate("/groups/discover")}
            className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
            style={{ borderBottom: "1px solid var(--echo-border-light)" }}
          >
            <div
              className="flex items-center justify-center rounded-xl shrink-0"
              style={{
                width: 44,
                height: 44,
                background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              }}
            >
              <Compass size={22} color="white" />
            </div>
            <div className="flex-1">
              <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                Discover Groups
              </div>
              <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
                Find public groups to join
              </div>
            </div>
          </button>
          <GroupsList
            groups={mockGroups.filter((g) =>
              g.name.toLowerCase().includes(searchQuery.toLowerCase())
            )}
          />
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto">
          {/* Discover banner */}
          <button
            onClick={() => navigate("/channels/discover")}
            className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
            style={{ borderBottom: "1px solid var(--echo-border-light)" }}
          >
            <div
              className="flex items-center justify-center rounded-xl shrink-0"
              style={{
                width: 44,
                height: 44,
                background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              }}
            >
              <Compass size={22} color="white" />
            </div>
            <div className="flex-1">
              <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                Discover Channels
              </div>
              <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
                Find public channels to join
              </div>
            </div>
          </button>
          <ChannelsList
            channels={mockChannels.filter((c) =>
              c.name.toLowerCase().includes(searchQuery.toLowerCase())
            )}
          />
        </div>
      )}

      {/* Biometric Auth Modal */}
      <BiometricAuthModal
        open={showBiometric}
        onClose={() => setShowBiometric(false)}
        onAuthenticate={handleBiometricSuccess}
        onDuressPin={(pin) => {
          const success = authenticateDuress(pin);
          if (success) {
            setTimeout(() => navigate("/hidden-folders"), 600);
          }
          return success;
        }}
        title="Hidden Folders"
        subtitle="Authenticate to access your hidden conversations"
      />

      {/* Move to Hidden Folder Sheet */}
      <AnimatePresence>
        {showMoveSheet && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50"
              style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
              onClick={() => {
                setShowMoveSheet(false);
                setMoveTargetConv(null);
              }}
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
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <FolderLock size={20} style={{ color: "var(--echo-primary)" }} />
                    <span style={{ fontSize: 17, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                      Move to Hidden Folder
                    </span>
                  </div>
                  <button
                    onClick={() => {
                      setShowMoveSheet(false);
                      setMoveTargetConv(null);
                    }}
                    className="bg-transparent border-none cursor-pointer p-1"
                  >
                    <X size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                </div>

                <div
                  className="rounded-xl px-3 py-2 mb-4"
                  style={{ backgroundColor: "var(--echo-bg-surface)" }}
                >
                  <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                    Messages will be re-encrypted with AES-256-GCM and hidden from the main chat list, search, and notifications.
                  </span>
                </div>

                {folders.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    {folders.map((folder) => (
                      <button
                        key={folder.id}
                        onClick={() => handleMoveToHidden(folder.id)}
                        className="flex items-center gap-3 rounded-xl px-4 py-3 border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
                        style={{ backgroundColor: "var(--echo-bg-surface)" }}
                      >
                        <FolderLock size={18} style={{ color: "var(--echo-primary)" }} />
                        <div className="flex-1">
                          <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                            {folder.name}
                          </div>
                          <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                            {folder.conversationCount} conversations
                          </div>
                        </div>
                      </button>
                    ))}
                  </div>
                ) : (
                  <div className="flex flex-col items-center py-6">
                    <FolderLock size={32} style={{ color: "var(--echo-text-tertiary)", marginBottom: 8 }} />
                    <span style={{ fontSize: 14, color: "var(--echo-text-tertiary)" }}>
                      No hidden folders yet
                    </span>
                  </div>
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}