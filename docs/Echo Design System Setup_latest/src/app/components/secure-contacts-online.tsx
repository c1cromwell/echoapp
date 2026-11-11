import React from "react";
import { useNavigate } from "react-router";
import { ImageWithFallback } from "./figma/ImageWithFallback";

interface OnlineContact {
  id: string;
  name: string;
  avatar: string;
  online: boolean;
  isGroup?: boolean;
}

const onlineContacts: OnlineContact[] = [
  { id: "1", name: "Sarah", avatar: "https://images.unsplash.com/photo-1655249493799-9cee4fe983bb?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcm9mZXNzaW9uYWwlMjB3b21hbiUyMHBvcnRyYWl0JTIwaGVhZHNob3R8ZW58MXx8fHwxNzc0ODgwNDc0fDA&ixlib=rb-4.1.0&q=80&w=1080", online: true },
  { id: "4", name: "HQ", avatar: "", online: true, isGroup: true },
  { id: "5", name: "Bob", avatar: "https://images.unsplash.com/photo-1769636930047-4478f12cf430?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcm9mZXNzaW9uYWwlMjBtYW4lMjBnbGFzc2VzJTIwcG9ydHJhaXR8ZW58MXx8fHwxNzc0ODk0NDQwfDA&ixlib=rb-4.1.0&q=80&w=1080", online: true },
  { id: "7", name: "Dev", avatar: "https://images.unsplash.com/photo-1769961982389-bb243681421a?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHx5b3VuZyUyMGFzaWFuJTIwbWFuJTIwcG9ydHJhaXQlMjBoZWFkc2hvdHxlbnwxfHx8fDE3NzQ5Nzk5Nzl8MA&ixlib=rb-4.1.0&q=80&w=1080", online: true },
  { id: "3", name: "Elena", avatar: "https://images.unsplash.com/photo-1581065178047-8ee15951ede6?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcm9mZXNzaW9uYWwlMjB3b21hbiUyMGJ1c2luZXNzJTIwcG9ydHJhaXR8ZW58MXx8fHwxNzc0OTMzMzI0fDA&ixlib=rb-4.1.0&q=80&w=1080", online: false },
];

export function SecureContactsOnline() {
  const navigate = useNavigate();

  return (
    <div className="px-4 pb-2">
      <h2 style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.2em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)", marginBottom: 12, marginLeft: 2 }}>
        Secure Contacts Online
      </h2>
      <div className="flex items-center gap-5 overflow-x-auto pb-2" style={{ scrollbarWidth: "none" }}>
        {onlineContacts.map((contact) => (
          <button
            key={contact.id}
            onClick={() => navigate(`/chat/${contact.id}`)}
            className="flex flex-col items-center gap-1.5 bg-transparent border-none cursor-pointer p-0 shrink-0"
            style={{ minWidth: 56 }}
          >
            <div className="relative">
              <div
                className="rounded-full overflow-hidden flex items-center justify-center"
                style={{
                  width: 56, height: 56, padding: 2,
                  background: contact.online
                    ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`
                    : "var(--echo-border)",
                }}
              >
                {contact.isGroup ? (
                  <div className="w-full h-full rounded-full flex items-center justify-center" style={{ backgroundColor: "var(--echo-bg-surface)" }}>
                    <span style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-tertiary)" }}>{contact.name.charAt(0)}</span>
                  </div>
                ) : (
                  <ImageWithFallback src={contact.avatar} alt={contact.name} className="w-full h-full rounded-full object-cover" style={{ border: "2px solid var(--echo-bg-elevated)" }} />
                )}
              </div>
              <div className="absolute rounded-full" style={{ bottom: 2, right: 2, width: 14, height: 14, backgroundColor: contact.online ? "var(--echo-online)" : "var(--echo-offline)", border: "2.5px solid var(--echo-bg-elevated)" }} />
            </div>
            <span style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-secondary)", maxWidth: 56, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" as const }}>
              {contact.name}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
