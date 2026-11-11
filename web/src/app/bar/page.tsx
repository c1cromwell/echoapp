"use client";

import { useState } from "react";
import { interpretBarQuery } from "@/lib/bar/query";

type Person = { did: string; name: string; allowed: boolean };
type Line = { id: string; mine: boolean; text: string };

export default function EchoBarPage() {
  const [people, setPeople] = useState<Person[]>([]);
  const [selected, setSelected] = useState<string | null>(null);
  const [lines, setLines] = useState<Record<string, Line[]>>({});
  const [draft, setDraft] = useState("");
  const [notice, setNotice] = useState(
    "Sign-in uses the device passkey. This page does not create an email account."
  );

  const person = people.find((item) => item.did === selected && item.allowed);
  const allowed = people.filter((item) => item.allowed);
  const candidates = people.filter((item) => !item.allowed);

  function submit() {
    if (person) {
      const text = draft.trim();
      if (!text) return;
      setLines((current) => ({
        ...current,
        [person.did]: [
          ...(current[person.did] ?? []),
          { id: crypto.randomUUID(), mine: true, text },
        ],
      }));
      setDraft("");
      setNotice("Saved in this window. Delivery uses the same sealed path once this browser can sign.");
      return;
    }
    const query = interpretBarQuery(draft);
    if (query.kind === "empty") return;
    if (query.kind === "emailNotLive") {
      setNotice("Email lookup is not live yet. Use an @username.");
      setDraft("");
      return;
    }
    setDraft("");
    setPeople((current) => {
      if (current.some((item) => item.name === `@${query.handle}`)) return current;
      return [
        ...current,
        { did: `pending:${query.handle}`, name: `@${query.handle}`, allowed: false },
      ];
    });
    setNotice(`@${query.handle} is a candidate. Allow is a separate step, and it needs the device passkey.`);
  }

  function allow(item: Person) {
    setPeople((current) =>
      current.map((row) => (row.did === item.did ? { ...row, allowed: true } : row))
    );
    setNotice("Allowed in this window. The phone’s passkey still has to confirm it on the server.");
  }

  return (
    <main
      style={{
        minHeight: "100vh",
        background: "#FBFBF7",
        color: "#0B0F10",
        display: "flex",
        justifyContent: "center",
      }}
    >
      <section
        style={{
          width: "100%",
          maxWidth: 420,
          minHeight: "100vh",
          display: "flex",
          flexDirection: "column",
          padding: 16,
          gap: 12,
        }}
      >
        <header>
          <h1 style={{ fontSize: 20, margin: 0 }}>{person?.name ?? "Echo"}</h1>
          <p style={{ margin: "4px 0 0", color: "#0B0F108C", fontSize: 13 }}>
            {person ? "Allowed" : "People who can reach you here"}
          </p>
          {person ? (
            <button type="button" onClick={() => setSelected(null)} style={{ marginTop: 8 }}>
              Switch
            </button>
          ) : null}
          {notice ? <p style={{ fontSize: 13 }}>{notice}</p> : null}
        </header>
        <div style={{ flex: 1 }}>
          {person
            ? (lines[person.did] ?? []).map((line) => (
                <p key={line.id} style={{ textAlign: line.mine ? "right" : "left" }}>
                  {line.text}
                </p>
              ))
            : null}
          {!person
            ? allowed.map((item) => (
                <button key={item.did} type="button" onClick={() => setSelected(item.did)}>
                  {item.name} · Allowed
                </button>
              ))
            : null}
          {!person
            ? candidates.map((item) => (
                <p key={item.did}>
                  {item.name}{" "}
                  <button type="button" onClick={() => allow(item)}>
                    Allow
                  </button>
                </p>
              ))
            : null}
        </div>
        <form
          onSubmit={(event) => {
            event.preventDefault();
            submit();
          }}
          style={{ display: "flex", gap: 8 }}
        >
          <input
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            placeholder={person ? `Message ${person.name}` : "Name, @username, or email"}
            aria-label={person ? `Message ${person.name}` : "Name, username, or email"}
            style={{ flex: 1, borderRadius: 999, border: "1px solid #0B0F101A", padding: "10px 14px" }}
          />
          <button type="submit">Send</button>
        </form>
      </section>
    </main>
  );
}
