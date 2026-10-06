---
version: alpha
name: Modal-design-analysis
description: "A dark infrastructure console rendered as a marketing site. Near-black ground (#181818) carries hairline-outlined panels one or two steps lighter, a single high-voltage lime (#7fee64) marks every primitive action and healthy state, and DM Mono does the technical talking — uppercase tracked section labels, metric values, identifiers, logs. Display type is a tight neutral grotesque (ABC Diatype) at weight 400–500 with negative tracking. Radii are deliberately small (3–10px) so panels read as instrument housing, while buttons go fully pill (48px). Ornament is close to zero: the energy comes from the contrast of lime against graphite and from motion that behaves like telemetry arriving."

source: https://modal.com
extracted: 2026-10-06
notes: "Token values below were read from Modal's own stylesheets (_app/immutable/assets/0.DxChTQCV.css), not eyeballed from screenshots."

colors:
  # Ground and panels. The whole product sits on graphite, not black.
  ground: "#181818"
  surface-1: "#1c1c1c"
  surface-2: "#242424"
  surface-3: "#272727"
  hairline: "#2f2f2f"
  hairline-strong: "#3b3b3b"
  hairline-loud: "#464646"

  # Ink. Pure white for display, then four steps down.
  ink: "#ffffff"
  ink-90: "#e8e8e8"
  ink-80: "#d1d1d1"
  ink-70: "#bababa"
  ink-60: "#a3a3a3"
  ink-50: "#8b8b8b"
  ink-40: "#747474"
  ink-30: "#5d5d5d"

  # The accent. One colour, used for primitive actions and healthy states.
  lime: "#7fee64"
  lime-90: "#75d95c"
  lime-80: "#6ac355"
  lime-60: "#569846"
  lime-40: "#416e36"
  lime-20: "#2d4327"
  lime-10: "#222d20"
  lime-05: "#1d231c"
  pale-green: "#c8f9b6"

  # State marks. Each is the 100 step of its own ramp.
  red: "#f87171"
  red-10: "#2e2121"
  yellow: "#ffea71"
  yellow-10: "#2f2d21"
  blue: "#91c8ef"
  pink: "#ff8de6"

  # Translucent washes. 5% of an accent over the ground, nothing stronger.
  wash-lime: "#7fee640d"
  wash-red: "#f871710d"
  scrim: "#1c1c1ccc"

typography:
  sans:
    family: ABC Diatype
    substitute: Inter
    features: "cv11"
    weights: [300, 400, 500, 600, 700]
  mono:
    family: DM Mono
    substitute: DM Mono   # freely available, so used as-is
    features: "ss02"

  display-xl:
    fontSize: 72px
    fontWeight: 400
    lineHeight: 1.0
    letterSpacing: -0.03em
  display-lg:
    fontSize: 60px
    fontWeight: 400
    lineHeight: 1.0
    letterSpacing: -0.028em
  display-md:
    fontSize: 48px
    fontWeight: 400
    lineHeight: 1.0
    letterSpacing: -0.025em
  title:
    fontSize: 30px
    fontWeight: 400
    lineHeight: 1.25
    letterSpacing: -0.02em
  card-title:
    fontSize: 20px
    fontWeight: 500
    lineHeight: 1.3
    letterSpacing: -0.012em
  body-lg:
    fontSize: 18px
    lineHeight: 1.55
  body:
    fontSize: 16px
    lineHeight: 1.5
  body-sm:
    fontSize: 14px
    lineHeight: 1.5
  caption:
    fontSize: 12px
    lineHeight: 1.4

  # The signature. Section labels are mono, uppercase, widely tracked, and
  # usually prefixed with a zero-padded index.
  eyebrow:
    family: mono
    fontSize: 12px
    fontWeight: 400
    textTransform: uppercase
    letterSpacing: 0.14em

radii:
  xs: 3px
  sm: 4px
  md: 7px
  lg: 10px
  pill: 48px   # buttons and chips only

spacing:
  gutter: 20px
  margin: 24px
  margin-mobile: 32px
  margin-tablet: 64px
  section: 144px

motion:
  # Modal's marketing curve is an out-quart; hovers run on the browser's
  # plain `ease` at 150ms.
  ease-quart: cubic-bezier(0.165, 0.84, 0.44, 1)
  ease-hover: ease
  duration-hover: 150ms

rules:
  - The ground is #181818, never pure black. Panels step up, never down.
  - Borders do the work shadows do elsewhere. One hairline, 1px, #2f2f2f.
    There are no drop shadows in this system.
  - Lime is for primitive actions and healthy states. A page gets one filled
    lime button. Everything else is an outlined button on graphite.
  - Mono is not decoration. If a string is a measurement, an identifier, a
    status, a timestamp, a count or a command, it is mono. If it is prose, it
    is sans.
  - Section labels are uppercase mono with a zero-padded index (01, 02, 03).
  - Radii stay small — 7px is the default panel. Only buttons and chips go
    pill. A 16px+ radius reads as a different product.
  - Numbers are tabular. A column of metrics must not reflow as it updates.
  - Motion behaves like telemetry: things draw in, count up, and stream.
    Nothing bounces, nothing scales up from nothing.
