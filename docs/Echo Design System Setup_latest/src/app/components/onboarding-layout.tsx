import React from "react";
import { Outlet } from "react-router";
import { OnboardingProvider } from "./onboarding-context";

export function OnboardingLayout() {
  return (
    <OnboardingProvider>
      <div
        className="flex justify-center min-h-screen"
        style={{ backgroundColor: "var(--echo-bg-secondary)" }}
      >
        <div
          className="flex flex-col w-full max-w-[430px] min-h-screen relative overflow-hidden"
          style={{ backgroundColor: "var(--echo-bg-primary)" }}
        >
          <Outlet />
        </div>
      </div>
    </OnboardingProvider>
  );
}
