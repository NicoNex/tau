# Editors

The syntax files for the editors that have them.

- **Zed**: the extension is a repository of its own,
  [tau-zed](https://github.com/NicoNex/tau-zed), and the grammar it builds is
  [tree-sitter-tau](https://github.com/NicoNex/tree-sitter-tau). See
  [UPDATING.md](https://github.com/NicoNex/tau-zed/blob/main/UPDATING.md) for
  how to change either of them.
- **Sublime Text**: `sublime-text/`, a syntax and a comment definition. Copy
  them into the `Packages/User` directory of Sublime.

There used to be a copy of the tree-sitter grammar here as well. It was a copy,
it drifted, and the one Zed actually builds is the one in tree-sitter-tau, so
it is gone: there is one grammar now, in one place.
