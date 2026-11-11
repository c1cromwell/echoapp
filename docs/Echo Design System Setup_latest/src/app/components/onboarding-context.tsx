import React, { createContext, useContext, useState, useCallback } from "react";
import type { VerificationResult, TrustScoreBreakdown } from "./credential-data";

export interface OnboardingState {
  phoneNumber: string;
  countryCode: string;
  otpCode: string;
  passkeyEnabled: boolean;
  passkeySkipped: boolean;
  recoveryMethod: "email" | "wallet" | "contact" | null;
  recoverySkipped: boolean;
  displayName: string;
  username: string;
  bio: string;
  avatarUrl: string;
  completed: boolean;
  // Credential verification
  verifiedCredentials: VerificationResult[];
  credentialSkipped: boolean;
  trustScoreBreakdown: TrustScoreBreakdown | null;
}

interface OnboardingContextType {
  state: OnboardingState;
  update: (partial: Partial<OnboardingState>) => void;
  skippedSteps: string[];
}

const defaultState: OnboardingState = {
  phoneNumber: "",
  countryCode: "+1",
  otpCode: "",
  passkeyEnabled: false,
  passkeySkipped: false,
  recoveryMethod: null,
  recoverySkipped: false,
  displayName: "",
  username: "",
  bio: "",
  avatarUrl: "",
  completed: false,
  verifiedCredentials: [],
  credentialSkipped: false,
  trustScoreBreakdown: null,
};

const OnboardingContext = createContext<OnboardingContextType>({
  state: defaultState,
  update: () => {},
  skippedSteps: [],
});

export function useOnboarding() {
  return useContext(OnboardingContext);
}

export function OnboardingProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<OnboardingState>(defaultState);

  const update = useCallback((partial: Partial<OnboardingState>) => {
    setState((prev) => ({ ...prev, ...partial }));
  }, []);

  const skippedSteps: string[] = [];
  if (state.passkeySkipped) skippedSteps.push("passkey");
  if (state.recoverySkipped) skippedSteps.push("recovery");
  if (state.credentialSkipped) skippedSteps.push("credentials");

  return (
    <OnboardingContext.Provider value={{ state, update, skippedSteps }}>
      {children}
    </OnboardingContext.Provider>
  );
}