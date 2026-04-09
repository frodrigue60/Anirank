# Design System Specification: Editorial Luminance

## 1. Overview & Creative North Star

**Creative North Star: The Lucid Curator**
This design system moves beyond the generic "SaaS White" aesthetic to embrace the spirit of a high-end digital broadsheet. It is characterized by expansive white space, razor-sharp typography, and a "Lucid" layering logic. We reject the standard "boxed-in" web experience in favor of an editorial layout where content breathes and hierarchy is established through tonal shifts rather than rigid lines.

By leveraging intentional asymmetry—such as offset headers and staggered grid placements—we break the "template" feel. The system utilizes a sophisticated light-theory approach where the signature neon purple acts as a rhythmic pulse across a pristine, high-contrast landscape.

---

## 2. Colors & Tonal Logic

The palette is rooted in a "Vibrant Purple" spectrum with high contrast. We use a base of `#faf9fd` (Surface) to provide a pristine, slightly cool off-white background that ensures maximum readability.

### Color Tokens (Opaque System)

- **Surface:** `#faf9fd` (Base background)
- **Surface Low:** `#f0ebf8` (Secondary backgrounds like sidebars)
- **Surface Container:** `#e5def2` (Cards and content grouping)
- **Surface Highest:** `#d8cdec` (Highest elevation/inputs)
- **On Surface:** `#09070e` (Primary text)
- **On Surface Variant:** `#3e3849` (Secondary text)
- **Primary:** `#683bc9` (Brand actions)
- **Primary Container:** `#7d4aeb` (Strong accents/Hover states)
- **Secondary Container:** `#a788e9` (Soft accents/Badges)
- **Outline Variant:** `#c4b9d8` (Subtle dividers/Borders)
- **Rating Star:** `#f59e0b` (Rating indicators)

### The "No-Line" Rule

**Explicit Instruction:** Designers are prohibited from using 1px solid borders to define major sections. Structural boundaries must be created through background color shifts.

- _Implementation:_ A hero section on `surface` transitions into a content feed on `surface-container-low`.
- _The Result:_ A seamless, high-end feel that mimics the flow of premium paper stocks.

### Surface Hierarchy & Nesting

Depth is achieved through the physical stacking of solid tones. We avoid transparency or background blurs.

- **Level 0 (Base):** `surface` (#faf9fd)
- **Level 1 (Sections):** `surface-low` (#f0ebf8)
- **Level 2 (Cards/Modules):** `surface-container` (#e5def2)
- **Level 3 (Popovers/Modals):** `surface-highest` (#d8cdec)

## 4. Elevation & Depth

In this system, we do not use drop shadows or glassmorphism. Hierarchy is created through solid color shifts.

- **Tonal Layering:** Hierarchy must be solved with a tone shift first. A `surface-container` card sitting on a `surface-low` background provides a clean, solid structure.
- **Solid Borders:** For interactive inputs or critical containers, use a solid `outline_variant` (#c4b9d8). Avoid using low-opacity borders or blurs.

---

## 5. Components

### Buttons

- **Primary:** High-impact. Solid `primary` background with White text. Border-radius: `md` (0.375rem).
- **Secondary:** Solid `surface-highest` background with `primary` (#683bc9) text. No border.
- **Tertiary:** Text only. Bold `primary` (#683bc9).

### Cards & Lists

- **Forbid Dividers:** Do not use horizontal lines between list items. Use solid background shifts or 16px–24px of vertical whitespace.
- **List Interaction:** On hover, a list item should transition its background to `surface-low`.

### Input Fields

- **Style:** Minimalist. Solid `surface-highest` background with a solid `outline_variant` (#c4b9d8) border.
- **Active State:** The border transitions to `primary` (#683bc9) with a `2px` stroke.

### Editorial Signature Components

- **The "Pull-Quote" Module:** Large `headline-lg` text with a `primary` (#683bc9) vertical accent bar (4px width) to the left.
- **Staggered Image Mesh:** Images should not be uniform. Use a mix of `DEFAULT` (0.25rem) and `xl` (0.75rem) corner radii across an image set to create a curated, collage-like feel.

---

## 6. Do’s and Don’ts

### Do

- **Do** use asymmetrical white space. If an element is centered, try offsetting the headline to the left to create "The Digital Curator" look.
- **Do** use `primary_fixed_dim` (#d8b9ff) for subtle highlights in text or small icon backgrounds to keep the "purple" theme present without being overwhelming.
- **Do** prioritize `body-lg` for readability. In an editorial system, bigger is usually better.

### Don’t

- **Don’t** use pure black (#000) for text. Use `on_surface` (#09070e) to keep the contrast high but sophisticated.
- **Don’t** use transparency or glassmorphism (blur, opacity backgrounds). Every layer must be a solid color from the surface hierarchy.
- **Don’t** use "Card-in-Card" patterns with shadows. Use tonal shifts (e.g., a `surface-container` card inside a `surface-low` section).
- **Don’t** use standard 1px grey dividers. If you must separate, use a solid `outline_variant` line or simple whitespace.
