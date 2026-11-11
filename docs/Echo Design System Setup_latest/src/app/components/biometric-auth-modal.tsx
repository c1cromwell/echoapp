import React, { useEffect, useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { ScanFace, Fingerprint, KeyRound, X } from "lucide-react";

interface BiometricAuthModalProps {
  open: boolean;
  onClose: () => void;
  onAuthenticate: () => Promise<boolean>;
  onDuressPin?: (pin: string) => boolean;
  title?: string;
  subtitle?: string;
}

export function BiometricAuthModal({
  open,
  onClose,
  onAuthenticate,
  onDuressPin,
  title = "Hidden Folders",
  subtitle = "Authenticate to access",
}: BiometricAuthModalProps) {
  const [phase, setPhase] = useState<"idle" | "scanning" | "success" | "failed" | "pin">("idle");
  const [pin, setPin] = useState("");
  const [pinError, setPinError] = useState(false);

  useEffect(() => {
    if (open) {
      setPhase("idle");
      setPin("");
      setPinError(false);
      // Auto-start scan after brief delay
      const timer = setTimeout(() => {
        setPhase("scanning");
        startAuth();
      }, 400);
      return () => clearTimeout(timer);
    }
  }, [open]);

  const startAuth = async () => {
    setPhase("scanning");
    const success = await onAuthenticate();
    if (success) {
      setPhase("success");
      setTimeout(onClose, 600);
    } else {
      setPhase("failed");
    }
  };

  const handlePinSubmit = (pinValue?: string) => {
    const currentPin = pinValue || pin;
    if (onDuressPin && currentPin.length >= 4) {
      const success = onDuressPin(currentPin);
      if (success) {
        setPhase("success");
        setTimeout(onClose, 600);
      } else {
        setPinError(true);
        setPin("");
        setTimeout(() => setPinError(false), 1500);
      }
    }
  };

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100]"
            style={{ backgroundColor: "rgba(0,0,0,0.7)", backdropFilter: "blur(8px)" }}
            onClick={onClose}
          />
          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 20 }}
            transition={{ type: "spring", damping: 25, stiffness: 300 }}
            className="fixed inset-0 z-[101] flex items-center justify-center px-10"
          >
            <div
              className="w-full max-w-[320px] rounded-3xl overflow-hidden"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "0 25px 50px -12px rgba(0,0,0,0.4)",
              }}
            >
              {/* Close */}
              <div className="flex justify-end p-3 pb-0">
                <button
                  onClick={onClose}
                  className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1 rounded-full hover:bg-[var(--echo-bg-surface)]"
                >
                  <X size={18} style={{ color: "var(--echo-text-tertiary)" }} />
                </button>
              </div>

              {phase !== "pin" ? (
                /* Biometric View */
                <div className="flex flex-col items-center px-8 pb-8">
                  {/* Icon */}
                  <motion.div
                    className="flex items-center justify-center rounded-full mb-5"
                    style={{
                      width: 88,
                      height: 88,
                      background:
                        phase === "success"
                          ? "linear-gradient(135deg, var(--echo-success), #34D399)"
                          : phase === "failed"
                          ? "linear-gradient(135deg, var(--echo-error), #F87171)"
                          : "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                    }}
                    animate={
                      phase === "scanning"
                        ? { scale: [1, 1.08, 1], opacity: [1, 0.8, 1] }
                        : phase === "success"
                        ? { scale: [1, 1.15, 1] }
                        : {}
                    }
                    transition={
                      phase === "scanning"
                        ? { duration: 1.2, repeat: Infinity, ease: "easeInOut" }
                        : { duration: 0.4 }
                    }
                  >
                    <ScanFace size={40} color="white" strokeWidth={1.5} />
                  </motion.div>

                  {/* Title */}
                  <div
                    style={{
                      fontSize: 18,
                      fontWeight: 700,
                      color: "var(--echo-text-primary)",
                      lineHeight: "24px",
                      textAlign: "center",
                      marginBottom: 6,
                    }}
                  >
                    {title}
                  </div>
                  <div
                    style={{
                      fontSize: 14,
                      color:
                        phase === "success"
                          ? "var(--echo-success)"
                          : phase === "failed"
                          ? "var(--echo-error)"
                          : "var(--echo-text-secondary)",
                      lineHeight: "20px",
                      textAlign: "center",
                      marginBottom: 24,
                    }}
                  >
                    {phase === "idle" && subtitle}
                    {phase === "scanning" && "Scanning..."}
                    {phase === "success" && "Authenticated"}
                    {phase === "failed" && "Authentication failed"}
                  </div>

                  {/* Scanning indicator */}
                  {phase === "scanning" && (
                    <div className="flex items-center gap-1.5 mb-4">
                      {[0, 1, 2].map((i) => (
                        <motion.div
                          key={i}
                          className="rounded-full"
                          style={{
                            width: 6,
                            height: 6,
                            backgroundColor: "var(--echo-primary)",
                          }}
                          animate={{ opacity: [0.3, 1, 0.3] }}
                          transition={{
                            duration: 1,
                            repeat: Infinity,
                            delay: i * 0.2,
                          }}
                        />
                      ))}
                    </div>
                  )}

                  {/* Retry / PIN buttons */}
                  {phase === "failed" && (
                    <div className="flex flex-col gap-2 w-full">
                      <button
                        onClick={() => {
                          setPhase("scanning");
                          startAuth();
                        }}
                        className="w-full flex items-center justify-center gap-2 rounded-xl border-none cursor-pointer py-3"
                        style={{
                          background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                        }}
                      >
                        <Fingerprint size={18} color="white" />
                        <span style={{ fontSize: 14, fontWeight: 600, color: "white" }}>
                          Try Again
                        </span>
                      </button>
                      {onDuressPin && (
                        <button
                          onClick={() => setPhase("pin")}
                          className="w-full flex items-center justify-center gap-2 rounded-xl border-none cursor-pointer py-3"
                          style={{
                            backgroundColor: "var(--echo-bg-surface)",
                          }}
                        >
                          <KeyRound size={16} style={{ color: "var(--echo-text-secondary)" }} />
                          <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
                            Use PIN
                          </span>
                        </button>
                      )}
                    </div>
                  )}
                </div>
              ) : (
                /* PIN Entry View */
                <div className="flex flex-col items-center px-8 pb-8">
                  <div
                    className="flex items-center justify-center rounded-full mb-5"
                    style={{
                      width: 64,
                      height: 64,
                      backgroundColor: "var(--echo-primary-subtle)",
                    }}
                  >
                    <KeyRound size={28} style={{ color: "var(--echo-primary)" }} />
                  </div>

                  <div
                    style={{
                      fontSize: 18,
                      fontWeight: 700,
                      color: "var(--echo-text-primary)",
                      textAlign: "center",
                      marginBottom: 6,
                    }}
                  >
                    Enter PIN
                  </div>
                  <div
                    style={{
                      fontSize: 13,
                      color: pinError ? "var(--echo-error)" : "var(--echo-text-tertiary)",
                      textAlign: "center",
                      marginBottom: 20,
                    }}
                  >
                    {pinError ? "Incorrect PIN" : "Enter your access PIN"}
                  </div>

                  {/* PIN dots */}
                  <div className="flex gap-3 mb-6">
                    {[0, 1, 2, 3].map((i) => (
                      <motion.div
                        key={i}
                        className="rounded-full"
                        style={{
                          width: 14,
                          height: 14,
                          backgroundColor:
                            pin.length > i
                              ? "var(--echo-primary)"
                              : "var(--echo-border)",
                          border: `2px solid ${
                            pin.length > i
                              ? "var(--echo-primary)"
                              : "var(--echo-border)"
                          }`,
                        }}
                        animate={pinError && pin.length === 0 ? { x: [0, -4, 4, -4, 0] } : {}}
                        transition={{ duration: 0.3 }}
                      />
                    ))}
                  </div>

                  {/* Number pad */}
                  <div className="grid grid-cols-3 gap-2 w-full max-w-[220px]">
                    {[1, 2, 3, 4, 5, 6, 7, 8, 9, null, 0, "del"].map((key, i) => (
                      <button
                        key={i}
                        onClick={() => {
                          if (key === null) return;
                          if (key === "del") {
                            setPin((prev) => prev.slice(0, -1));
                          } else {
                            const newPin = pin + key;
                            setPin(newPin);
                            if (newPin.length === 4) {
                              setTimeout(() => handlePinSubmit(newPin), 200);
                            }
                          }
                        }}
                        className="flex items-center justify-center rounded-xl border-none cursor-pointer"
                        style={{
                          height: 48,
                          backgroundColor: key === null ? "transparent" : "var(--echo-bg-surface)",
                          fontSize: key === "del" ? 12 : 20,
                          fontWeight: 500,
                          color: "var(--echo-text-primary)",
                          visibility: key === null ? "hidden" : "visible",
                        }}
                        disabled={key === null}
                      >
                        {key === "del" ? (
                          <X size={18} style={{ color: "var(--echo-text-tertiary)" }} />
                        ) : (
                          key
                        )}
                      </button>
                    ))}
                  </div>

                  {/* Back to biometric */}
                  <button
                    onClick={() => {
                      setPhase("idle");
                      setPin("");
                      setTimeout(() => {
                        setPhase("scanning");
                        startAuth();
                      }, 300);
                    }}
                    className="mt-4 flex items-center gap-2 bg-transparent border-none cursor-pointer"
                  >
                    <ScanFace size={16} style={{ color: "var(--echo-primary)" }} />
                    <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-primary)" }}>
                      Use Face ID
                    </span>
                  </button>
                </div>
              )}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}