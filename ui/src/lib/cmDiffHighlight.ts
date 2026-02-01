import { diffLines } from "diff";
import { RangeSetBuilder } from "@codemirror/state";
import type { Extension } from "@codemirror/state";
import { Decoration, EditorView, ViewPlugin, type DecorationSet } from "@codemirror/view";

function splitLines(value: string): string[] {
  // Keep empty lines, but ignore the trailing empty element when the string ends with '\n'
  const parts = value.split("\n");
  if (parts.length > 0 && parts[parts.length - 1] === "") parts.pop();
  return parts;
}

export function computeChangedLineNumbers(left: string, right: string): {
  left: number[];
  right: number[];
} {
  const leftChanged = new Set<number>();
  const rightChanged = new Set<number>();

  let leftLine = 1;
  let rightLine = 1;

  const parts = diffLines(left, right);
  for (const p of parts) {
    const lines = splitLines(p.value);
    const count = lines.length;
    if (count === 0) continue;

    if ((p as { added?: boolean }).added) {
      for (let i = 0; i < count; i++) rightChanged.add(rightLine + i);
      rightLine += count;
      continue;
    }
    if ((p as { removed?: boolean }).removed) {
      for (let i = 0; i < count; i++) leftChanged.add(leftLine + i);
      leftLine += count;
      continue;
    }

    leftLine += count;
    rightLine += count;
  }

  return {
    left: Array.from(leftChanged).sort((a, b) => a - b),
    right: Array.from(rightChanged).sort((a, b) => a - b),
  };
}

export function lineHighlightExtension(
  lineNumbers: number[],
  className: string,
): Extension {
  const lines = new Set(lineNumbers);

  const build = (view: EditorView): DecorationSet => {
    const builder = new RangeSetBuilder<Decoration>();
    const total = view.state.doc.lines;
    for (let i = 1; i <= total; i++) {
      if (!lines.has(i)) continue;
      const line = view.state.doc.line(i);
      builder.add(line.from, line.from, Decoration.line({ class: className }));
    }
    return builder.finish();
  };

  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;
      constructor(view: EditorView) {
        this.decorations = build(view);
      }
      update(update: { view: EditorView; docChanged: boolean; viewportChanged: boolean }) {
        if (update.docChanged || update.viewportChanged) {
          this.decorations = build(update.view);
        }
      }
    },
    {
      decorations: (v) => v.decorations,
    },
  );
}

