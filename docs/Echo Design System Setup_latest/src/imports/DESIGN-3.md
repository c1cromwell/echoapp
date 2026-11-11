# Design System Specification: The Glacial Interface

## 1. Overview & Creative North Star
**Creative North Star: "The Glacial Interface"**
This design system moves away from the "flat web" era into a realm of crystalline clarity and architectural stillness. It is designed to feel like a high-end physical object—specifically, a series of carved, frosted glass panes layered over a deep, subterranean digital architecture. 

To break the "template" look, we employ **Intentional Asymmetry**. In this system, we do not center-align for the sake of balance; we use weighted typography and significant negative space (the "ice fields") to guide the eye. We prioritize editorial-grade hierarchy where the scale between a headline and body text is dramatic, conveying an authoritative, secure, and premium experience for decentralized communication.

---

## 2. Colors & Tonal Soul
The palette is rooted in cold, high-trust hues. We avoid pure neutrals in favor of "Living Greys" tinted with cobalt and sky.

### Color Tokens (Material Translation)
- **Primary (The Core):** `#006591` — Used for high-action states and brand presence.
- **Primary Container (The Accent):** `#0EA5E9` — Our signature Sky Blue, reserved for primary CTAs and hero highlights.
- **Surface (The Base):** `#FAF8FF` (Light) / `#0F172A` (Dark) — The foundation of the app.
- **Surface Container Tiers:** From `lowest` (#FFFFFF) to `highest` (#DAE2FD).

### The "No-Line" Rule
Standard UI relies on 1px borders to separate content. This design system **prohibits** 1px solid borders for sectioning. Boundaries are defined exclusively through:
1.  **Tonal Shifts:** A `surface-container-low` card sitting on a `surface` background.
2.  **Soft Shadows:** Diffused light to imply edges.
3.  **Negative Space:** Using the `8` (2rem) and `10` (2.5rem) spacing tokens to create mental boundaries.

### Signature Textures
For hero elements and high-security states (e.g., "Encrypted Connection Established"), use a **Signature Gradient**:
*   *Direction:* 135deg (Top-Left to Bottom-Right)
*   *Colors:* From Deep Navy (`#0F172A`) to Sky Blue (`#0EA5E9`).
This gradient provides the "visual soul" that flat colors cannot achieve, mimicking the depth of a frozen lake.

---

## 3. Typography: The Editorial Voice
We use **Inter** not as a system font, but as a precision tool. Our hierarchy is intentionally steep to create a "High-End Editorial" feel.

- **Display-LG (3.5rem / 56px):** Used for "Hello" moments or empty states. Massive, bold, and authoritative.
- **Headline-SM (1.5rem / 24px):** The standard for screen titles.
- **Body-LG (1rem / 16px):** The primary chat bubble and message text size for maximum readability.
- **Label-MD (0.75rem / 12px):** Used for timestamps and metadata, set with slightly increased letter-spacing (+2%) for a premium "small print" feel.

**Director's Note:** Always pair `Display` sizes with `Body-MD`. The contrast in scale creates a sophisticated, "magazine-like" layout that feels custom-built rather than generated.

---

## 4. Elevation & Depth: Tonal Layering
In a decentralized app, "Trust" is conveyed through clarity. Depth in this system is achieved by "stacking" layers of light rather than casting dark shadows.

### The Layering Principle
Think of the UI as physical layers of frosted glass. 
- **Base Layer:** `surface`
- **Secondary Layer:** `surface-container-low` (for grouped content)
- **Action Layer:** `surface-container-lowest` (for cards that require interaction)

### Glassmorphism & Ghost Borders
Floating elements (Modals, Navigation Bars) must use a **Backdrop Blur (20px - 32px)**.
*   **Background:** `surface_variant` at 60% opacity.
*   **The Ghost Border:** If a boundary is required for accessibility, use the `outline-variant` token at **15% opacity**. Never 100%. This creates a "glint" on the edge of the glass rather than a hard line.

### Ambient Shadows
Shadows are never grey. They are tinted:
*   **Color:** `on-surface` (at 4%–8% opacity).
*   **Blur:** 24px to 48px for a "soft-glow" lift.

---

## 5. Components

### Message Bubbles (Cards & Lists)
*   **Inbound:** `surface-container-highest` with a `24px` rounded corner on all sides except the bottom-left (`8px`).
*   **Outbound:** Signature Gradient (Deep Navy to Sky Blue) with white text (`on-primary`).
*   **Spacing:** No dividers. Use `spacing.3` (0.75rem) between bubbles in a cluster, and `spacing.6` (1.5rem) between different speakers.

### Buttons
*   **Primary:** High-gloss. Use the Signature Gradient with a `full` (9999px) roundedness for a pill shape.
*   **Secondary (Glass):** Semi-transparent `surface_container_low` with a `20px` backdrop blur. No fill, just the "Ghost Border" at 20% opacity.

### Input Fields
*   **Style:** Minimalist. No bounding box.
*   **Structure:** A subtle `surface-container-low` background with `16px` (xl) corner radius.
*   **State:** On focus, the "Ghost Border" opacity increases to 40% using the `primary_container` color.

### Secure Thread Indicator (Custom)
A custom component for this system: A thin, 2px glowing line using `primary_container` that sits at the very top of the navigation bar, pulsating slightly (opacity 0.6 to 1.0) to indicate an active, decentralized, encrypted connection.

---

## 6. Do's and Don'ts

### Do
- **DO** use the `xl` (1.5rem) and `full` corner radii to maintain the "organic" feel of carved ice.
- **DO** use white space as a structural element. If the layout feels "crowded," remove a container, don't add a line.
- **DO** use `surface-bright` for hover states to simulate light hitting the surface of ice.

### Don't
- **DON'T** use pure black `#000000`. It breaks the "Icy" transparency of the system.
- **DON'T** use 1px solid borders at 100% opacity. This is the quickest way to make the design look "cheap" or "default."
- **DON'T** use standard iOS "Separator" lines in lists. Use tonal shifts between `surface` and `surface-container-low`.
- **DON'T** use harsh, fast animations. Use "Spring" physics with a high damping ratio to mimic the weight of premium glass.