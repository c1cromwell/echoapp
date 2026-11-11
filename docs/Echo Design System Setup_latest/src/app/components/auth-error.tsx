import React, { useEffect } from "react";
import { AlertCircle } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

interface AuthErrorProps {
  message: string;
  onDismiss: () => void;
  autoDismissMs?: number;
}

/**
 * Standard auth error display per Echo Auth UX Spec Section 14.
 * - Position: below the relevant input or action, 8px top spacing
 * - Text: 14px, Regular, error color
 * - Icon: 16px exclamationmark.circle inline before text
 * - Appearance: Fade in 200ms, slide down 4px
 * - Disappearance: On next user input or after autoDismissMs (default 5000ms)
 */
export function AuthError({ message, onDismiss, autoDismissMs = 5000 }: AuthErrorProps) {
  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => {
      onDismiss();
    }, autoDismissMs);
    return () => clearTimeout(timer);
  }, [message, onDismiss, autoDismissMs]);

  return (
    <AnimatePresence>
      {message && (
        <motion.div
          initial={{ opacity: 0, y: -4 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -4 }}
          transition={{ duration: 0.2 }}
          className="flex items-center gap-1.5 mt-2"
          role="alert"
          aria-live="polite"
        >
          <AlertCircle
            size={16}
            style={{ color: "var(--echo-error)", flexShrink: 0 }}
          />
          <span
            style={{
              fontSize: 14,
              color: "var(--echo-error)",
              lineHeight: "20px",
            }}
          >
            {message}
          </span>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
