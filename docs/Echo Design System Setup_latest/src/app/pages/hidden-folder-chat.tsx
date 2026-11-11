import React, { useState, useRef, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  ArrowLeft, Lock, Send, Paperclip, Smile, VolumeX, Timer,
  EyeOff, Ban, Shield,
} from "lucide-react";
import { EchoAvatar } from "../components/echo-avatar";
import { MessageBubble, type Message } from "../components/message-bubble";
import { useHiddenFolders } from "../components/hidden-folder-data";
import { motion } from "motion/react";

export function HiddenFolderChatPage() {
  const navigate = useNavigate();
  const { id, chatId } = useParams();
  const { folders, lock } = useHiddenFolders();
  const folder = folders.find((f) => f.id === id);
  const conversation = folder?.conversations.find((c) => c.id === chatId);

  const [newMessage, setNewMessage] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (conversation) {
      setMessages(
        conversation.messages.map((m) => ({
          id: m.id,
          text: m.text,
          time: m.time,
          sent: m.sent,
          status: m.status as Message["status"],
          isDisappearing: m.isDisappearing,
          disappearTimer: m.disappearTimer,
          isSilent: m.isSilent,
        }))
      );
    }
  }, [conversation]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  if (!folder || !conversation) {
    return (
      <div className="flex flex-col h-full items-center justify-center">
        <EyeOff size={48} style={{ color: "var(--echo-text-tertiary)", marginBottom: 12 }} />
        <div style={{ fontSize: 16, color: "var(--echo-text-tertiary)" }}>Conversation not found</div>
      </div>
    );
  }

  const handleSend = () => {
    if (!newMessage.trim()) return;
    const msg: Message = {
      id: Date.now().toString(),
      text: newMessage.trim(),
      time: new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" }),
      sent: true,
      status: "sending",
    };
    setMessages((prev) => [...prev, msg]);
    setNewMessage("");
    setTimeout(() => {
      setMessages((prev) =>
        prev.map((m) => (m.id === msg.id ? { ...m, status: "sent" } : m))
      );
    }, 500);
    setTimeout(() => {
      setMessages((prev) =>
        prev.map((m) => (m.id === msg.id ? { ...m, status: "delivered" } : m))
      );
    }, 1500);
  };

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
        {/* Hidden folder badge */}
        <div
          className="flex items-center justify-center gap-1.5 py-1"
          style={{
            backgroundColor: "var(--echo-primary-subtle)",
            borderBottom: "1px solid var(--echo-primary-lighter)",
            paddingTop: 44,
          }}
        >
          <Shield size={10} style={{ color: "var(--echo-primary)" }} />
          <span
            style={{
              fontSize: 10,
              fontWeight: 600,
              color: "var(--echo-primary)",
              letterSpacing: "0.5px",
            }}
          >
            HIDDEN FOLDER &middot; {folder.name.toUpperCase()} &middot; AES-256-GCM
          </span>
        </div>

        <div className="flex items-center gap-3 px-3 py-2">
          <button
            onClick={() => navigate(`/hidden-folders/${folder.id}`)}
            className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          >
            <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
          </button>
          <EchoAvatar
            name={conversation.name}
            size="sm"
            online={conversation.online}
            gradient={conversation.gradient}
          />
          <div className="flex-1 min-w-0">
            <span
              style={{
                fontSize: 16,
                fontWeight: 600,
                color: "var(--echo-text-primary)",
                lineHeight: "22px",
              }}
            >
              {conversation.name}
            </span>
            <div
              style={{
                fontSize: 12,
                color: conversation.online ? "var(--echo-success)" : "var(--echo-text-tertiary)",
                lineHeight: "16px",
              }}
            >
              {conversation.online ? "Active now" : "Last seen recently"}
            </div>
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
      </div>

      {/* Restrictions banner */}
      <div
        className="flex items-center gap-2 px-4 py-2"
        style={{
          backgroundColor: "var(--echo-bg-surface)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        <Ban size={12} style={{ color: "var(--echo-text-tertiary)" }} />
        <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", flex: 1 }}>
          Forward, share & export disabled
        </span>
        <EyeOff size={12} style={{ color: "var(--echo-text-tertiary)" }} />
        <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
          Screenshot blocked
        </span>
      </div>

      {/* Messages */}
      <div
        className="flex-1 overflow-y-auto py-4"
        style={{ backgroundColor: "var(--echo-bg-secondary)" }}
      >
        {/* Encryption notice */}
        <div className="flex justify-center mb-4">
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex items-center gap-1.5 rounded-full px-3 py-1.5"
            style={{
              backgroundColor: "var(--echo-primary-subtle)",
              border: "1px solid var(--echo-primary-lighter)",
            }}
          >
            <Lock size={11} style={{ color: "var(--echo-primary)" }} />
            <span style={{ fontSize: 11, fontWeight: 500, color: "var(--echo-primary)" }}>
              Layer 2 encrypted &middot; Biometric-bound
            </span>
          </motion.div>
        </div>

        {messages.map((msg) => (
          <MessageBubble key={msg.id} message={msg} />
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
            border: "2px solid var(--echo-primary-lighter)",
          }}
        >
          <input
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && handleSend()}
            placeholder="Encrypted message..."
            className="flex-1 bg-transparent border-none outline-none py-2.5"
            style={{ fontSize: 16, color: "var(--echo-text-primary)" }}
          />
          <button className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2">
            <Smile size={22} style={{ color: "var(--echo-text-tertiary)" }} />
          </button>
        </div>
        <button
          onClick={handleSend}
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
      </div>
    </div>
  );
}
