import React, { useState, useRef, useEffect, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { ArrowLeft, Loader2 } from "lucide-react";
import { useOnboarding } from "../../components/onboarding-context";
import { motion } from "motion/react";
import { AuthError } from "../../components/auth-error";

type OtpContext = "registration" | "stepup" | "recovery";

const contextConfig: Record<OtpContext, { subtitle: string; showWrongNumber: boolean; nextPath: string }> = {
  registration: { subtitle: "Sent to", showWrongNumber: true, nextPath: "/onboarding/passkey" },
  stepup: { subtitle: "We sent a code to your phone", showWrongNumber: false, nextPath: "/" },
  recovery: { subtitle: "Verify your phone to recover", showWrongNumber: false, nextPath: "/onboarding/recover-complete" },
};

// The "correct" demo code
const DEMO_CODE = "123456";

export function OtpVerifyPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const context = (searchParams.get("context") as OtpContext) || "registration";
  const config = contextConfig[context];
  const { state } = useOnboarding();
  const [otp, setOtp] = useState<string[]>(["", "", "", "", "", ""]);
  const [timer, setTimer] = useState(45);
  const [verifying, setVerifying] = useState(false);
  const [error, setError] = useState("");
  const [shaking, setShaking] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const [verifyAttempts, setVerifyAttempts] = useState(0);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const phoneDisplay = state.phoneNumber
    ? `${state.countryCode} (${state.phoneNumber.slice(0, 3)}) ${state.phoneNumber.slice(3, 6)}-${state.phoneNumber.slice(6)}`
    : "+1 (555) 123-4567";

  useEffect(() => {
    if (timer > 0) {
      const t = setTimeout(() => setTimer(timer - 1), 1000);
      return () => clearTimeout(t);
    }
  }, [timer]);

  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

  const filled = otp.join("").length;
  const isComplete = filled === 6;

  const handleChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return;
    // Clear error on new input
    if (error) setError("");

    const digit = value.slice(-1);
    const newOtp = [...otp];
    newOtp[index] = digit;
    setOtp(newOtp);

    if (digit && index < 5) {
      inputRefs.current[index + 1]?.focus();
      setActiveIndex(index + 1);
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === "Backspace" && !otp[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
      setActiveIndex(index - 1);
      const newOtp = [...otp];
      newOtp[index - 1] = "";
      setOtp(newOtp);
      if (error) setError("");
    }
  };

  const handleVerify = useCallback(() => {
    if (!isComplete) return;
    const newAttempts = verifyAttempts + 1;
    setVerifyAttempts(newAttempts);

    // Too many attempts
    if (newAttempts > 3) {
      setError("Too many tries. Request a new code.");
      setShaking(true);
      setOtp(["", "", "", "", "", ""]);
      setTimeout(() => inputRefs.current[0]?.focus(), 500);
      setActiveIndex(0);
      return;
    }

    setVerifying(true);
    setError("");

    setTimeout(() => {
      setVerifying(false);
      const code = otp.join("");

      // Check if code expired (timer at 0 and no resend)
      if (timer <= 0) {
        setError("Code expired. Tap Resend to get a new one.");
        setShaking(true);
        setOtp(["", "", "", "", "", ""]);
        setTimeout(() => inputRefs.current[0]?.focus(), 500);
        setActiveIndex(0);
        return;
      }

      // Simulate wrong code (anything other than DEMO_CODE or all same digits)
      if (code !== DEMO_CODE && !/^(\d)\1{5}$/.test(code)) {
        setError("That code didn't work. Check and try again.");
        setShaking(true);
        setOtp(["", "", "", "", "", ""]);
        setTimeout(() => inputRefs.current[0]?.focus(), 500);
        setActiveIndex(0);
        return;
      }

      // Success
      navigate(config.nextPath);
    }, 1500);
  }, [isComplete, navigate, config.nextPath, otp, timer, verifyAttempts]);

  const handleResend = () => {
    setTimer(45);
    setOtp(["", "", "", "", "", ""]);
    setError("");
    setVerifyAttempts(0);
    inputRefs.current[0]?.focus();
    setActiveIndex(0);
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <div className="flex items-center px-4 pt-14 pb-2">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
      </div>

      <div className="flex-1 px-6">
        <h1
          style={{
            fontSize: 28,
            fontWeight: 700,
            color: "var(--echo-text-primary)",
            lineHeight: "36px",
            letterSpacing: "-0.5px",
            margin: "0 0 8px",
          }}
        >
          Enter verification code
        </h1>
        <p
          style={{
            fontSize: 16,
            color: "var(--echo-text-secondary)",
            lineHeight: "24px",
            margin: "0 0 40px",
          }}
        >
          {context === "registration" ? `${config.subtitle} ${phoneDisplay}` : config.subtitle}
        </p>

        {/* OTP Boxes with shake animation */}
        <motion.div
          animate={shaking ? { x: [0, -6, 6, -6, 6, 0] } : {}}
          transition={{ duration: 0.4 }}
          onAnimationComplete={() => setShaking(false)}
          className="flex justify-center gap-2 mb-2"
        >
          {otp.map((digit, i) => (
            <input
              key={i}
              ref={(el) => { inputRefs.current[i] = el; }}
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={digit}
              onChange={(e) => handleChange(i, e.target.value)}
              onKeyDown={(e) => handleKeyDown(i, e)}
              onFocus={() => setActiveIndex(i)}
              className="text-center rounded-xl outline-none transition-all"
              style={{
                width: 48,
                height: 56,
                fontSize: 24,
                fontWeight: 600,
                color: "var(--echo-text-primary)",
                backgroundColor: error
                  ? "var(--echo-error-light)"
                  : digit
                  ? "var(--echo-primary-subtle)"
                  : "var(--echo-bg-surface)",
                border: `2px solid ${
                  error
                    ? "var(--echo-error)"
                    : activeIndex === i
                    ? "var(--echo-border-focus)"
                    : "var(--echo-border)"
                }`,
              }}
              aria-label={`Digit ${i + 1}`}
            />
          ))}
        </motion.div>

        {/* Error message with auto-dismiss */}
        <div className="flex justify-center mb-4">
          <AuthError message={error} onDismiss={() => setError("")} autoDismissMs={5000} />
        </div>

        {/* Resend */}
        <div className="text-center mb-8">
          {timer > 0 ? (
            <span
              style={{
                fontSize: 14,
                color: "var(--echo-text-secondary)",
              }}
            >
              Resend code ({Math.floor(timer / 60)}:{(timer % 60).toString().padStart(2, "0")})
            </span>
          ) : (
            <button
              onClick={handleResend}
              className="bg-transparent border-none cursor-pointer"
              style={{
                fontSize: 14,
                fontWeight: 600,
                color: "var(--echo-primary)",
              }}
            >
              Resend code
            </button>
          )}
        </div>

        {/* Verify button */}
        <button
          onClick={handleVerify}
          disabled={!isComplete || verifying}
          className="w-full border-none rounded-xl flex items-center justify-center gap-2 transition-all active:scale-[0.98]"
          style={{
            height: 56,
            background: isComplete
              ? "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))"
              : "var(--echo-bg-surface)",
            color: isComplete ? "white" : "var(--echo-text-tertiary)",
            fontSize: 16,
            fontWeight: 600,
            cursor: isComplete ? "pointer" : "not-allowed",
            boxShadow: isComplete ? "var(--echo-shadow-primary)" : "none",
          }}
        >
          {verifying ? (
            <>
              <Loader2 size={20} className="animate-spin" />
              Verifying...
            </>
          ) : (
            "Verify"
          )}
        </button>

        {/* Demo hint */}
        <p
          className="text-center mt-4"
          style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}
        >
          Demo: enter 123456 or any repeating digit (e.g. 111111)
        </p>
      </div>

      {/* Wrong number — only in registration context */}
      {config.showWrongNumber && (
        <div className="text-center px-6 pb-10">
          <span style={{ fontSize: 14, color: "var(--echo-text-tertiary)" }}>
            Wrong number?{" "}
          </span>
          <button
            onClick={() => navigate(-1)}
            className="bg-transparent border-none cursor-pointer p-0"
            style={{
              fontSize: 14,
              fontWeight: 600,
              color: "var(--echo-primary)",
            }}
          >
            Change
          </button>
        </div>
      )}
    </div>
  );
}
