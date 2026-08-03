# PicoGoGUI implementation phases

| Phase | Scope | Status |
|---|---|---|
| 1 | Window, Application, event loop, WebView2 host, RPC/message bridge | Implemented; lifecycle, queued startup messages, asynchronous error channel, multi-window cleanup |
| 2 | Layouts, Button, Label, TextBox, NumberBox, CheckBox, ComboBox, binding | Implemented; origin-aware two-way binding, Canvas, Grid, Split, Tabs, Dock |
| 3 | Tables, TreeView, Forms, Dialogs, Notifications | Implemented; table sort/filter, keyboard tree, form validation, prompt and native file/folder dialogs |
| 4 | Themes, Animations, System Tray, Clipboard, Drag & Drop | Implemented; custom/system-following themes, reduced-motion animations, nested tray menus, file DropZone |
| 5 | Visual Designer | Implemented MVP+; validation, undo/redo, native save/load and geometry-preserving Go export |
| 6 | Compile-time Plugin System | Implemented; controls/assets/designer hooks, API compatibility, dependencies and lifecycle |

The plugin system intentionally uses compile-time Go extensions so deployed applications remain a single executable. It does not load native Go DLL plugins at runtime.
