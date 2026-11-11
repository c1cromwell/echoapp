import React, { useId } from "react";

interface TrustRingProps {
  score: number;
  size?: number;
  strokeWidth?: number;
  label?: string;
}

export function TrustRing({ score, size = 140, strokeWidth = 10, label }: TrustRingProps) {
  const gradientId = useId();
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const progress = (score / 100) * circumference;
  const offset = circumference - progress;

  return (
    <div className="flex flex-col items-center gap-2">
      <div className="relative" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="-rotate-90">
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="var(--echo-border)"
            strokeWidth={strokeWidth}
          />
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke={`url(#${gradientId})`}
            strokeWidth={strokeWidth}
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            className="transition-all duration-1000 ease-out"
          />
          <defs>
            <linearGradient id={gradientId} x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="var(--echo-gradient-from)" />
              <stop offset="100%" stopColor="var(--echo-gradient-to)" />
            </linearGradient>
          </defs>
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span
            style={{
              fontSize: 36,
              fontWeight: 700,
              color: "var(--echo-primary)",
              lineHeight: 1,
              letterSpacing: "-1px",
            }}
          >
            {score}
          </span>
          <span
            style={{
              fontSize: 14,
              fontWeight: 400,
              color: "var(--echo-text-tertiary)",
              lineHeight: 1,
              marginTop: 2,
            }}
          >
            /100
          </span>
        </div>
      </div>
      {label && (
        <span
          style={{
            fontSize: 14,
            fontWeight: 500,
            color: "var(--echo-text-secondary)",
          }}
        >
          {label}
        </span>
      )}
    </div>
  );
}