# TuiTop

The Luxury Console
(construction in progress)

![Luxury Console](path/to/first.gif)

## Using a graphical desktop?

Sure you are, because it's easy:

- Icons multiply the value of the screen
  -- Emojis could bring this to a terminal
- Discoverability is easy
  -- though, we could make menus in a terminal too
- Windows with resize, drag-and-drop and more makes things easy
  -- But isn't mouse access available in terminals?

## Thinking in terms of X11

- Display Manager:  VT100 or Linux Console
  -- github.com/snadrus/tcellblit  -- render images fullscreen
- Compositor
  -- cview (branch of tview) has a great window renderer
- Window Manager
  -- TuiTop will provide this and a compositor
- Toolkit
  -- VT100 is the standard for consoles. Lets offer it.
  -- tcell-term
  -- A "new window" API is needed as well as more in the future.
- App Store
  -- Coming soon. A way to bring amazing TUI apps to everyone.

## Progress

- Verified good: CView (/verifications/compositor) and tcellblit (its example)
- In Progress: A working "final" layout, minus a usable new-window API

Next up:

- resize
- window close (on exit & on X)
- Re-imagine ctrl+c:
-- TuiTop menu should have at-least: exit, settings

Later:

- window auto-renaming
- in-house cview for box

Menu Items:

- Exit
- Settings?
- Calculator
- Selected Apps that have Brew+APT+WindowsCliThing install plans.
