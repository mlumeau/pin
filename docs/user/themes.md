# Themes

Themes control both the public profile and the settings UI.

## Built-in themes
Select a built-in theme from **Settings > Appearance**. The default is `classic`.

- `classic` - Warm paper, serif body, muted orange highlights. Default look.
- `noir` - Deep midnight panels with electric cyan accents for a dramatic dark mode.
- `mono` - Developer-focused, near-black canvas with neon green details and monospace type.
- `sunrise` - Bold modern gradient vibe with coral and amber accents.
- `forest` - Fresh botanical palette with crisp sans headings and soft shadows.
- `slate` - Cool graphite neutrals with lavender accents and transitional typography.
- `desert` - Sandy neutrals with turquoise punches and a humanist serif/sans mix.
- `ocean` - Deep navy gradients with seafoam accents and rounded sans headings.
- `pastel` - Playful candy palette with soft rounded cards and friendly fonts.
- `brutalist` - High-contrast black/white with bold yellow accents and rigid grids.
- `neonpop` - Black canvas, hyper-saturated magenta/cyan accents, bold geometric type.
- `velvet` - Deep burgundy and gold with elegant serif headings and soft glow.
- `glass` - Frosted glass blur vibe with icy blues and sleek sans typography.
- `midcentury` - Muted teals and mustards with rounded cards and retro sans serif.
- `tech` - Slate/teal gradient, neon lime highlights, condensed tech-forward font.

## Custom CSS
You can either:
- Upload a `.css` file, or
- Paste inline CSS.

Custom CSS loads after built-ins. To scope styles to a specific theme:
```
[data-theme="<name>"] {
  /* ... */
}
```

## Default theme
Admins can pick a server default theme and optionally lock it for users.

### Custom CSS storage
Uploaded theme CSS is stored under `PIN_UPLOADS_DIR/themes` and served as `/static/uploads/themes/<file>.css`.

### User theme policy
Admins can allow or disallow user-selected themes and custom CSS in **Settings > Appearance**.
