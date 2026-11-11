import React, { useState } from "react";
import { useNavigate } from "react-router";
import {
  ArrowLeft, Plus, Shield, ChevronRight, Trash2,
  ShieldCheck, Key, Fingerprint,
} from "lucide-react";
import { EchoAvatar } from "../components/echo-avatar";
import {
  mockPersonas,
  mockMasterIdentity,
  categoryInfo,
  trustLevelLimits,
  type Persona,
} from "../components/persona-data";
import { motion, AnimatePresence } from "motion/react";

export function PersonasPage() {
  const navigate = useNavigate();
  const [personas, setPersonas] = useState(mockPersonas);
  const [showDeleteId, setShowDeleteId] = useState<string | null>(null);
  const master = mockMasterIdentity;
  const limits = trustLevelLimits[master.trustLevel];

  const handleSetActive = (personaId: string) => {
    setPersonas(
      personas.map((p) => ({ ...p, isActive: p.personaId === personaId }))
    );
  };

  const handleDelete = (personaId: string) => {
    setPersonas(personas.filter((p) => p.personaId !== personaId));
    setShowDeleteId(null);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="flex items-center gap-3 px-3 shrink-0"
        style={{
          height: 96,
          paddingTop: 44,
          backgroundColor: "var(--echo-bg-elevated)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          style={{ width: 44, height: 44 }}
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <h1
          style={{
            fontSize: 18,
            fontWeight: 600,
            color: "var(--echo-text-primary)",
            margin: 0,
            flex: 1,
          }}
        >
          My Personas
        </h1>
        {personas.length < limits.maxPersonas && (
          <button
            onClick={() => navigate("/personas/create")}
            className="flex items-center justify-center rounded-full border-none cursor-pointer"
            style={{
              width: 36,
              height: 36,
              background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            }}
          >
            <Plus size={18} color="white" strokeWidth={2.5} />
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        {/* Master Identity Card */}
        <div className="px-4 pt-4 pb-2">
          <div
            className="rounded-2xl p-4"
            style={{
              background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            }}
          >
            <div className="flex items-center gap-3 mb-3">
              <div
                className="flex items-center justify-center rounded-xl"
                style={{ width: 44, height: 44, backgroundColor: "rgba(255,255,255,0.2)" }}
              >
                <Fingerprint size={24} color="white" />
              </div>
              <div>
                <div style={{ fontSize: 16, fontWeight: 600, color: "white" }}>
                  Master Identity
                </div>
                <div style={{ fontSize: 12, color: "rgba(255,255,255,0.7)" }}>
                  {master.did.slice(0, 24)}...
                </div>
              </div>
            </div>

            <div className="flex gap-3">
              {[
                { icon: ShieldCheck, label: "Trust", value: master.trustLevel },
                { icon: Key, label: "Score", value: `${master.trustScore}` },
                { icon: Shield, label: "Personas", value: `${personas.length}/${limits.maxPersonas}` },
              ].map((item) => (
                <div
                  key={item.label}
                  className="flex-1 rounded-xl p-2.5"
                  style={{ backgroundColor: "rgba(255,255,255,0.15)" }}
                >
                  <item.icon size={16} color="rgba(255,255,255,0.8)" />
                  <div style={{ fontSize: 14, fontWeight: 700, color: "white", marginTop: 4 }}>
                    {item.value}
                  </div>
                  <div style={{ fontSize: 10, color: "rgba(255,255,255,0.6)" }}>
                    {item.label}
                  </div>
                </div>
              ))}
            </div>

            {/* Verification badges */}
            <div className="flex gap-2 mt-3">
              {[
                { label: "Phone", verified: master.phoneVerified },
                { label: "Email", verified: master.emailVerified },
                { label: "KYC", verified: master.kycVerified },
              ].map((v) => (
                <span
                  key={v.label}
                  className="rounded-full px-2.5 py-1 flex items-center gap-1"
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: v.verified ? "white" : "rgba(255,255,255,0.5)",
                    backgroundColor: v.verified ? "rgba(255,255,255,0.25)" : "rgba(255,255,255,0.08)",
                  }}
                >
                  {v.verified ? "✓" : "○"} {v.label}
                </span>
              ))}
            </div>
          </div>
        </div>

        {/* Persona slots header */}
        <div className="flex items-center justify-between px-6 py-3">
          <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
            Your Personas ({personas.length})
          </span>
          {!limits.customCategories && (
            <span
              className="rounded-full px-2 py-0.5"
              style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-warning)", backgroundColor: "var(--echo-warning-light)" }}
            >
              Custom categories locked
            </span>
          )}
        </div>

        {/* Persona List */}
        <div className="px-4 flex flex-col gap-3 pb-6">
          {personas.map((persona, index) => {
            const cat = categoryInfo[persona.category];
            return (
              <motion.div
                key={persona.personaId}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.05 }}
              >
                <div
                  className="rounded-2xl overflow-hidden"
                  style={{
                    backgroundColor: "var(--echo-bg-elevated)",
                    border: persona.isActive ? "2px solid var(--echo-primary)" : "2px solid var(--echo-border)",
                    boxShadow: persona.isActive ? "var(--echo-shadow-primary)" : "var(--echo-shadow-sm)",
                  }}
                >
                  <button
                    onClick={() => navigate(`/personas/${persona.personaId}`)}
                    className="flex items-center gap-3 w-full px-4 py-3.5 bg-transparent border-none cursor-pointer text-left"
                  >
                    <div className="relative">
                      <EchoAvatar name={persona.displayName} size="lg" gradient={persona.gradient} />
                      {persona.isActive && (
                        <div
                          className="absolute -top-1 -right-1 flex items-center justify-center rounded-full"
                          style={{
                            width: 20,
                            height: 20,
                            backgroundColor: "var(--echo-primary)",
                            border: "2px solid var(--echo-bg-elevated)",
                          }}
                        >
                          <ShieldCheck size={10} color="white" />
                        </div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5">
                        <span
                          className="truncate"
                          style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}
                        >
                          {persona.displayName}
                        </span>
                        <span style={{ fontSize: 14 }}>{cat.emoji}</span>
                        {persona.isDefault && (
                          <span
                            className="rounded-full px-1.5"
                            style={{ fontSize: 9, fontWeight: 700, color: "var(--echo-primary)", backgroundColor: "var(--echo-primary-subtle)", lineHeight: "14px" }}
                          >
                            DEFAULT
                          </span>
                        )}
                      </div>
                      <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                        @{persona.username} · {cat.label}
                      </div>
                      {persona.bio && (
                        <div
                          className="truncate mt-1"
                          style={{ fontSize: 12, color: "var(--echo-text-tertiary)", maxWidth: 220 }}
                        >
                          {persona.bio}
                        </div>
                      )}
                      <div className="flex items-center gap-3 mt-2">
                        <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                          {persona.grants.length} access grants
                        </span>
                        <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                          {persona.privacy.searchable ? "Searchable" : "Hidden"}
                        </span>
                      </div>
                    </div>
                    <ChevronRight size={18} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>

                  {/* Actions row */}
                  <div
                    className="flex items-center px-4 py-2 gap-2"
                    style={{
                      borderTop: "1px solid var(--echo-border-light)",
                      backgroundColor: "var(--echo-bg-secondary)",
                    }}
                  >
                    <button
                      onClick={() => handleSetActive(persona.personaId)}
                      className="flex-1 rounded-lg border-none cursor-pointer py-1.5 transition-colors"
                      style={{
                        fontSize: 12,
                        fontWeight: 600,
                        color: persona.isActive ? "var(--echo-success)" : "var(--echo-primary)",
                        backgroundColor: persona.isActive ? "var(--echo-success-light)" : "var(--echo-primary-subtle)",
                      }}
                    >
                      {persona.isActive ? "Active" : "Switch To"}
                    </button>
                    {!persona.isDefault && (
                      <button
                        onClick={() => setShowDeleteId(persona.personaId)}
                        className="flex items-center justify-center rounded-lg border-none cursor-pointer p-1.5"
                        style={{ backgroundColor: "var(--echo-error-light)" }}
                      >
                        <Trash2 size={14} style={{ color: "var(--echo-error)" }} />
                      </button>
                    )}
                  </div>
                </div>
              </motion.div>
            );
          })}

          {/* Add persona slot */}
          {personas.length < limits.maxPersonas && (
            <button
              onClick={() => navigate("/personas/create")}
              className="flex items-center justify-center gap-2 rounded-2xl border-2 border-dashed cursor-pointer bg-transparent py-6 transition-colors hover:bg-[var(--echo-bg-surface)]"
              style={{ borderColor: "var(--echo-border)" }}
            >
              <Plus size={20} style={{ color: "var(--echo-primary)" }} />
              <span style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-primary)" }}>
                Create New Persona
              </span>
              <span
                className="rounded-full px-2 py-0.5 ml-1"
                style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-tertiary)", backgroundColor: "var(--echo-bg-surface)" }}
              >
                {limits.maxPersonas - personas.length} slots left
              </span>
            </button>
          )}
        </div>
      </div>

      {/* Delete confirmation overlay */}
      <AnimatePresence>
        {showDeleteId && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-end justify-center"
            style={{ backgroundColor: "rgba(0,0,0,0.4)" }}
            onClick={() => setShowDeleteId(null)}
          >
            <motion.div
              initial={{ y: 100 }}
              animate={{ y: 0 }}
              exit={{ y: 100 }}
              onClick={(e) => e.stopPropagation()}
              className="w-full max-w-[430px] rounded-t-3xl p-6"
              style={{ backgroundColor: "var(--echo-bg-elevated)" }}
            >
              <div className="flex flex-col items-center text-center mb-6">
                <div
                  className="flex items-center justify-center rounded-2xl mb-3"
                  style={{ width: 56, height: 56, backgroundColor: "var(--echo-error-light)" }}
                >
                  <Trash2 size={28} style={{ color: "var(--echo-error)" }} />
                </div>
                <h3 style={{ fontSize: 18, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px" }}>
                  Delete Persona?
                </h3>
                <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: 0, lineHeight: "20px" }}>
                  This will permanently remove this persona, its messages, and all access grants. This action cannot be undone.
                </p>
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowDeleteId(null)}
                  className="flex-1 rounded-xl border-none cursor-pointer py-3.5"
                  style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)", backgroundColor: "var(--echo-bg-surface)" }}
                >
                  Cancel
                </button>
                <button
                  onClick={() => handleDelete(showDeleteId)}
                  className="flex-1 rounded-xl border-none cursor-pointer py-3.5"
                  style={{ fontSize: 15, fontWeight: 600, color: "white", backgroundColor: "var(--echo-error)" }}
                >
                  Delete
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
