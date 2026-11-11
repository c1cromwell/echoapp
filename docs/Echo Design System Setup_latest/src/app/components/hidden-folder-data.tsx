import React, { createContext, useContext, useState, useCallback } from "react";

// ─── Types ───────────────────────────────────────────────────────────────────

export type SecurityTier = "standard" | "elevated" | "maximum";
export type NotificationMode = "suppressed" | "redacted" | "unlockedOnly";

export interface HiddenFolder {
  id: string;
  name: string;
  createdAt: string;
  lastAccessedAt: string;
  securityTier: SecurityTier;
  notificationMode: NotificationMode;
  screenshotProtection: boolean;
  autoLockTimeout: number; // seconds
  wipeLockoutThreshold: number;
  isDecoy: boolean;
  conversationCount: number;
  mediaUsedBytes: number;
  mediaQuotaBytes: number;
  conversations: HiddenConversation[];
}

export interface HiddenConversation {
  id: string;
  name: string;
  lastMessage: string;
  time: string;
  unread: number;
  online: boolean;
  gradient: "primary" | "secondary" | "warm" | "cool";
  messages: HiddenMessage[];
}

export interface HiddenMessage {
  id: string;
  text: string;
  time: string;
  sent: boolean;
  status?: "sending" | "sent" | "delivered" | "read";
  isDisappearing?: boolean;
  disappearTimer?: number;
  isSilent?: boolean;
}

// ─── Security Tier Defaults ──────────────────────────────────────────────────

export const tierDefaults: Record<
  SecurityTier,
  {
    label: string;
    description: string;
    autoLockTimeout: number;
    wipeLockoutThreshold: number;
    screenshotProtection: boolean;
    color: string;
    icon: string;
  }
> = {
  standard: {
    label: "Standard",
    description: "5 min auto-lock, wipe disabled",
    autoLockTimeout: 300,
    wipeLockoutThreshold: 0,
    screenshotProtection: true,
    color: "var(--echo-success)",
    icon: "shield",
  },
  elevated: {
    label: "Elevated",
    description: "30s auto-lock, wipe at 15 attempts",
    autoLockTimeout: 30,
    wipeLockoutThreshold: 15,
    screenshotProtection: true,
    color: "var(--echo-warning)",
    icon: "shield-alert",
  },
  maximum: {
    label: "Maximum",
    description: "Immediate lock, wipe at 10 attempts",
    autoLockTimeout: 0,
    wipeLockoutThreshold: 10,
    screenshotProtection: true,
    color: "var(--echo-error)",
    icon: "shield-off",
  },
};

// ─── Mock Data ───────────────────────────────────────────────────────────────

const initialFolders: HiddenFolder[] = [
  {
    id: "hf-1",
    name: "Personal",
    createdAt: "2026-02-15T10:00:00Z",
    lastAccessedAt: "2026-03-03T08:30:00Z",
    securityTier: "elevated",
    notificationMode: "suppressed",
    screenshotProtection: true,
    autoLockTimeout: 30,
    wipeLockoutThreshold: 15,
    isDecoy: false,
    conversationCount: 2,
    mediaUsedBytes: 524288000,
    mediaQuotaBytes: 2147483648,
    conversations: [
      {
        id: "hc-1",
        name: "Jordan Rivera",
        lastMessage: "The documents are ready for signing",
        time: "8:30 AM",
        unread: 1,
        online: false,
        gradient: "warm",
        messages: [
          { id: "hm-1", text: "Hey, I need to discuss something sensitive", time: "Yesterday 6:00 PM", sent: false, status: "read" },
          { id: "hm-2", text: "Of course, go ahead", time: "Yesterday 6:02 PM", sent: true, status: "read" },
          { id: "hm-3", text: "The merger details are finalized. Attached the NDA.", time: "Yesterday 6:05 PM", sent: false, status: "read" },
          { id: "hm-4", text: "Got it. Reviewing now.", time: "Yesterday 6:10 PM", sent: true, status: "read" },
          { id: "hm-5", text: "The documents are ready for signing", time: "8:30 AM", sent: false, isDisappearing: true, disappearTimer: 2400 },
        ],
      },
      {
        id: "hc-2",
        name: "Dr. Amara Singh",
        lastMessage: "Your results look normal",
        time: "Yesterday",
        unread: 0,
        online: true,
        gradient: "cool",
        messages: [
          { id: "hm-6", text: "Hi, just checking on those test results", time: "Mon 2:00 PM", sent: true, status: "read" },
          { id: "hm-7", text: "Your results look normal. I'll send the full report shortly.", time: "Mon 4:30 PM", sent: false, status: "read" },
          { id: "hm-8", text: "That's a relief, thank you!", time: "Mon 4:32 PM", sent: true, status: "read" },
        ],
      },
    ],
  },
  {
    id: "hf-2",
    name: "Finance",
    createdAt: "2026-02-20T14:00:00Z",
    lastAccessedAt: "2026-03-02T16:45:00Z",
    securityTier: "maximum",
    notificationMode: "suppressed",
    screenshotProtection: true,
    autoLockTimeout: 0,
    wipeLockoutThreshold: 10,
    isDecoy: false,
    conversationCount: 1,
    mediaUsedBytes: 102400000,
    mediaQuotaBytes: 2147483648,
    conversations: [
      {
        id: "hc-3",
        name: "Alex Whitfield (Advisor)",
        lastMessage: "Portfolio rebalance scheduled for Friday",
        time: "4:45 PM",
        unread: 2,
        online: false,
        gradient: "secondary",
        messages: [
          { id: "hm-9", text: "Market update: tech sector is up 3.2% this week", time: "3:00 PM", sent: false, status: "read" },
          { id: "hm-10", text: "Should we adjust the allocation?", time: "3:15 PM", sent: true, status: "read" },
          { id: "hm-11", text: "I recommend shifting 5% from bonds to growth. Here's the analysis.", time: "4:00 PM", sent: false, isSilent: true },
          { id: "hm-12", text: "Portfolio rebalance scheduled for Friday", time: "4:45 PM", sent: false, isDisappearing: true, disappearTimer: 1800 },
        ],
      },
    ],
  },
  {
    id: "hf-3",
    name: "Decoy",
    createdAt: "2026-02-15T10:05:00Z",
    lastAccessedAt: "2026-03-01T12:00:00Z",
    securityTier: "standard",
    notificationMode: "suppressed",
    screenshotProtection: false,
    autoLockTimeout: 300,
    wipeLockoutThreshold: 0,
    isDecoy: true,
    conversationCount: 1,
    mediaUsedBytes: 0,
    mediaQuotaBytes: 2147483648,
    conversations: [
      {
        id: "hc-4",
        name: "Shopping List",
        lastMessage: "Don't forget to pick up groceries",
        time: "Sat",
        unread: 0,
        online: false,
        gradient: "primary",
        messages: [
          { id: "hm-13", text: "Don't forget to pick up groceries", time: "Sat 10:00 AM", sent: true, status: "read" },
        ],
      },
    ],
  },
];

// ─── Context ─────────────────────────────────────────────────────────────────

interface HiddenFolderContextType {
  folders: HiddenFolder[];
  isAuthenticated: boolean;
  isDuressMode: boolean;
  failedAttempts: number;
  authenticate: () => Promise<boolean>;
  authenticateDuress: (pin: string) => boolean;
  lock: () => void;
  createFolder: (name: string, tier: SecurityTier) => void;
  deleteFolder: (id: string) => void;
  renameFolder: (id: string, name: string) => void;
  updateFolderSettings: (id: string, updates: Partial<HiddenFolder>) => void;
  moveConversationToHidden: (conversationId: string, folderId: string) => void;
  moveConversationOut: (conversationId: string, folderId: string) => void;
}

const HiddenFolderContext = createContext<HiddenFolderContextType>({
  folders: [],
  isAuthenticated: false,
  isDuressMode: false,
  failedAttempts: 0,
  authenticate: async () => false,
  authenticateDuress: () => false,
  lock: () => {},
  createFolder: () => {},
  deleteFolder: () => {},
  renameFolder: () => {},
  updateFolderSettings: () => {},
  moveConversationToHidden: () => {},
  moveConversationOut: () => {},
});

export function useHiddenFolders() {
  return useContext(HiddenFolderContext);
}

export function HiddenFolderProvider({ children }: { children: React.ReactNode }) {
  const [folders, setFolders] = useState<HiddenFolder[]>(initialFolders);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isDuressMode, setIsDuressMode] = useState(false);
  const [failedAttempts, setFailedAttempts] = useState(0);

  const authenticate = useCallback(async (): Promise<boolean> => {
    // Simulate biometric success after a brief delay
    return new Promise((resolve) => {
      setTimeout(() => {
        setIsAuthenticated(true);
        setIsDuressMode(false);
        setFailedAttempts(0);
        resolve(true);
      }, 1500);
    });
  }, []);

  const authenticateDuress = useCallback(
    (pin: string): boolean => {
      if (pin === "1234") {
        // Mock duress PIN
        setIsAuthenticated(true);
        setIsDuressMode(true);
        return true;
      }
      setFailedAttempts((prev) => prev + 1);
      return false;
    },
    []
  );

  const lock = useCallback(() => {
    setIsAuthenticated(false);
    setIsDuressMode(false);
  }, []);

  const createFolder = useCallback((name: string, tier: SecurityTier) => {
    const defaults = tierDefaults[tier];
    const newFolder: HiddenFolder = {
      id: `hf-${Date.now()}`,
      name,
      createdAt: new Date().toISOString(),
      lastAccessedAt: new Date().toISOString(),
      securityTier: tier,
      notificationMode: "suppressed",
      screenshotProtection: defaults.screenshotProtection,
      autoLockTimeout: defaults.autoLockTimeout,
      wipeLockoutThreshold: defaults.wipeLockoutThreshold,
      isDecoy: false,
      conversationCount: 0,
      mediaUsedBytes: 0,
      mediaQuotaBytes: 2147483648,
      conversations: [],
    };
    setFolders((prev) => [...prev, newFolder]);
  }, []);

  const deleteFolder = useCallback((id: string) => {
    setFolders((prev) => prev.filter((f) => f.id !== id));
  }, []);

  const renameFolder = useCallback((id: string, name: string) => {
    setFolders((prev) =>
      prev.map((f) => (f.id === id ? { ...f, name } : f))
    );
  }, []);

  const updateFolderSettings = useCallback((id: string, updates: Partial<HiddenFolder>) => {
    setFolders((prev) =>
      prev.map((f) => (f.id === id ? { ...f, ...updates } : f))
    );
  }, []);

  const moveConversationToHidden = useCallback(
    (conversationId: string, folderId: string) => {
      // In a real app, this would migrate messages with re-encryption
      setFolders((prev) =>
        prev.map((f) => {
          if (f.id !== folderId) return f;
          const alreadyExists = f.conversations.some((c) => c.id === conversationId);
          if (alreadyExists) return f;
          const newConv: HiddenConversation = {
            id: conversationId,
            name: `Contact ${conversationId}`,
            lastMessage: "Conversation moved to hidden folder",
            time: "Now",
            unread: 0,
            online: false,
            gradient: "primary",
            messages: [],
          };
          return {
            ...f,
            conversations: [...f.conversations, newConv],
            conversationCount: f.conversationCount + 1,
          };
        })
      );
    },
    []
  );

  const moveConversationOut = useCallback(
    (conversationId: string, folderId: string) => {
      setFolders((prev) =>
        prev.map((f) => {
          if (f.id !== folderId) return f;
          return {
            ...f,
            conversations: f.conversations.filter((c) => c.id !== conversationId),
            conversationCount: Math.max(0, f.conversationCount - 1),
          };
        })
      );
    },
    []
  );

  const visibleFolders = isDuressMode
    ? folders.filter((f) => f.isDecoy)
    : folders.filter((f) => !f.isDecoy);

  return (
    <HiddenFolderContext.Provider
      value={{
        folders: visibleFolders,
        isAuthenticated,
        isDuressMode,
        failedAttempts,
        authenticate,
        authenticateDuress,
        lock,
        createFolder,
        deleteFolder,
        renameFolder,
        updateFolderSettings,
        moveConversationToHidden,
        moveConversationOut,
      }}
    >
      {children}
    </HiddenFolderContext.Provider>
  );
}
