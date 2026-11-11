import React, { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { ArrowLeft, Users, MoreHorizontal, Send, Paperclip, Smile, ShieldCheck } from "lucide-react";
import { GroupAvatar } from "../components/echo-avatar";
import { MessageBubble, type Message } from "../components/message-bubble";
import { mockGroups, roleLabels, roleColors } from "../components/group-data";

// Mock group chat messages with member attribution
interface GroupMessage extends Message {
  senderName?: string;
  senderRole?: string;
  isVerified?: boolean;
}

const groupMessages: Record<string, GroupMessage[]> = {
  g1: [
    { id: "1", text: "Hey everyone! Just pushed the new authentication module. Can someone review?", time: "10:15 AM", sent: false, senderName: "Sarah Chen", senderRole: "owner", isVerified: true },
    { id: "2", text: "I'll take a look at it after lunch 👍", time: "10:18 AM", sent: false, senderName: "Marcus Johnson", senderRole: "admin", isVerified: true },
    { id: "3", text: "Great work Sarah! The architecture looks clean.", time: "10:22 AM", sent: true, status: "read" },
    { id: "4", text: "Thanks! Make sure to check the edge cases in the passkey flow", time: "10:24 AM", sent: false, senderName: "Sarah Chen", senderRole: "owner", isVerified: true },
    { id: "5", text: "Just shipped the new feature! 🚀", time: "10:30 AM", sent: false, senderName: "Chase Williams", senderRole: "member" },
  ],
  g2: [
    { id: "1", text: "New design system is live! Check it out at the shared link", time: "2:00 PM", sent: false, senderName: "Mia Davis", senderRole: "owner", isVerified: true },
    { id: "2", text: "Love the new color tokens, very comprehensive", time: "2:05 PM", sent: true, status: "read" },
    { id: "3", text: "The component library is beautifully organized", time: "2:08 PM", sent: false, senderName: "Liam Wilson", senderRole: "verified_member", isVerified: true },
  ],
  g3: [
    { id: "1", text: "ETH merge was a game changer for scalability", time: "9:00 AM", sent: false, senderName: "Raj Patel", senderRole: "owner", isVerified: true },
    { id: "2", text: "Agreed. Layer 2 solutions are really maturing now", time: "9:05 AM", sent: true, status: "read" },
  ],
  g5: [
    { id: "1", text: "Anyone up for dinner Friday?", time: "6:00 PM", sent: true, status: "read" },
    { id: "2", text: "I'm in! That new Italian place?", time: "6:05 PM", sent: false, senderName: "Sarah Chen", senderRole: "admin", isVerified: true },
    { id: "3", text: "Count me in too 🍝", time: "6:10 PM", sent: false, senderName: "Marcus Johnson", senderRole: "member", isVerified: true },
  ],
};

export function GroupChatPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const group = mockGroups.find((g) => g.id === id);
  const [newMessage, setNewMessage] = useState("");
  const [messages, setMessages] = useState<GroupMessage[]>(
    groupMessages[id || ""] || [
      { id: "1", text: "Welcome to the group!", time: "Now", sent: false, senderName: "System" },
    ]
  );

  if (!group) {
    return (
      <div className="flex flex-col items-center justify-center h-full">
        <p style={{ color: "var(--echo-text-tertiary)" }}>Group not found</p>
      </div>
    );
  }

  const handleSend = () => {
    if (!newMessage.trim()) return;
    setMessages([
      ...messages,
      {
        id: Date.now().toString(),
        text: newMessage.trim(),
        time: new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" }),
        sent: true,
        status: "sent",
      },
    ]);
    setNewMessage("");
  };

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
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          style={{ width: 44, height: 44 }}
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <button
          onClick={() => navigate(`/groups/${group.id}`)}
          className="flex items-center gap-3 flex-1 min-w-0 bg-transparent border-none cursor-pointer text-left p-0"
        >
          <GroupAvatar
            members={group.members.slice(0, 4).map((m) => ({ name: m.name }))}
            size={40}
          />
          <div className="flex-1 min-w-0">
            <div
              className="truncate"
              style={{
                fontSize: 16,
                fontWeight: 600,
                color: "var(--echo-text-primary)",
                lineHeight: "22px",
              }}
            >
              {group.name}
            </div>
            <div className="flex items-center gap-1">
              <Users size={12} style={{ color: "var(--echo-text-tertiary)" }} />
              <span
                style={{
                  fontSize: 12,
                  color: "var(--echo-text-tertiary)",
                  lineHeight: "16px",
                }}
              >
                {group.memberCount} members • {group.members.filter((m) => m.online).length} online
              </span>
            </div>
          </div>
        </button>
        <button
          onClick={() => navigate(`/groups/${group.id}`)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-2 rounded-full hover:bg-[var(--echo-bg-surface)]"
        >
          <MoreHorizontal size={20} style={{ color: "var(--echo-text-secondary)" }} />
        </button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto py-4" style={{ backgroundColor: "var(--echo-bg-secondary)" }}>
        {messages.map((msg) => (
          <div key={msg.id}>
            {/* Sender name for received messages */}
            {!msg.sent && msg.senderName && (
              <div className="flex items-center gap-1.5 px-5 mb-1" style={{ marginTop: 8 }}>
                <span
                  style={{
                    fontSize: 12,
                    fontWeight: 600,
                    color: msg.senderRole ? (roleColors as any)[msg.senderRole] || "var(--echo-text-secondary)" : "var(--echo-text-secondary)",
                  }}
                >
                  {msg.senderName}
                </span>
                {msg.isVerified && <ShieldCheck size={12} style={{ color: "var(--echo-success)" }} />}
                {msg.senderRole && msg.senderRole !== "member" && (
                  <span
                    className="rounded-full px-1.5 py-0"
                    style={{
                      fontSize: 9,
                      fontWeight: 700,
                      color: (roleColors as any)[msg.senderRole],
                      backgroundColor: `${(roleColors as any)[msg.senderRole]}15`,
                      lineHeight: "14px",
                    }}
                  >
                    {(roleLabels as any)[msg.senderRole]}
                  </span>
                )}
              </div>
            )}
            <MessageBubble message={msg} />
          </div>
        ))}
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
            border: "2px solid var(--echo-border)",
          }}
        >
          <input
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSend()}
            placeholder="Message the group..."
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
