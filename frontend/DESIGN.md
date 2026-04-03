# Design System Specification: Editorial Luminance

## 1. Overview & Creative North Star

**Creative North Star: The Lucid Curator**
This design system moves beyond the generic "SaaS White" aesthetic to embrace the spirit of a high-end digital broadsheet. It is characterized by expansive white space, razor-sharp typography, and a "Lucid" layering logic. We reject the standard "boxed-in" web experience in favor of an editorial layout where content breathes and hierarchy is established through tonal shifts rather than rigid lines.

By leveraging intentional asymmetry—such as offset headers and staggered grid placements—we break the "template" feel. The system utilizes a sophisticated light-theory approach where the signature neon purple acts as a rhythmic pulse across a pristine, high-contrast landscape.

---

## 2. Colors & Tonal Logic

The palette is rooted in a "Hyper-Light" spectrum. We use a base of `#fff7ff` (Surface) to provide a warm, sophisticated ivory undertone that prevents the clinical "blue-light" fatigue of pure white.

### The "No-Line" Rule

**Explicit Instruction:** Designers are prohibited from using 1px solid borders to define major sections. Structural boundaries must be created through background color shifts.

- _Implementation:_ A hero section on `surface` transitions into a content feed on `surface-container-low`.
- _The Result:_ A seamless, high-end feel that mimics the flow of premium paper stocks.

### Surface Hierarchy & Nesting

Depth is achieved through the physical stacking of tones.

- **Level 0 (Base):** `surface` (#fff7ff)
- **Level 1 (Sections):** `surface-container-low` (#faf0ff)
- **Level 2 (Cards/Modules):** `surface-container` (#f5eafa)
- **Level 3 (Popovers/Modals):** `surface-container-highest` (#e9dfef)

### The "Glass & Gradient" Rule

To add soul to the "neon purple" accent, avoid flat applications on large surfaces.

- **Signature Gradient:** Use a linear transition from `primary` (#6100ba) to `primary_container` (#7f13ec) at 135 degrees for primary CTAs.
- **Glassmorphism:** Floating navigation or top-level headers should use `surface` at 80% opacity with a `24px` backdrop-blur. This allows the high-contrast content to "ghost" through the UI as the user scrolls.

---

## 3. Typography: Spline Sans

We use **Spline Sans** exclusively. It is a typeface that balances geometric precision with a professional, "sharp" editorial edge.

- **Display (lg/md/sm):** Used for "Statement Content." Set with tight letter-spacing (-0.02em). These should often be placed with asymmetrical margins to create a custom editorial look.
- **Headline & Title:** Use `on_surface` (#1e1924) for maximum contrast. These are your anchors.
- **Body (lg/md):** Our primary reading grade. Ensure a generous line-height (1.6x) to maintain the "clean and modern" promise.
- **Label (md/sm):** Reserved for metadata. Use `on_surface_variant` (#4c4355) to create a clear secondary tier of information.

---

## 4. Elevation & Depth

In this system, we do not "drop shadows"; we "emit light."

- **Tonal Layering:** Always attempt to solve hierarchy with a tone shift first. A `surface-container-lowest` card sitting on a `surface-container-low` background provides a crisp, "tucked-in" look.
- **Ambient Shadows:** If a shadow is required for a floating state, use the `on_surface` color at 6% opacity with a 32px blur and 8px Y-offset. This mimics a soft, natural overhead gallery light.
- **The Ghost Border Fallback:** For interactive inputs or accessibility-critical containers, use a "Ghost Border": `outline_variant` (#cec2d8) at **15% opacity**. Never use a 100% opaque border.

---

## 5. Components

### Buttons

- **Primary:** High-impact. Gradient of `primary` to `primary_container`. White text. Border-radius: `md` (0.375rem).
- **Secondary:** Ghost style. `surface-container-highest` background with `on_primary_fixed_variant` (#6200bc) text. No border.
- **Tertiary:** Text only. Bold `primary` (#6100ba) with a subtle `4px` bottom-padding underline that expands on hover.

### Cards & Lists

- **Forbid Dividers:** Do not use horizontal lines between list items. Use 16px–24px of vertical whitespace.
- **List Interaction:** On hover, a list item should transition its background to `surface-container-low`.

### Input Fields

- **Style:** Minimalist. `surface-container-lowest` background with a `Ghost Border` (15% opacity `outline_variant`).
- **Active State:** The border transitions to 100% opacity `primary_container` (#7f13ec) with a `2px` stroke.

### Editorial Signature Components

- **The "Pull-Quote" Module:** Large `headline-lg` text, center-aligned, with a `secondary_container` (#c297fd) vertical accent bar (4px width) to the left.
- **Staggered Image Mesh:** Images should not be uniform. Use a mix of `DEFAULT` (0.25rem) and `xl` (0.75rem) corner radii across an image set to create a curated, collage-like feel.

---

## 6. Do’s and Don’ts

### Do

- **Do** use asymmetrical white space. If an element is centered, try offsetting the headline to the left to create "The Digital Curator" look.
- **Do** use `primary_fixed_dim` (#d8b9ff) for subtle highlights in text or small icon backgrounds to keep the "purple" theme present without being overwhelming.
- **Do** prioritize `body-lg` for readability. In an editorial system, bigger is usually better.

### Don’t

- **Don’t** use pure black (#000) for text. Use `on_background` (#1e1924) to keep the contrast high but the "ink" sophisticated.
- **Don’t** use "Card-in-Card" patterns with shadows. Use tonal shifts (e.g., a `surface-container-highest` card inside a `surface-container-low` section).
- **Don’t** use standard 1px grey dividers. If you must separate, use a `surface-dim` (#e0d7e6) block of 8px height to create a "gutter" rather than a line.
