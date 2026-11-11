import React, { useState, useRef, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  ArrowLeft, Phone, Video, Send, Paperclip, Smile,
  Clock, VolumeX, Timer, Settings,
} from "lucide-react";
import { EchoAvatar } from "../components/echo-avatar";
import { MessageBubble, type Message } from "../components/message-bubble";
import { MessageActionsSheet } from "../components/message-actions-sheet";
import { ScheduleMessageModal } from "../components/schedule-message-modal";
import { ChatSettingsSheet } from "../components/chat-settings-sheet";
import { motion, AnimatePresence } from "motion/react";

const chatData: Record<
  string,
  {
    name: string;
    online: boolean;
    gradient: "primary" | "secondary" | "warm" | "cool";
    timezone?: string;
    messages: Message[];
  }
> = {
  "1": {
    name: "Sarah Chen",
    online: true,
    gradient: "primary",
    timezone: "PST (UTC-8)",
    messages: [
      { id: "1", text: "Hey! Are you free this weekend?", time: "9:30 AM", sent: false },
      { id: "2", text: "I was thinking we could check out that new art exhibition", time: "9:31 AM", sent: false },
      { id: "3", text: "Oh that sounds amazing! Which one?", time: "9:35 AM", sent: true, status: "read" },
      {
        id: "4",
        text: "The modern art one at the downtown gallery. They have a new interactive installation!",
        time: "9:36 AM",
        sent: false,
      },
      {
        id: "5",
        text: "That sounds great! Let me check my schedule",
        time: "9:41 AM",
        sent: true,
        status: "delivered",
        isEdited: true,
      },
      {
        id: "6",
        text: "Reminder: Don't forget the tickets!",
        time: "10:00 AM",
        sent: true,
        status: "sent",
        isScheduled: true,
        scheduledTime: "Tomorrow 9 AM",
        conditionalTrigger: "time",
      },
    ],
  },
  "2": {
    name: "Marcus Johnson",
    online: true,
    gradient: "secondary",
    timezone: "EST (UTC-5)",
    messages: [
      { id: "1", text: "Just finished the Q4 deck", time: "2:10 PM", sent: false },
      { id: "2", text: "Nice! Can you share the link?", time: "2:15 PM", sent: true, status: "read" },
      { id: "3", text: "The presentation is ready for review", time: "2:20 PM", sent: false },
      {
        id: "4",
        text: "This message was recalled",
        time: "2:22 PM",
        sent: false,
        isRecalled: true,
      },
      {
        id: "5",
        text: "Will follow up after the meeting",
        time: "2:25 PM",
        sent: true,
        status: "failed",
        deliveryFailure: { type: "network", retryCount: 1, canRetry: true },
      },
    ],
  },
  "3": {
    name: "Elena Rodriguez",
    online: false,
    gradient: "warm",
    messages: [
      {
        id: "1",
        text: "This is a confidential discussion",
        time: "4:00 PM",
        sent: false,
        isSilent: true,
        isDisappearing: true,
        disappearTimer: 3540,
      },
      {
        id: "2",
        text: "Understood. Everything stays between us.",
        time: "4:02 PM",
        sent: true,
        status: "read",
        isSilent: true,
        isDisappearing: true,
        disappearTimer: 3420,
      },
      {
        id: "3",
        text: "I'll share the document via the secure channel",
        time: "4:05 PM",
        sent: false,
        isSilent: true,
        isDisappearing: true,
        disappearTimer: 3240,
      },
    ],
  },
};

const defaultChat = {
  name: "Contact",
  online: false,
  gradient: "primary" as const,
  messages: [
    { id: "1", text: "Hello there!", time: "10:00 AM", sent: false } as Message,
    { id: "2", text: "Hi! How are you?", time: "10:01 AM", sent: true, status: "read" as const } as Message,
  ],
};

export function ChatPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const chat = chatData[id || ""] || defaultChat;
  const [newMessage, setNewMessage] = useState("");
  const [messages, setMessages] = useState<Message[]>(chat.messages);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Modals
  const [selectedMessage, setSelectedMessage] = useState<Message | null>(null);
  const [showActions, setShowActions] = useState(false);
  const [showSchedule, setShowSchedule] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [showSendMenu, setShowSendMenu] = useState(false);

  // Chat-level settings
  const [silentMode, setSilentMode] = useState(id === "3");
  const [disappearingDuration, setDisappearingDuration] = useState(id === "3" ? 3600 : 0);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (opts?: { silent?: boolean }) => {
    if (!newMessage.trim()) return;
    const msg: Message = {
      id: Date.now().toString(),
      text: newMessage.trim(),
      time: new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" }),
      sent: true,
      status: "sending",
      isSilent: silentMode || opts?.silent,
      isDisappearing: disappearingDuration > 0,
      disappearTimer: disappearingDuration > 0 ? disappearingDuration : undefined,
    };
    setMessages([...messages, msg]);
    setNewMessage("");
    setShowSendMenu(false);
    // Simulate delivery
    setTimeout(() => {
      setMessages((prev) => prev.map((m) => (m.id === msg.id ? { ...m, status: "sent" } : m)));
    }, 500);
    setTimeout(() => {
      setMessages((prev) => prev.map((m) => (m.id === msg.id ? { ...m, status: "delivered" } : m)));
    }, 1500);
  };

  const handleScheduleSubmit = (config: { trigger: any; scheduledTime?: string; text: string }) => {
    const msg: Message = {
      id: Date.now().toString(),
      text: config.text || newMessage.trim(),
      time: new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" }),
      sent: true,
      status: "sent",
      isScheduled: true,
      scheduledTime: config.scheduledTime || undefined,
      conditionalTrigger: config.trigger,
      isSilent: silentMode,
    };
    setMessages([...messages, msg]);
    setNewMessage("");
  };

  const handleMessageAction = (action: "reply" | "edit" | "recall" | "schedule" | "copy" | "forward" | "pin") => {
    if (!selectedMessage) return;

    if (action === "recall") {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === selectedMessage.id ? { ...m, isRecalled: true, text: "" } : m
        )
      );
    } else if (action === "edit") {
      setNewMessage(selectedMessage.text);
      // Mark as editing (simplified)
      setMessages((prev) =>
        prev.map((m) =>
          m.id === selectedMessage.id ? { ...m, isEdited: true } : m
        )
      );
    } else if (action === "copy") {
      navigator.clipboard?.writeText(selectedMessage.text);
    }
    setSelectedMessage(null);
  };

  const handleRetry = (msg: Message) => {
    setMessages((prev) =>
      prev.map((m) =>
        m.id === msg.id
          ? { ...m, status: "sending", deliveryFailure: undefined }
          : m
      )
    );
    setTimeout(() => {
      setMessages((prev) =>
        prev.map((m) => (m.id === msg.id ? { ...m, status: "delivered" } : m))
      );
    }, 1200);
  };

  const isSilentConversation = id === "3";

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="flex items-center gap-3 px-3 shrink-0"
        style={{
          height: 96,
          paddingTop: 44,
          backgroundColor: "var(--echo-bg-elevated)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        <button
          onClick={() => navigate("/")}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <EchoAvatar name={chat.name} size="sm" online={chat.online} gradient={chat.gradient} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            <span
              style={{
                fontSize: 16,
                fontWeight: 600,
                color: "var(--echo-text-primary)",
                lineHeight: "22px",
              }}
            >
              {chat.name}
            </span>
            {silentMode && (
              <VolumeX size={13} style={{ color: "var(--echo-text-tertiary)" }} />
            )}
            {disappearingDuration > 0 && (
              <Timer size={13} style={{ color: "var(--echo-warning)" }} />
            )}
          </div>
          <div
            style={{
              fontSize: 13,
              fontWeight: 400,
              color: chat.online ? "var(--echo-success)" : "var(--echo-text-tertiary)",
              lineHeight: "18px",
            }}
          >
            {chat.online ? "Active now" : "Last seen 2h ago"}
          </div>
        </div>
        <div className="flex items-center gap-1">
          {[Phone, Video].map((Icon, i) => (
            <button
              key={i}
              className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2 rounded-full hover:bg-[var(--echo-bg-surface)]"
            >
              <Icon size={20} style={{ color: "var(--echo-text-secondary)" }} />
            </button>
          ))}
          <button
            onClick={() => setShowSettings(true)}
            className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2 rounded-full hover:bg-[var(--echo-bg-surface)]"
          >
            <Settings size={20} style={{ color: "var(--echo-text-secondary)" }} />
          </button>
        </div>
      </div>

      {/* Silent conversation banner */}
      {isSilentConversation && (
        <div
          className="flex items-center gap-2 px-4 py-2"
          style={{
            backgroundColor: "var(--echo-primary-subtle)",
            borderBottom: "1px solid var(--echo-primary-lighter)",
          }}
        >
          <VolumeX size={14} style={{ color: "var(--echo-primary)" }} />
          <span style={{ fontSize: 12, fontWeight: 500, color: "var(--echo-primary)", flex: 1 }}>
            Silent conversation — messages auto-delete after 1 hour
          </span>
          <Timer size={14} style={{ color: "var(--echo-warning)" }} />
        </div>
      )}

      {/* Messages */}
      <div className="flex-1 overflow-y-auto py-4" style={{ backgroundColor: "var(--echo-bg-secondary)" }}>
        {messages.map((msg) => (
          <MessageBubble
            key={msg.id}
            message={msg}
            onLongPress={(m) => {
              setSelectedMessage(m);
              setShowActions(true);
            }}
            onRetry={handleRetry}
          />
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div
        className="flex items-end gap-2 px-4 shrink-0"
        style={{
          paddingTop: 12,
          paddingBottom: 32,
          backgroundColor: "var(--echo-bg-elevated)",
          borderTop: "1px solid var(--echo-border-light)",
        }}
      >
        <button className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2">
          <Paperclip size={22} style={{ color: "var(--echo-text-tertiary)" }} />
        </button>
        <div
          className="flex-1 flex items-end rounded-3xl px-4"
          style={{
            minHeight: 44,
            backgroundColor: "var(--echo-bg-surface)",
            border: `2px solid ${silentMode ? "var(--echo-primary-lighter)" : "var(--echo-border)"}`,
          }}
        >
          <input
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && handleSend()}
            placeholder={silentMode ? "Silent message..." : "Type a message..."}
            className="flex-1 bg-transparent border-none outline-none py-2.5"
            style={{ fontSize: 16, color: "var(--echo-text-primary)" }}
          />
          <button className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2">
            <Smile size={22} style={{ color: "var(--echo-text-tertiary)" }} />
          </button>
        </div>

        {/* Send button with long-press for schedule */}
        <div className="relative">
          <button
            onClick={() => handleSend()}
            onContextMenu={(e) => {
              e.preventDefault();
              if (newMessage.trim()) setShowSendMenu(true);
            }}
            className="flex items-center justify-center rounded-full border-none cursor-pointer shrink-0"
            style={{
              width: 44,
              height: 44,
              background: newMessage.trim()
                ? "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))"
                : "var(--echo-bg-surface)",
            }}
          >
            <Send
              size={20}
              style={{ color: newMessage.trim() ? "white" : "var(--echo-text-tertiary)" }}
            />
          </button>

          {/* Send menu */}
          <AnimatePresence>
            {showSendMenu && (
              <>
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="fixed inset-0 z-40"
                  onClick={() => setShowSendMenu(false)}
                />
                <motion.div
                  initial={{ opacity: 0, y: 10, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 10, scale: 0.95 }}
                  className="absolute bottom-14 right-0 z-50 rounded-2xl overflow-hidden"
                  style={{
                    backgroundColor: "var(--echo-bg-elevated)",
                    boxShadow: "var(--echo-shadow-lg)",
                    border: "1px solid var(--echo-border)",
                    minWidth: 180,
                  }}
                >
                  <button
                    onClick={() => {
                      handleSend();
                      setShowSendMenu(false);
                    }}
                    className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer hover:bg-[var(--echo-bg-surface)] text-left"
                    style={{ borderBottom: "1px solid var(--echo-border-light)" }}
                  >
                    <Send size={16} style={{ color: "var(--echo-primary)" }} />
                    <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                      Send now
                    </span>
                  </button>
                  <button
                    onClick={() => {
                      handleSend({ silent: true });
                      setShowSendMenu(false);
                    }}
                    className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer hover:bg-[var(--echo-bg-surface)] text-left"
                    style={{ borderBottom: "1px solid var(--echo-border-light)" }}
                  >
                    <VolumeX size={16} style={{ color: "var(--echo-text-secondary)" }} />
                    <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                      Send silently
                    </span>
                  </button>
                  <button
                    onClick={() => {
                      setShowSendMenu(false);
                      setShowSchedule(true);
                    }}
                    className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer hover:bg-[var(--echo-bg-surface)] text-left"
                  >
                    <Clock size={16} style={{ color: "var(--echo-accent)" }} />
                    <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                      Schedule
                    </span>
                  </button>
                </motion.div>
              </>
            )}
          </AnimatePresence>
        </div>
      </div>

      {/* Message Actions Sheet */}
      <MessageActionsSheet
        open={showActions}
        onClose={() => {
          setShowActions(false);
          setSelectedMessage(null);
        }}
        message={selectedMessage}
        onAction={handleMessageAction}
      />

      {/* Schedule Message Modal */}
      <ScheduleMessageModal
        open={showSchedule}
        onClose={() => setShowSchedule(false)}
        onSchedule={handleScheduleSubmit}
        messageText={newMessage}
        contactName={chat.name}
      />

      {/* Chat Settings Sheet */}
      <ChatSettingsSheet
        open={showSettings}
        onClose={() => setShowSettings(false)}
        contactName={chat.name}
        silentMode={silentMode}
        onToggleSilent={() => setSilentMode(!silentMode)}
        disappearingEnabled={disappearingDuration > 0}
        disappearingDuration={disappearingDuration}
        onSetDisappearing={setDisappearingDuration}
        timezone={(chat as any).timezone}
      />
    </div>
  );
}