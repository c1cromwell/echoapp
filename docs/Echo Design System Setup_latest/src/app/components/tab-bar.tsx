import React from "react";
import { useNavigate, useLocation } from "react-router";
import { MessageCircle, Users, Star, User } from "lucide-react";

const tabs = [
  { path: "/", label: "Messages", icon: MessageCircle },
  { path: "/contacts", label: "Contacts", icon: Users },
  { path: "/rewards", label: "Rewards", icon: Star },
  { path: "/profile", label: "Profile", icon: User },
];

export function TabBar() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <nav
      className="flex items-end justify-around px-2 shrink-0"
      style={{
        height: 82,
        paddingBottom: 20,
        backgroundColor: "var(--echo-bg-elevated)",
        borderTop: "1px solid var(--echo-border-light)",
      }}
    >
      {tabs.map((tab) => {
        const isActive =
          tab.path === "/"
            ? location.pathname === "/" || location.pathname.startsWith("/chat")
            : location.pathname.startsWith(tab.path);
        const Icon = tab.icon;

        return (
          <button
            key={tab.path}
            onClick={() => navigate(tab.path)}
            className="flex flex-col items-center gap-1 bg-transparent border-none cursor-pointer px-3 py-1 transition-colors"
            style={{ minWidth: 64 }}
          >
            <div className="relative">
              <Icon
                size={24}
                strokeWidth={isActive ? 2.2 : 1.8}
                style={{ color: isActive ? "var(--echo-accent)" : "var(--echo-text-tertiary)" }}
              />
              {/* Unread badge on Messages */}
              {tab.label === "Messages" && (
                <div
                  className="absolute -top-0.5 -right-1.5 rounded-full"
                  style={{
                    width: 8,
                    height: 8,
                    backgroundColor: "var(--echo-error)",
                  }}
                />
              )}
            </div>
            <span
              style={{
                fontSize: 10,
                fontWeight: isActive ? 600 : 500,
                color: isActive ? "var(--echo-accent)" : "var(--echo-text-tertiary)",
                lineHeight: "14px",
              }}
            >
              {tab.label}
            </span>
          </button>
        );
      })}
    </nav>
  );
}