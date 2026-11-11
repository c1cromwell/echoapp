import React, { useState } from "react";
import { useNavigate } from "react-router";
import { ArrowLeft, Loader2 } from "lucide-react";
import { motion } from "motion/react";
import { AuthError } from "../../components/auth-error";

const WORD_COUNT = 12;

export function RecoverPhrasePage() {
  const navigate = useNavigate();
  const [words, setWords] = useState<string[]>(Array(WORD_COUNT).fill(""));
  const [verifying, setVerifying] = useState(false);
  const [error, setError] = useState("");

  const filledCount = words.filter((w) => w.trim().length > 0).length;
  const isComplete = filledCount === WORD_COUNT;

  const handleWordChange = (index: number, value: string) => {
    if (error) setError("");
    const cleaned = value.toLowerCase().replace(/[^a-z]/g, "");
    const newWords = [...words];
    newWords[index] = cleaned;
    setWords(newWords);
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    const text = e.clipboardData.getData("text");
    const pastedWords = text
      .trim()
      .split(/[\s,]+/)
      .filter(Boolean)
      .slice(0, WORD_COUNT);
    if (pastedWords.length >= 2) {
      e.preventDefault();
      if (error) setError("");
      const newWords = Array(WORD_COUNT).fill("");
      pastedWords.forEach((w, i) => {
        if (i < WORD_COUNT) newWords[i] = w.toLowerCase().replace(/[^a-z]/g, "");
      });
      setWords(newWords);
    }
  };

  const handleVerify = () => {
    if (!isComplete) return;
    setVerifying(true);
    setError("");

    setTimeout(() => {
      setVerifying(false);
      // Simulate: accept any 12 filled words for demo
      navigate("/onboarding/recover-complete");
    }, 2000);
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

      <div className="flex-1 px-6 overflow-y-auto">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
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
            Recovery phrase
          </h1>
          <p
            style={{
              fontSize: 16,
              color: "var(--echo-text-secondary)",
              lineHeight: "24px",
              margin: "0 0 24px",
            }}
          >
            Enter your 12-word backup phrase in order. You can also paste all 12 words at once.
          </p>
        </motion.div>

        {/* Word grid */}
        <div
          className="grid grid-cols-2 gap-2 mb-4"
          onPaste={handlePaste}
        >
          {words.map((word, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.03, duration: 0.2 }}
              className="flex items-center gap-2 rounded-xl px-3"
              style={{
                height: 44,
                backgroundColor: word
                  ? "var(--echo-primary-subtle)"
                  : "var(--echo-bg-surface)",
                border: `1.5px solid ${
                  error
                    ? "var(--echo-error)"
                    : word
                    ? "var(--echo-primary)"
                    : "var(--echo-border)"
                }`,
              }}
            >
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: "var(--echo-text-tertiary)",
                  width: 20,
                  textAlign: "right",
                  flexShrink: 0,
                }}
              >
                {i + 1}.
              </span>
              <input
                type="text"
                value={word}
                onChange={(e) => handleWordChange(i, e.target.value)}
                className="flex-1 bg-transparent border-none outline-none p-0"
                style={{
                  fontSize: 14,
                  fontWeight: 500,
                  color: "var(--echo-text-primary)",
                  height: "100%",
                }}
                placeholder="word"
                autoComplete="off"
                autoCapitalize="off"
                spellCheck={false}
              />
            </motion.div>
          ))}
        </div>

        {/* Error display */}
        <div className="mb-4">
          <AuthError message={error} onDismiss={() => setError("")} />
        </div>

        {/* Progress indicator */}
        <div className="flex items-center justify-between mb-6">
          <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
            {filledCount} of {WORD_COUNT} words entered
          </span>
          <div
            className="rounded-full overflow-hidden"
            style={{ width: 80, height: 4, backgroundColor: "var(--echo-bg-surface)" }}
          >
            <div
              className="h-full rounded-full transition-all"
              style={{
                width: `${(filledCount / WORD_COUNT) * 100}%`,
                backgroundColor: "var(--echo-primary)",
              }}
            />
          </div>
        </div>

        {/* Verify button */}
        <button
          onClick={handleVerify}
          disabled={!isComplete || verifying}
          className="w-full border-none rounded-xl flex items-center justify-center gap-2 transition-all active:scale-[0.98] mb-8"
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
              Verifying phrase...
            </>
          ) : (
            "Recover Account"
          )}
        </button>
      </div>
    </div>
  );
}
