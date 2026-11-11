import React, { useEffect } from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import { TabBar } from "./tab-bar";

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();

  // Redirect to onboarding if not completed
  useEffect(() => {
    const isOnboarded = localStorage.getItem("echo-onboarded") === "true";
    if (!isOnboarded) {
      navigate("/onboarding", { replace: true });
    }
  }, [navigate]);

  const hideTabBar = location.pathname.startsWith("/chat") ||
    location.pathname.startsWith("/groups/") ||
    location.pathname.startsWith("/group-chat") ||
    location.pathname.startsWith("/personas") ||
    location.pathname.startsWith("/analytics") ||
    location.pathname.startsWith("/hidden-folders") ||
    location.pathname.startsWith("/channels") ||
    location.pathname.startsWith("/devices") ||
    location.pathname.startsWith("/login-audit") ||
    location.pathname.startsWith("/settings") ||
    location.pathname.startsWith("/trust");

  return (
    <div className="flex justify-center min-h-screen" style={{ backgroundColor: "var(--echo-bg-secondary)" }}>
      <div
        className="flex flex-col w-full max-w-[430px] min-h-screen relative"
        style={{ backgroundColor: "var(--echo-bg-primary)" }}
      >
        <div className="flex-1 overflow-hidden flex flex-col">
          <Outlet />
        </div>
        {!hideTabBar && <TabBar />}
      </div>
    </div>
  );
}