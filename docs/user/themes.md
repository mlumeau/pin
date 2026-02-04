# Themes

Themes control both the public profile and the settings UI.

## Built-in themes
Select a built-in theme from **Settings > Appearance**. The default is `classic`.

- `classic` - Warm paper-inspired layout with editorial serif typography and clay accents.
- `noir` - Cinematic dark interface with ocean-tinted cyan highlights and high legibility contrast.
- `mono` - Monospace terminal style with deep black surfaces and vivid green focus states.
- `forest` - Calm botanical greens with practical sans-serif typography and gentle depth.
- `slate` - Pragmatic blue-gray theme with subtle glass-like surfaces for long-form reading.
- `pastel` - Soft candy palette with sunrise-inspired coral actions and playful headings.
- `highcontrast` - Accessibility-first high-contrast black/white style with strong focus states.
- `neonpop` - Expressive night palette with electric pink/cyan accents and punchy typography.
- `velvet` - Luxurious dark plum surfaces with gold details and refined serif titles.
- `tech` - Focused dark productivity theme with mint accents and an ocean-depth backdrop.

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
