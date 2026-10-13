# Section 5: Cross-Browser and Responsive Testing

## Setups

| Setup | Details |
|---|---|
| Desktop: Google Chrome | _version (chrome://version)_ on _OS_, window _width × height_ |
| Desktop: Edge or Firefox | _browser + version_ on _OS_ |
| Mobile app | AbleSpace app _version_ from _Play Store / App Store_ on _device, OS version_ |
| Phone emulation | Chrome DevTools device mode (Ctrl+Shift+M), **390 × 844** |
| iPad Mini emulation | Chrome DevTools device mode, **768 × 1024** |
| iPad Pro emulation | Chrome DevTools device mode, **1024 × 1366** |

**What I checked on every setup:** broken layout, text overflow or truncation, overlapping elements, unwanted horizontal scrolling, buttons or inputs that are hard to tap (too small or too close together), content cut off, dialogs that don't fit the screen, and features that behave differently between setups.

**Legend:** ✅ Pass · ❌ Fail (see the note) · ⚠️ Works, with a minor issue

## Compatibility matrix

| Area / Feature | Chrome (desktop) | Edge / Firefox (desktop) | Mobile app | 390 × 844 | iPad Mini | iPad Pro | Notes |
|---|---|---|---|---|---|---|---|
| **Sign-up and login** | _TODO_ | _TODO_ | _TODO_ | ⚠️ | ✅ | ✅ | Public sign-in and sign-up pages were checked at 390, 768 and 1024 px. There is no horizontal scroll (page width equals viewport width), the social-login buttons stack full-width, and the inline validation messages are readable. ⚠️ At 390 px the "User not found!" toast covers the logo and header (BUG-03). _TODO: password step and actual login on each setup._ |
| **Main workflow** (add, search and edit a student) | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | |
| **Forms** (Add Student dialog, Add Goal pop-up, Take Data) | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | The Add Goal pop-up has three panels (data types, form, preview); check how they fit on 390 px and iPad Mini. |
| **Navigation** (tabs, menus, back/forward, logout) | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | |
| **Data and graphs** (Data tab, goal progress) | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | _TODO_ | Wide data tables are a common source of horizontal scroll on small screens. |

## Notes for every Fail

_For each ❌ above, add one line: the setup, what broke, and a link to the screenshot. For example: "390 × 844, Add Goal pop-up: the Save Goal button is below the fold and the dialog does not scroll, so the goal can't be saved. evidence/…png"_

-

## Evidence

- Sign-in at 390 × 844: [08-signin-mobile-390x844.png](evidence/public-pages/08-signin-mobile-390x844.png)
- Sign-in on iPad Mini, 768 × 1024: [09-signin-ipad-mini-768x1024.png](evidence/public-pages/09-signin-ipad-mini-768x1024.png)
- Sign-in on iPad Pro, 1024 × 1366: [10-signin-ipad-pro-1024x1366.png](evidence/public-pages/10-signin-ipad-pro-1024x1366.png)
- Sign-up at 390 × 844: [11-signup-mobile-390x844.png](evidence/public-pages/11-signup-mobile-390x844.png)
- Toast over the header at 390 px: [05-signin-user-not-found-toast-mobile-390.png](evidence/public-pages/05-signin-user-not-found-toast-mobile-390.png)
