# TuiTop

The Luxury Console
(construction in progress)

![Luxury Console](/first.gif)

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

- In Progress: A working "final" layout, minus a usable new-window API

Next up:

- window close (on exit & on X)
-- LOL! there is NO CLOSE for windows!
- Menu: exit (rm ctrl-C), settings, apps
- Add env TUITOP=/tmp/foo1234 (symlink to binary, allow ran like this to add window)
- Selection + Copy/Paste

Later:

- window auto-renaming
  github.com/shirou/gopsutil/v3 process.NewProcess(pid) [ .Children() .Name() ]
- in-house cview for box
- persist thru SIGHUP

Easy Starter Tasks:

- Show/Hide cursor
- Scrollback
- text selection within a window

Config folder: ~/.config/tuitop/

- bin/upt   (TRUE TODAY)
- bin/ (src-build-binaries-linked-here)
- src/ git'd sources (FUTURE)
- menu/ items mods (FUTURE)
- config/ config files (FUTURE)
- logs/ logs (FUTURE)
- cache/ cache for which menu items have been installed and which cannot be (FUTURE)
