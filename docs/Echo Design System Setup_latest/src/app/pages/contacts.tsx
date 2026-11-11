import React, { useState } from "react";
import { useNavigate } from "react-router";
import { Search, UserPlus, ChevronRight, ShieldCheck, Settings } from "lucide-react";
import { ImageWithFallback } from "../components/figma/ImageWithFallback";

type TrustTier = "inner-circle" | "trusted" | "verified" | "peer" | "basic";

interface Contact {
  id: string;
  name: string;
  avatar: string;
  online: boolean;
  trustTier: TrustTier;
  isVerified?: boolean;
}

interface GroupItem {
  id: string;
  name: string;
  icon: string;
  members: number;
  status?: string;
  isLocked?: boolean;
  memberAvatars?: string[];
}

const tierStyles: Record<TrustTier, { textVar: string; bgVar: string; borderVar: string; label: string }> = {
  "inner-circle": { textVar: "--echo-tier-inner-text", bgVar: "--echo-tier-inner-bg", borderVar: "--echo-tier-inner-border", label: "Inner Circle" },
  "trusted": { textVar: "--echo-tier-trusted-text", bgVar: "--echo-tier-trusted-bg", borderVar: "--echo-tier-trusted-border", label: "Trusted" },
  "verified": { textVar: "--echo-tier-verified-text", bgVar: "--echo-tier-verified-bg", borderVar: "--echo-tier-verified-border", label: "Verified" },
  "peer": { textVar: "--echo-tier-trusted-text", bgVar: "--echo-tier-trusted-bg", borderVar: "--echo-tier-trusted-border", label: "Peer" },
  "basic": { textVar: "--echo-tier-trusted-text", bgVar: "--echo-tier-trusted-bg", borderVar: "--echo-tier-trusted-border", label: "Basic" },
};

const contacts: Contact[] = [
  { id: "1", name: "Alice Eth", avatar: "https://images.unsplash.com/photo-1769636929388-99eff95d3bf1?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHx3b21hbiUyMHByb2Zlc3Npb25hbCUyMGhlYWRzaG90JTIwc3R1ZGlvfGVufDF8fHx8MTc3NTEzOTg0MHww&ixlib=rb-4.1.0&q=80&w=1080", online: true, trustTier: "inner-circle", isVerified: true },
  { id: "2", name: "Bob", avatar: "https://images.unsplash.com/photo-1543132220-e7fef0b974e7?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxtYW4lMjBjYXN1YWwlMjBwb3J0cmFpdCUyMHN0dWRpb3xlbnwxfHx8fDE3NzUxMzk4NDB8MA&ixlib=rb-4.1.0&q=80&w=1080", online: false, trustTier: "trusted" },
  { id: "3", name: "Dev", avatar: "https://images.unsplash.com/photo-1563237023-b1e970526dcb?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxkZXZlbG9wZXIlMjBwb3J0cmFpdCUyMGhlYWRzaG90JTIwbWluaW1hbHxlbnwxfHx8fDE3NzUxMzk4NDB8MA&ixlib=rb-4.1.0&q=80&w=1080", online: false, trustTier: "verified" },
  { id: "4", name: "Sarah Chen", avatar: "https://images.unsplash.com/photo-1655249493799-9cee4fe983bb?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcm9mZXNzaW9uYWwlMjB3b21hbiUyMHBvcnRyYWl0JTIwaGVhZHNob3R8ZW58MXx8fHwxNzc0ODgwNDc0fDA&ixlib=rb-4.1.0&q=80&w=1080", online: true, trustTier: "inner-circle", isVerified: true },
  { id: "5", name: "Marcus J.", avatar: "https://images.unsplash.com/photo-1769636930047-4478f12cf430?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcm9mZXNzaW9uYWwlMjBtYW4lMjBnbGFzc2VzJTIwcG9ydHJhaXR8ZW58MXx8fHwxNzc0ODk0NDQwfDA&ixlib=rb-4.1.0&q=80&w=1080", online: true, trustTier: "peer" },
  { id: "6", name: "Elena R.", avatar: "https://images.unsplash.com/photo-1763677594421-f58e50cce64d?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHx3b21hbiUyMGVsZWdhbnQlMjBwb3J0cmFpdCUyMG1pbmltYWwlMjBiYWNrZ3JvdW5kfGVufDF8fHx8MTc3NTEzOTg0MXww&ixlib=rb-4.1.0&q=80&w=1080", online: false, trustTier: "verified" },
];

const groups: GroupItem[] = [
  { id: "g1", name: "Constellation HQ", icon: "hub", members: 12, isLocked: true, memberAvatars: ["https://images.unsplash.com/photo-1769636929388-99eff95d3bf1?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHx3b21hbiUyMHByb2Zlc3Npb25hbCUyMGhlYWRzaG90JTIwc3R1ZGlvfGVufDF8fHx8MTc3NTEzOTg0MHww&ixlib=rb-4.1.0&q=80&w=1080", "https://images.unsplash.com/photo-1543132220-e7fef0b974e7?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxtYW4lMjBjYXN1YWwlMjBwb3J0cmFpdCUyMHN0dWRpb3xlbnwxfHx8fDE3NzUxMzk4NDB8MA&ixlib=rb-4.1.0&q=80&w=1080"] },
  { id: "g2", name: "Node Runners", icon: "terminal", members: 4, status: "SYNCING" },
];

function GroupCard({ group, onClick }: { group: GroupItem; onClick: () => void }) {
  return (
    <div
      onClick={onClick}
      className="flex items-center justify-between rounded-3xl p-4 cursor-pointer transition-all"
      style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}
    >
      <div className="flex items-center gap-4">
        <div
          className="flex items-center justify-center rounded-2xl"
          style={{ width: 56, height: 56, backgroundColor: "var(--echo-bg-elevated)", border: "1px solid var(--echo-ghost-border)" }}
        >
          <span style={{ fontSize: 24, color: group.id === "g1" ? "var(--echo-sky)" : "var(--echo-text-secondary)" }}>
            {group.icon === "hub" ? "⬡" : "⌘"}
          </span>
        </div>
        <div>
          <div className="flex items-center gap-2">
            <span style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-text-primary)" }}>{group.name}</span>
            {group.isLocked && <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>🔒</span>}
          </div>
          <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
            {group.status ? `${group.members} active protocols` : `${group.members} nodes connected`}
          </span>
        </div>
      </div>
      <div className="flex items-center">
        {group.memberAvatars ? (
          <div className="flex -space-x-2">
            {group.memberAvatars.map((av, i) => (
              <ImageWithFallback key={i} src={av} alt="" className="rounded-full object-cover" style={{ width: 28, height: 28, border: "2px solid var(--echo-bg-surface)" }} />
            ))}
            <div className="flex items-center justify-center rounded-full" style={{ width: 28, height: 28, backgroundColor: "var(--echo-sky)", border: "2px solid var(--echo-bg-surface)", fontSize: 10, fontWeight: 700, color: "white" }}>
              +{group.members - 2}
            </div>
          </div>
        ) : group.status ? (
          <span className="rounded px-2 py-0.5" style={{ fontSize: 10, fontWeight: 700, color: "var(--echo-tier-inner-text)", backgroundColor: "var(--echo-tier-inner-bg)", border: "1px solid var(--echo-tier-inner-border)" }}>
            {group.status}
          </span>
        ) : null}
      </div>
    </div>
  );
}

export function ContactsPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [activeTab, setActiveTab] = useState<"contacts" | "groups">("contacts");

  const onlineCount = contacts.filter((c) => c.online).length;
  const filteredContacts = contacts.filter((c) => c.name.toLowerCase().includes(search.toLowerCase()));
  const filteredGroups = groups.filter((g) => g.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <div className="flex flex-col h-full" style={{ backgroundColor: "var(--echo-bg-primary)" }}>
      <div className="px-5 pt-14 pb-2">
        <div className="flex items-center justify-between mb-4">
          <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--echo-text-primary)", letterSpacing: "-0.5px", margin: 0 }}>
            Contacts
          </h1>
          <div className="flex items-center gap-2">
            <button className="flex items-center justify-center rounded-xl border-none cursor-pointer" style={{ width: 40, height: 40, backgroundColor: "transparent" }}>
              <UserPlus size={22} style={{ color: "var(--echo-accent)" }} />
            </button>
            <button onClick={() => navigate("/settings")} className="flex items-center justify-center rounded-xl border-none cursor-pointer" style={{ width: 40, height: 40, backgroundColor: "transparent" }}>
              <Settings size={22} style={{ color: "var(--echo-text-tertiary)" }} />
            </button>
          </div>
        </div>

        <div className="flex items-center gap-3 rounded-xl px-4" style={{ height: 48, backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}>
          <Search size={18} style={{ color: "var(--echo-text-tertiary)" }} />
          <input type="text" placeholder="Search decentralized nodes..." value={search} onChange={(e) => setSearch(e.target.value)} className="flex-1 bg-transparent border-none outline-none" style={{ fontSize: 15, color: "var(--echo-text-primary)" }} />
        </div>

        <div className="flex p-1 rounded-2xl mt-4 mb-2" style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}>
          {(["contacts", "groups"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className="flex-1 py-2.5 rounded-xl border-none cursor-pointer transition-all"
              style={{
                fontSize: 14,
                fontWeight: 600,
                color: activeTab === tab ? "white" : "var(--echo-text-tertiary)",
                background: activeTab === tab ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` : "transparent",
                boxShadow: activeTab === tab ? "var(--echo-shadow-md)" : "none",
              }}
            >
              {tab === "contacts" ? "Contacts" : "Groups"}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-4 pb-4">
        {activeTab === "contacts" ? (
          <div className="space-y-3 mt-2">
            <div className="flex items-center justify-between px-2">
              <span style={{ fontSize: 11, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>
                Active Contacts
              </span>
              <span className="rounded-full px-2 py-0.5" style={{ fontSize: 10, fontWeight: 700, color: "var(--echo-tier-inner-text)", backgroundColor: "var(--echo-tier-inner-bg)", border: "1px solid var(--echo-tier-inner-border)" }}>
                {onlineCount} ONLINE
              </span>
            </div>

            {filteredContacts.map((contact) => {
              const tier = tierStyles[contact.trustTier];
              return (
                <div
                  key={contact.id}
                  onClick={() => navigate(`/chat/${contact.id}`)}
                  className="flex items-center justify-between rounded-3xl p-4 cursor-pointer transition-all"
                  style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}
                >
                  <div className="flex items-center gap-4">
                    <div className="relative">
                      <div
                        className="rounded-2xl overflow-hidden"
                        style={{
                          width: 56, height: 56,
                          padding: contact.trustTier === "inner-circle" ? 2 : 0,
                          background: contact.trustTier === "inner-circle" ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` : "var(--echo-bg-elevated)",
                        }}
                      >
                        <ImageWithFallback src={contact.avatar} alt={contact.name} className="w-full h-full object-cover" style={{ borderRadius: contact.trustTier === "inner-circle" ? 14 : 16 }} />
                      </div>
                      <div className="absolute -bottom-1 -right-1 rounded-full" style={{ width: 16, height: 16, backgroundColor: contact.online ? "var(--echo-online)" : "var(--echo-offline)", border: "2px solid var(--echo-bg-surface)" }} />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-text-primary)" }}>{contact.name}</span>
                        {contact.isVerified && <ShieldCheck size={14} style={{ color: "var(--echo-sky)" }} />}
                      </div>
                      <div className="mt-1">
                        <span className="rounded px-1.5 py-0.5" style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: `var(${tier.textVar})`, backgroundColor: `var(${tier.bgVar})`, border: `1px solid var(${tier.borderVar})` }}>
                          {tier.label}
                        </span>
                      </div>
                    </div>
                  </div>
                  <ChevronRight size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                </div>
              );
            })}

            <div className="flex items-center justify-between px-2 pt-4">
              <span style={{ fontSize: 11, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>Groups</span>
            </div>
            {filteredGroups.map((g) => <GroupCard key={g.id} group={g} onClick={() => navigate(`/groups/${g.id}`)} />)}
          </div>
        ) : (
          <div className="space-y-3 mt-2">
            {filteredGroups.map((g) => <GroupCard key={g.id} group={g} onClick={() => navigate(`/groups/${g.id}`)} />)}
          </div>
        )}
      </div>
    </div>
  );
}
